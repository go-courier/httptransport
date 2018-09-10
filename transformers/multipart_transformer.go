package transformers

import (
	"bytes"
	"fmt"
	"go/ast"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"reflect"
	"strconv"

	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/reflectx"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/validator"
	"github.com/go-courier/validator/errors"
)

func init() {
	TransformerMgrDefault.Register(&MultipartTransformer{})
}

type MultipartTransformer struct {
	fieldTransformers map[string]Transformer
	fieldOpts         map[string]TransformerOption
	validators        map[string]validator.Validator
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

func (transformer *MultipartTransformer) NewValidator(typ typesutil.Type, mgr validator.ValidatorMgr) (validator.Validator, error) {
	transformer.validators = map[string]validator.Validator{}

	typ = typesutil.Deref(typ)

	errSet := errors.NewErrorSet("")

	typesutil.EachField(typ, "name", func(field typesutil.StructField, fieldDisplayName string, omitempty bool) bool {
		fieldName := field.Name()
		fieldValidator, err := NewValidator(field, field.Tag().Get("validate"), omitempty, transformer.fieldTransformers[fieldName], mgr)
		if err != nil {
			errSet.AddErr(err, fieldName)
			return true
		}
		transformer.validators[fieldName] = fieldValidator
		return true
	})

	return transformer, errSet.Err()
}

func (transformer *MultipartTransformer) Validate(v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	errSet := errors.NewErrorSet("")
	transformer.validate(rv, errSet)
	return errSet.Err()
}

func (transformer *MultipartTransformer) validate(rv reflect.Value, errSet *errors.ErrorSet) {
	typ := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := rv.Field(i)
		fieldName, _, exists := typesutil.FieldDisplayName(field.Tag, "name", field.Name)

		if !ast.IsExported(field.Name) || fieldName == "-" {
			continue
		}

		fieldType := reflectx.Deref(field.Type)
		isStructType := fieldType.Kind() == reflect.Struct

		if field.Anonymous && isStructType && !exists {
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				fieldValue = reflectx.New(fieldType)
			}
			transformer.validate(fieldValue, errSet)
			continue
		}

		if fieldValidator, ok := transformer.validators[field.Name]; ok {
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				fieldValue = reflectx.New(field.Type)
			}
			err := fieldValidator.Validate(fieldValue)
			errSet.AddErr(err, fieldName)
		}
	}
}

func (transformer *MultipartTransformer) String() string {
	return transformer.Names()[0]
}

func (MultipartTransformer) New(typ typesutil.Type, mgr TransformerMgr) (Transformer, error) {
	transformer := &MultipartTransformer{}
	transformer.fieldTransformers = map[string]Transformer{}
	transformer.fieldOpts = map[string]TransformerOption{}

	typ = typesutil.Deref(typ)
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("content transformer `%s` should be used for struct type", transformer)
	}

	errSet := errors.NewErrorSet("")

	typesutil.EachField(typ, "name", func(field typesutil.StructField, fieldDisplayName string, omitempty bool) bool {
		opt := TransformerOptionFromStructField(field)
		targetType := field.Type()
		fieldName := field.Name()

		if !IsBytes(targetType) {
			switch targetType.Kind() {
			case reflect.Array, reflect.Slice:
				opt.Explode = true
				targetType = targetType.Elem()
			}
		}

		fieldTransformer, err := mgr.NewTransformer(targetType, opt)
		if err != nil {
			errSet.AddErr(err, fieldName)
			return true
		}

		transformer.fieldTransformers[fieldName] = fieldTransformer
		transformer.fieldOpts[fieldName] = opt
		return true
	})

	return transformer, errSet.Err()
}

func NewValidator(field typesutil.StructField, validateStr string, omitempty bool, transformer Transformer, mgr validator.ValidatorMgr) (validator.Validator, error) {
	if validateStr == "" && typesutil.Deref(field.Type()).Kind() == reflect.Struct {
		validateStr = "@struct" + "<" + transformer.NamedByTag() + ">"
	}

	if t, ok := transformer.(interface {
		NewValidator(typ typesutil.Type, mgr validator.ValidatorMgr) (validator.Validator, error)
	}); ok {
		return t.NewValidator(field.Type(), mgr)
	}
	return mgr.Compile([]byte(validateStr), field.Type(), func(rule *validator.Rule) {
		if omitempty {
			rule.Optional = true
		}
		if defaultValue, ok := field.Tag().Lookup("default"); ok {
			rule.DefaultValue = []byte(defaultValue)
		}
	})
}

func (transformer *MultipartTransformer) EncodeToWriter(w io.Writer, v interface{}) (string, error) {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	multipartWriter := multipart.NewWriter(w)
	errSet := errors.NewErrorSet("")

	addPart := func(rv reflect.Value, fieldName string, fieldTransformer Transformer) error {
		buf := bytes.NewBuffer(nil)
		contentType, err := fieldTransformer.EncodeToWriter(buf, rv)
		if err != nil {
			return err
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
					appendFile(fieldOpt.FieldName, v[i])
				}
				return
			case *multipart.FileHeader:
				appendFile(fieldOpt.FieldName, v)
				return
			}
		}

		if fieldOpt.Explode {
			for i := 0; i < fieldValue.Len(); i++ {
				if err := addPart(fieldValue.Index(i), fieldOpt.FieldName, fieldTransformer); err != nil {
					errSet.AddErr(err, fieldOpt.FieldName, i)
				}
			}
		} else {
			if err := addPart(fieldValue, fieldOpt.FieldName, fieldTransformer); err != nil {
				errSet.AddErr(err, fieldOpt.FieldName)
			}
		}
	})

	if err := errSet.Err(); err != nil {
		return "", err
	}

	if err := multipartWriter.Close(); err != nil {
		return "", err
	}

	return multipartWriter.FormDataContentType(), nil
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

	errSet := errors.NewErrorSet("")

	setValue := func(rv reflect.Value, fieldTransformer Transformer, fieldName string, idx int, omitempty bool) error {
		if rv.Type().ConvertibleTo(typeFileHeader) {
			if len(form.File[fieldName]) > idx {
				rv.Set(reflect.ValueOf(form.File[fieldName][idx]))
			}
			return nil
		}
		if omitempty && reflectx.IsEmptyValue(rv) {
			return nil
		}
		if len(form.Value[fieldName]) > idx {
			fieldTransformer.DecodeFromReader(bytes.NewBufferString(form.Value[fieldName][idx]), rv)
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
