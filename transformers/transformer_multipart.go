package transformers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strconv"

	"github.com/go-courier/httptransport/httpx"
	verrors "github.com/go-courier/httptransport/validator"
	typesutil "github.com/go-courier/x/types"
	"github.com/pkg/errors"
)

func init() {
	TransformerMgrDefault.Register(&TransformerMultipart{})
}

/*
TransformerMultipart for multipart/form-data
*/
type TransformerMultipart struct {
	*FlattenParams
}

func (TransformerMultipart) Names() []string {
	return []string{"multipart/form-data", "multipart", "form-data"}
}

func (TransformerMultipart) NamedByTag() string {
	return "name"
}

func (transformer *TransformerMultipart) String() string {
	return transformer.Names()[0]
}

func (TransformerMultipart) New(ctx context.Context, typ typesutil.Type) (Transformer, error) {
	transformer := &TransformerMultipart{}

	typ = typesutil.Deref(typ)
	if typ.Kind() != reflect.Struct {
		return nil, errors.Errorf("content transformer `%s` should be used for struct type", transformer)
	}

	transformer.FlattenParams = &FlattenParams{}

	if err := transformer.FlattenParams.CollectParams(ctx, typ); err != nil {
		return nil, err
	}

	return transformer, nil
}

func (transformer *TransformerMultipart) EncodeTo(ctx context.Context, w io.Writer, v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	multipartWriter := multipart.NewWriter(w)

	httpx.MaybeWriteHeader(ctx, w, multipartWriter.FormDataContentType(), nil)

	errSet := verrors.NewErrorSet()

	for i := range transformer.Parameters {
		p := transformer.Parameters[i]

		fieldValue := p.FieldValue(rv)

		if p.Transformer != nil {
			st := NewTransformerSuper(p.Transformer, &p.TransformerOption.CommonTransformOption)

			partWriter := NewFormPartWriter(func(header textproto.MIMEHeader) (io.Writer, error) {
				paramFilename := ""
				if v := header.Get("Content-Disposition"); v != "" {
					_, disposition, err := mime.ParseMediaType(v)
					if err == nil {
						if f, ok := disposition["filename"]; ok {
							paramFilename = fmt.Sprintf("; filename=%s", strconv.Quote(f))
						}
					}
				}
				// always overwrite name
				header.Set("Content-Disposition", fmt.Sprintf("form-data; name=%s%s", strconv.Quote(p.Name), paramFilename))
				return multipartWriter.CreatePart(header)
			})

			if err := st.EncodeTo(ctx, partWriter, fieldValue); err != nil {
				errSet.AddErr(err, p.Name)
				continue
			}
		}
	}

	return multipartWriter.Close()
}

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

func (transformer *TransformerMultipart) DecodeFrom(ctx context.Context, r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	header := MIMEHeader(headers...)
	_, params, err := mime.ParseMediaType(header.Get(httpx.HeaderContentType))
	if err != nil {
		return err
	}

	reader := multipart.NewReader(r, params["boundary"])
	form, err := reader.ReadForm(defaultMaxMemory)
	if err != nil {
		return err
	}

	errSet := verrors.NewErrorSet()

	for i := range transformer.Parameters {
		p := transformer.Parameters[i]

		if p.Transformer != nil {
			st := NewTransformerSuper(p.Transformer, &p.TransformerOption.CommonTransformOption)

			if files, ok := form.File[p.Name]; ok {
				readers := NewFileHeaderReaders(files)
				if err := st.DecodeFrom(ctx, readers, p.FieldValue(rv).Addr()); err != nil {
					errSet.AddErr(err, p.Name)
				}
				continue
			}

			if fieldValues, ok := form.Value[p.Name]; ok {
				readers := NewStringReaders(fieldValues)

				if err := st.DecodeFrom(ctx, readers, p.FieldValue(rv).Addr()); err != nil {
					errSet.AddErr(err, p.Name)
				}
			}
		}
	}

	return nil
}

func MustNewFileHeader(fieldName string, filename string, r io.Reader) *multipart.FileHeader {
	fileHeader, err := NewFileHeader(fieldName, filename, r)
	if err != nil {
		panic(err)
	}
	return fileHeader
}

func NewFileHeader(fieldName string, filename string, r io.Reader) (*multipart.FileHeader, error) {
	buffer := bytes.NewBuffer(nil)
	multipartWriter := multipart.NewWriter(buffer)

	filePart, err := multipartWriter.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(filePart, r); err != nil {
		return nil, err
	}
	multipartWriter.Close()

	reader := multipart.NewReader(buffer, multipartWriter.Boundary())
	form, err := reader.ReadForm(int64(buffer.Len()))
	if err != nil {
		return nil, err
	}

	return form.File[fieldName][0], nil
}

func NewFormPartWriter(createPartWriter func(header textproto.MIMEHeader) (io.Writer, error)) *FormPartWriter {
	return &FormPartWriter{
		createPartWriter: createPartWriter,
		header:           http.Header{},
	}
}

type FormPartWriter struct {
	createPartWriter func(header textproto.MIMEHeader) (io.Writer, error)
	partWriter       io.Writer
	header           http.Header
}

func (w *FormPartWriter) NextWriter() io.Writer {
	return NewFormPartWriter(w.createPartWriter)
}

func (w *FormPartWriter) Header() http.Header {
	return w.header
}

func (w *FormPartWriter) Write(p []byte) (n int, err error) {
	if w.partWriter == nil {
		w.partWriter, err = w.createPartWriter(textproto.MIMEHeader(w.header))
		if err != nil {
			return -1, err
		}
	}
	return w.partWriter.Write(p)
}

func NewFileHeaderReaders(fileHeaders []*multipart.FileHeader) *StringReaders {
	bs := make([]io.Reader, len(fileHeaders))
	for i := range fileHeaders {
		bs[i] = &FileHeaderReader{v: fileHeaders[i]}
	}

	return &StringReaders{
		readers: bs,
	}
}

type FileHeaderReader struct {
	v      *multipart.FileHeader
	opened multipart.File
}

func (f *FileHeaderReader) Interface() interface{} {
	return f.v
}

func (f *FileHeaderReader) Read(p []byte) (int, error) {
	if f.opened == nil {
		file, err := f.v.Open()
		if err != nil {
			return -1, err
		}
		f.opened = file
	}
	n, err := f.opened.Read(p)
	if err == io.EOF {
		return n, f.opened.Close()
	}
	return n, err
}
