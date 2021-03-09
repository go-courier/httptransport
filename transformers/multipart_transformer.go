package transformers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"reflect"
	"strconv"

	"github.com/pkg/errors"

	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/reflectx/typesutil"
	verrors "github.com/go-courier/validator/errors"
)

func init() {
	TransformerMgrDefault.Register(&MultipartTransformer{})
}

type MultipartTransformer struct {
	*FlattenParams
}

/*
transformer for multipart/form-data
*/
func (MultipartTransformer) Names() []string {
	return []string{"multipart/form-data", "multipart", "form-data"}
}

func (MultipartTransformer) NamedByTag() string {
	return "name"
}

func (transformer *MultipartTransformer) String() string {
	return transformer.Names()[0]
}

func (MultipartTransformer) New(ctx context.Context, typ typesutil.Type) (Transformer, error) {
	transformer := &MultipartTransformer{}

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

func (transformer *MultipartTransformer) EncodeToWriter(w io.Writer, v interface{}) (string, error) {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	multipartWriter := multipart.NewWriter(w)

	return superWrite(w, func(w io.Writer) error {
		errSet := verrors.NewErrorSet("")

		addPart := func(rv reflect.Value, fieldName string, fieldTransformer Transformer, omitempty bool) error {
			buf := bytes.NewBuffer(nil)
			contentType, err := fieldTransformer.EncodeToWriter(buf, rv)
			if err != nil {
				return err
			}

			if buf.Len() == 0 && omitempty {
				return nil
			}

			h := make(textproto.MIMEHeader)
			h.Set(httpx.HeaderContentType, contentType)
			h.Set(httpx.HeaderContentDisposition, fmt.Sprintf(`form-data; name=%s`, strconv.Quote(fieldName)))

			part, err := multipartWriter.CreatePart(h)
			if err != nil {
				return err
			}
			if _, err := part.Write(buf.Bytes()); err != nil {
				return err
			}
			return nil
		}

		appendFile := func(fieldName string, fileHeader *multipart.FileHeader) error {
			if fileHeader == nil {
				return nil
			}

			filePart, err := multipartWriter.CreateFormFile(fieldName, fileHeader.Filename)
			if err != nil {
				return err
			}

			file, err := fileHeader.Open()
			if err != nil {
				return err
			}

			if _, err := io.Copy(filePart, file); err != nil {
				return err
			}

			return nil
		}

		NamedStructFieldValueRange(reflect.Indirect(rv), func(fieldValue reflect.Value, field *reflect.StructField) {
			fieldOpt := transformer.fieldOpts[field.Name]
			fieldTransformer := transformer.fieldTransformers[field.Name]

			if fieldValue.CanInterface() {
				switch v := fieldValue.Interface().(type) {
				case []*multipart.FileHeader:
					for i := range v {
						_ = appendFile(fieldOpt.FieldName, v[i])
					}
					return
				case *multipart.FileHeader:
					_ = appendFile(fieldOpt.FieldName, v)
					return
				}
			}

			if fieldOpt.Explode {
				for i := 0; i < fieldValue.Len(); i++ {
					if err := addPart(fieldValue.Index(i), fieldOpt.FieldName, fieldTransformer, fieldOpt.Omitempty); err != nil {
						errSet.AddErr(err, fieldOpt.FieldName, i)
					}
				}
			} else {
				if err := addPart(fieldValue, fieldOpt.FieldName, fieldTransformer, fieldOpt.Omitempty); err != nil {
					errSet.AddErr(err, fieldOpt.FieldName)
				}
			}
		})

		if err := errSet.Err(); err != nil {
			return err
		}

		return multipartWriter.Close()
	}, multipartWriter.FormDataContentType())
}

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

var typeFileHeader = reflect.TypeOf(&multipart.FileHeader{})

func (transformer *MultipartTransformer) DecodeFromReader(r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
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

	errSet := verrors.NewErrorSet("")

	setValue := func(rv reflect.Value, fieldTransformer Transformer, fieldName string, idx int, omitempty bool) error {
		if rv.Type().ConvertibleTo(typeFileHeader) {
			if len(form.File[fieldName]) > idx {
				rv.Set(reflect.ValueOf(form.File[fieldName][idx]))
			}
			return nil
		}

		if len(form.Value[fieldName]) > idx {
			if err := fieldTransformer.DecodeFromReader(bytes.NewBufferString(form.Value[fieldName][idx]), rv); err != nil {
				return err
			}
		}
		return nil
	}

	NamedStructFieldValueRange(reflect.Indirect(rv), func(fieldValue reflect.Value, field *reflect.StructField) {
		fieldOpt := transformer.fieldOpts[field.Name]

		fieldTransformer := transformer.fieldTransformers[field.Name]

		if fieldOpt.Explode {
			lenOfValues := 0
			if field.Type.Elem().ConvertibleTo(typeFileHeader) {
				lenOfValues = len(form.File[fieldOpt.FieldName])
			} else {
				lenOfValues = len(form.Value[fieldOpt.FieldName])
			}

			if fieldOpt.Omitempty && lenOfValues == 0 {
				return
			}

			if field.Type.Kind() == reflect.Slice {
				fieldValue.Set(reflect.MakeSlice(field.Type, lenOfValues, lenOfValues))
			}

			for idx := 0; idx < fieldValue.Len(); idx++ {
				if err := setValue(fieldValue.Index(idx), fieldTransformer, fieldOpt.FieldName, idx, fieldOpt.Omitempty); err != nil {
					errSet.AddErr(err, fieldOpt.FieldName, idx)
				}
			}
		} else {
			if err := setValue(fieldValue, fieldTransformer, fieldOpt.FieldName, 0, fieldOpt.Omitempty); err != nil {
				errSet.AddErr(err, fieldOpt.FieldName)
				return
			}
		}
	})

	return errSet.Err()
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
