package httptransport

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/reflectx"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/statuserror"
	"github.com/go-courier/validator"
	"github.com/go-courier/validator/errors"
	"github.com/julienschmidt/httprouter"
)

func NewRequestTransformerMgr(
	transformerMgr transformers.TransformerMgr,
	validatorMgr validator.ValidatorMgr,
) *RequestTransformerMgr {
	r := &RequestTransformerMgr{
		ValidatorMgr:   validatorMgr,
		TransformerMgr: transformerMgr,
	}
	r.SetDefaults()
	return r
}

func (mgr *RequestTransformerMgr) SetDefaults() {
	if mgr.ValidatorMgr == nil {
		mgr.ValidatorMgr = validator.ValidatorMgrDefault
	}
	if mgr.TransformerMgr == nil {
		mgr.TransformerMgr = transformers.TransformerMgrDefault
	}
}

func (mgr *RequestTransformerMgr) NewRequest(method string, rawUrl string, v interface{}) (*http.Request, error) {
	if v == nil {
		return http.NewRequest(method, rawUrl, nil)
	}
	rt, err := mgr.NewRequestTransformer(context.Background(), reflect.TypeOf(v))
	if err != nil {
		return nil, err
	}
	return rt.NewRequest(method, rawUrl, v)
}

func (mgr *RequestTransformerMgr) NewRequestTransformer(ctx context.Context, typ reflect.Type) (*RequestTransformer, error) {
	key := reflectx.FullTypeName(typ)
	if v, ok := mgr.cache.Load(key); ok {
		return v.(*RequestTransformer), nil
	}
	rt, err := mgr.newRequestTransformer(ctx, typ)
	if err != nil {
		return nil, err
	}
	mgr.cache.Store(key, rt)
	return rt, nil
}

func (mgr *RequestTransformerMgr) newRequestTransformer(ctx context.Context, typ reflect.Type) (*RequestTransformer, error) {
	errSet := errors.NewErrorSet("")

	rt := &RequestTransformer{}
	rt.Type = reflectx.Deref(typ)
	rt.Parameters = map[string]*RequestParameter{}

	typesutil.EachField(typesutil.FromRType(rt.Type), "name", func(field typesutil.StructField, fieldDisplayName string, omitempty bool) bool {
		tag := field.Tag()
		fieldName := field.Name()

		in, exists := tag.Lookup("in")
		if !exists {
			panic(fmt.Errorf("missing tag `in` of %s", field.Name()))
		}

		if in == "path" {
			omitempty = false
		}

		parameter := NewRequestParameter(fieldDisplayName, in)
		parameter.Omitempty = omitempty

		transformOpt := transformers.TransformerOptionFromStructField(field)

		getTransformer := func() (transformers.Transformer, error) {
			targetType := field.Type()
			if !(in == "body" || in == "path") {
				if !transformers.IsBytes(targetType) {
					switch targetType.Kind() {
					case reflect.Array, reflect.Slice:
						parameter.Explode = true
						targetType = targetType.Elem()
					}
				}
			}
			return mgr.NewTransformer(ctx, targetType, transformOpt)
		}

		transformer, err := getTransformer()
		if err != nil {
			errSet.AddErr(err, field.Name())
			return true
		}
		parameter.Transformer = transformer

		parameterValidator, err := transformers.NewValidator(validator.ContextWithValidatorMgr(context.Background(), mgr.ValidatorMgr), field, tag.Get(validator.TagValidate), omitempty, transformer)
		if err != nil {
			errSet.AddErr(err, field.Name())
			return true
		}

		parameter.Validator = parameterValidator
		rt.Parameters[fieldName] = parameter

		return true
	}, "in")

	return rt, errSet.Err()
}

type RequestTransformerMgr struct {
	validator.ValidatorMgr
	transformers.TransformerMgr
	cache sync.Map
}

type RequestTransformer struct {
	Type       reflect.Type
	Parameters map[string]*RequestParameter
}

func (t *RequestTransformer) NewRequest(method string, rawUrl string, v interface{}) (*http.Request, error) {
	if v == nil {
		return http.NewRequest(method, rawUrl, nil)
	}

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	typ := reflectx.Deref(rv.Type())
	if !typ.ConvertibleTo(t.Type) {
		return nil, fmt.Errorf("unmatched request transformer, need %s but got %s", t.Type, typ)
	}

	errSet := errors.NewErrorSet("")
	params := httprouter.Params{}
	query := url.Values{}
	header := http.Header{}
	cookies := make([]*http.Cookie, 0)
	body := bytes.NewBuffer(nil)

	addParam := func(param *RequestParameter, value string) {
		if param.Omitempty && !param.Explode && value == "" {
			return
		}
		switch param.In {
		case "path":
			params = append(params, httprouter.Param{
				Key:   param.Name,
				Value: value,
			})
		case "query":
			query.Add(param.Name, value)
		case "header":
			header.Add(param.Name, value)
		case "cookie":
			// just set value
			cookies = append(cookies, &http.Cookie{
				Name:  param.Name,
				Value: value,
			})
		}
	}

	transformers.NamedStructFieldValueRange(reflect.Indirect(rv), func(fieldValue reflect.Value, field *reflect.StructField) {
		param := t.Parameters[field.Name]
		if param == nil {
			return
		}

		if param.In == "body" {
			contentType, err := param.Transformer.EncodeToWriter(body, fieldValue)
			if err != nil {
				errSet.AddErr(err, param.Name)
				return
			}
			header.Set(httpx.HeaderContentType, contentType)
			return
		}

		if param.Explode {
			if fieldValue.IsValid() {
				for i := 0; i < fieldValue.Len(); i++ {
					buf := bytes.NewBuffer(nil)
					if _, err := param.Transformer.EncodeToWriter(buf, fieldValue.Index(i)); err != nil {
						errSet.AddErr(err, param.Name, i)
						return
					}
					addParam(param, buf.String())
				}
			}
		} else {
			buf := bytes.NewBuffer(nil)
			if _, err := param.Transformer.EncodeToWriter(buf, fieldValue); err != nil {
				errSet.AddErr(err, param.Name)
				return
			}
			addParam(param, buf.String())
		}
	}, "in")

	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	u.Path = NewPathnamePattern(u.Path).Stringify(params)
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header = header

	for i := range cookies {
		req.AddCookie(cookies[i])
	}

	if len(params) > 0 {
		return req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params)), nil
	}

	return req, nil
}

type BadRequest struct {
	errTalk     bool
	msg         string
	errorFields []*statuserror.ErrorField
}

func (e *BadRequest) EnableErrTalk() {
	e.errTalk = true
}

func (e *BadRequest) SetMsg(msg string) {
	e.msg = msg
}

func (e *BadRequest) AddErr(err error, in string, nameOrIdx ...interface{}) {
	errSet := errors.NewErrorSet("")

	if es, ok := err.(*errors.ErrorSet); ok && in == "body" {
		errSet = es
	} else {
		errSet.AddErr(err, nameOrIdx...)
	}

	errSet.Flatten().Each(func(fieldErr *errors.FieldError) {
		e.errorFields = append(e.errorFields, statuserror.NewErrorField(in, fieldErr.Field.String(), fieldErr.Error.Error()))
	})
}

func (e *BadRequest) Err() error {
	if e.errorFields == nil {
		return nil
	}
	msg := e.msg
	if msg == "" {
		msg = "invalid Parameters"
	}
	err := statuserror.
		NewStatusErr("BadRequest", http.StatusBadRequest*1e6, msg).
		AppendErrorFields(e.errorFields...)

	if e.errTalk {
		err = err.EnableErrTalk()
	}
	return err
}

type PostValidator interface {
	PostValidate(badRequest *BadRequest)
}

func (t *RequestTransformer) DecodeFrom(info *RequestInfo, meta *courier.OperatorFactory, v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	typ := reflectx.Deref(rv.Type())
	if !typ.ConvertibleTo(t.Type) {
		return fmt.Errorf("unmatched request transformer, need %s but got %s", t.Type, typ)
	}

	badRequestError := &BadRequest{}

	getValues := func(in string, name string) []string {
		if in == "meta" {
			if meta.Params != nil {
				return meta.Params[name]
			}
			return []string{}
		}
		return info.Values(in, name)
	}

	transformers.NamedStructFieldValueRange(reflect.Indirect(rv), func(fieldValue reflect.Value, field *reflect.StructField) {
		param := t.Parameters[field.Name]
		if param == nil {
			return
		}

		if param.In == "body" {
			if err := param.Transformer.DecodeFromReader(info.Body(), fieldValue, textproto.MIMEHeader(info.Request.Header)); err != nil && err != io.EOF {
				badRequestError.AddErr(err, param.In, param.Name)
			}
		} else {
			maybe := transformers.NewMaybeTransformer(param.Transformer, &param.CommonTransformOption)
			values := getValues(param.In, param.Name)

			if param.Explode {
				lenOfValues := len(values)

				if param.Omitempty && lenOfValues == 0 {
					return
				}

				if field.Type.Kind() == reflect.Slice {
					fieldValue.Set(reflect.MakeSlice(field.Type, lenOfValues, lenOfValues))
				}

				for idx := 0; idx < fieldValue.Len(); idx++ {
					if lenOfValues > idx {
						if err := maybe.DecodeFromReader(bytes.NewBufferString(values[idx]), fieldValue.Index(idx)); err != nil {
							badRequestError.AddErr(err, param.In, param.Name, idx)
						}
					}
				}
			} else {
				value := ""
				if len(values) > 0 {
					value = values[0]
				}
				if err := maybe.DecodeFromReader(bytes.NewBufferString(value), fieldValue); err != nil {
					badRequestError.AddErr(err, param.In, param.Name)
				}
			}
		}

		if param.Validator != nil {
			if err := param.Validator.Validate(fieldValue); err != nil {
				badRequestError.AddErr(err, param.In, param.Name)
			}
		}

	}, "in")

	if postValidator, ok := rv.Interface().(PostValidator); ok {
		postValidator.PostValidate(badRequestError)
	}

	return badRequestError.Err()
}

func NewRequestParameter(name string, in string) *RequestParameter {
	return &RequestParameter{
		Name: name,
		In:   in,
	}
}

type RequestParameter struct {
	Name string
	In   string
	transformers.CommonTransformOption
	Transformer transformers.Transformer
	Validator   validator.Validator
}

func NewRequestInfo(r *http.Request) *RequestInfo {
	params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		params = httprouter.Params{}
	}

	return &RequestInfo{
		Request:    r,
		params:     params,
		receivedAt: time.Now(),
	}
}

type RequestInfo struct {
	Request    *http.Request
	receivedAt time.Time
	query      url.Values
	cookies    []*http.Cookie
	params     httprouter.Params
}

func (info *RequestInfo) Value(in string, name string) string {
	values := info.Values(in, name)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (info *RequestInfo) Values(in string, name string) []string {
	switch in {
	case "path":
		v := info.Param(name)
		if v == "" {
			return []string{}
		}
		return []string{v}
	case "query":
		return info.QueryValues(name)
	case "cookie":
		return info.CookieValues(name)
	case "header":
		return info.HeaderValues(name)
	}
	return []string{}
}

func (info *RequestInfo) Param(name string) string {
	return info.params.ByName(name)
}

func (info *RequestInfo) QueryValues(name string) []string {
	if info.query == nil {
		info.query = info.Request.URL.Query()
	}
	return info.query[name]
}

func (info *RequestInfo) HeaderValues(name string) []string {
	return info.Request.Header[textproto.CanonicalMIMEHeaderKey(name)]
}

func (info *RequestInfo) CookieValues(name string) []string {
	values := make([]string, 0)
	if info.cookies == nil {
		info.cookies = info.Request.Cookies()
	}
	for _, c := range info.cookies {
		if c.Name == name {
			if c.Expires.IsZero() {
				values = append(values, c.Value)
			} else if c.Expires.After(info.receivedAt) {
				values = append(values, c.Value)
			}
		}
	}
	return values
}

func (info *RequestInfo) Body() io.Reader {
	return info.Request.Body
}

func OperatorParamsFromStruct(v interface{}) map[string][]string {
	rv := reflectx.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		panic(fmt.Errorf("must struct"))
	}

	params := map[string][]string{}

	transformers.NamedStructFieldValueRange(rv, func(fieldValue reflect.Value, field *reflect.StructField) {
		tag, ok := field.Tag.Lookup("in")
		if !ok || tag != "meta" {
			return
		}

		fieldDisplayName, _, _ := typesutil.FieldDisplayName(field.Tag, "name", field.Name)

		if !ast.IsExported(field.Name) || fieldDisplayName == "-" {
			return
		}

		values := make([]string, 0)

		switch fieldValue.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < fieldValue.Len(); i++ {
				v, err := reflectx.MarshalText(fieldValue.Index(i).Interface())
				if err == nil {
					values = append(values, string(v))
				}
			}
		default:
			v, err := reflectx.MarshalText(fieldValue.Interface())
			if err == nil {
				values = append(values, string(v))
			}
		}

		params[fieldDisplayName] = values
	})

	return params
}
