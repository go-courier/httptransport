package httptransport

import (
	"bytes"
	"context"
	"io"
	"mime"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"sort"
	"sync"

	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/httptransport/validator"
	verrors "github.com/go-courier/httptransport/validator"
	"github.com/go-courier/statuserror"
	contextx "github.com/go-courier/x/context"
	reflectx "github.com/go-courier/x/reflect"
	typex "github.com/go-courier/x/types"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
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
	return mgr.NewRequestWithContext(context.Background(), method, rawUrl, v)
}

func (mgr *RequestTransformerMgr) NewRequestWithContext(ctx context.Context, method string, rawUrl string, v interface{}) (*http.Request, error) {
	if v == nil {
		return http.NewRequestWithContext(ctx, method, rawUrl, nil)
	}
	rt, err := mgr.NewRequestTransformer(AsRequestOut(ctx), reflect.TypeOf(v))
	if err != nil {
		return nil, err
	}
	return rt.NewRequestWithContext(ctx, method, rawUrl, v)
}

func (mgr *RequestTransformerMgr) NewRequestTransformer(ctx context.Context, typ reflect.Type) (*RequestTransformer, error) {
	if v, ok := mgr.cache.Load(typ); ok {
		return v.(*RequestTransformer), nil
	}
	rt, err := mgr.newRequestTransformer(ctx, typ)
	if err != nil {
		return nil, err
	}
	mgr.cache.Store(typ, rt)
	return rt, nil
}

type contextKeyForRequestOut struct{}

func AsRequestOut(ctx context.Context) context.Context {
	return contextx.WithValue(ctx, contextKeyForRequestOut{}, true)
}

func IsRequestOut(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if t, ok := ctx.Value(contextKeyForRequestOut{}).(bool); ok {
		return t
	}
	return false
}

func (mgr *RequestTransformerMgr) newRequestTransformer(ctx context.Context, typ reflect.Type) (*RequestTransformer, error) {
	rt := &RequestTransformer{}

	rt.InParameters = map[string][]transformers.RequestParameter{}
	rt.Type = reflectx.Deref(typ)

	ctx = transformers.ContextWithTransformerMgr(ctx, mgr.TransformerMgr)
	ctx = validator.ContextWithValidatorMgr(ctx, mgr.ValidatorMgr)

	err := transformers.EachRequestParameter(ctx, typex.FromRType(rt.Type), func(rp *transformers.RequestParameter) {
		if rp.In == "" {
			return
		}
		rt.InParameters[rp.In] = append(rt.InParameters[rp.In], *rp)
	})

	return rt, err
}

type RequestTransformerMgr struct {
	validator.ValidatorMgr
	transformers.TransformerMgr
	cache sync.Map
}

type RequestTransformer struct {
	Type         reflect.Type
	InParameters map[string][]transformers.RequestParameter
}

func (t *RequestTransformer) NewRequest(method string, rawUrl string, v interface{}) (*http.Request, error) {
	return t.NewRequestWithContext(context.Background(), method, rawUrl, v)
}

func (t *RequestTransformer) NewRequestWithContext(ctx context.Context, method string, rawUrl string, v interface{}) (*http.Request, error) {
	if v == nil {
		return http.NewRequestWithContext(ctx, method, rawUrl, nil)
	}

	typ := reflectx.Deref(reflect.TypeOf(v))

	if t.Type != typ {
		return nil, errors.Errorf("unmatched request transformer, need %s but got %s", t.Type, typ)
	}

	errSet := verrors.NewErrorSet("")

	params := httprouter.Params{}
	query := url.Values{}
	header := http.Header{}
	cookies := url.Values{}
	body := bytes.NewBuffer(nil)

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	rv = reflectx.Indirect(rv)

	for in := range t.InParameters {
		parameters := t.InParameters[in]

		for i := range parameters {
			p := parameters[i]

			fieldValue := p.FieldValue(rv)

			if !fieldValue.IsValid() {
				continue
			}

			if p.In == "body" {
				err := p.Transformer.EncodeTo(ctx, transformers.WriterWithHeader(body, header), fieldValue)
				if err != nil {
					errSet.AddErr(err, p.Name)
				}
				continue
			}

			writers := transformers.NewStringBuilders()

			if err := transformers.NewTransformerSuper(p.Transformer, &p.TransformerOption.CommonTransformOption).EncodeTo(ctx, writers, fieldValue); err != nil {
				errSet.AddErr(err, p.Name)
				continue
			}

			values := writers.StringSlice()

			switch p.In {
			case "path":
				params = append(params, httprouter.Param{
					Key:   p.Name,
					Value: values[0],
				})
			case "query":
				query[p.Name] = values
			case "header":
				header[textproto.CanonicalMIMEHeaderKey(p.Name)] = values
			case "cookie":
				cookies[p.Name] = values
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, rawUrl, nil)
	if err != nil {
		return nil, err
	}

	if len(params) > 0 {
		req = req.WithContext(contextx.WithValue(req.Context(), httprouter.ParamsKey, params))
		req.URL.Path = transformers.NewPathnamePattern(req.URL.Path).Stringify(params)
	}

	if len(query) > 0 {
		if method == http.MethodGet && ShouldQueryInBodyForHttpGet(ctx) {
			header.Set("Content-Type", mime.FormatMediaType("application/x-www-form-urlencoded", map[string]string{
				"param": "value",
			}))
			body = bytes.NewBufferString(query.Encode())
		} else {
			req.URL.RawQuery = query.Encode()
		}
	}

	req.Header = header

	if n := len(cookies); n > 0 {
		names := make([]string, n)
		i := 0
		for name := range cookies {
			names[i] = name
			i++
		}
		sort.Strings(names)

		for _, name := range names {
			values := cookies[name]
			for i := range values {
				req.AddCookie(&http.Cookie{
					Name:  name,
					Value: values[i],
				})
			}
		}
	}

	if n := int64(body.Len()); n != 0 {
		req.ContentLength = n
		rc := io.NopCloser(body)
		req.Body = rc
		req.GetBody = func() (io.ReadCloser, error) {
			return rc, nil
		}
	}

	return req, nil
}

type WithFromRequestInfo interface {
	FromRequestInfo(req *httpx.RequestInfo) error
}

func (t *RequestTransformer) DecodeAndValidate(ctx context.Context, info httpx.RequestInfo, v interface{}) error {
	if err := t.DecodeFromRequestInfo(ctx, info, v); err != nil {
		return err
	}
	return t.validate(v)
}

func (t *RequestTransformer) DecodeFromRequestInfo(ctx context.Context, info httpx.RequestInfo, v interface{}) error {
	if canValidate, ok := v.(httpx.WithFromRequestInfo); ok {
		if err := canValidate.FromRequestInfo(info); err != nil {
			if est := err.(interface {
				ToFieldErrors() statuserror.ErrorFields
			}); ok {
				if errorFields := est.ToFieldErrors(); len(errorFields) > 0 {
					return (&badRequest{errorFields: errorFields}).Err()
				}
			}
			return err
		}
		return nil
	}

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if rv.Kind() != reflect.Ptr {
		return errors.Errorf("decode target must be an ptr value")
	}

	rv = reflectx.Indirect(rv)

	if tpe := rv.Type(); tpe != t.Type {
		return errors.Errorf("unmatched request transformer, need %s but got %s", t.Type, tpe)
	}

	errSet := validator.NewErrorSet()

	for in := range t.InParameters {
		parameters := t.InParameters[in]

		for i := range parameters {
			param := parameters[i]

			if param.In == "body" {
				body := info.Body()
				if err := param.Transformer.DecodeFrom(ctx, body, param.FieldValue(rv).Addr(), textproto.MIMEHeader(info.Header())); err != nil && err != io.EOF {
					errSet.AddErr(err, validator.Location(param.In))
				}
				body.Close()
				continue
			}

			var values []string

			if param.In == "meta" {
				params := OperatorFactoryFromContext(ctx).Params
				if params != nil {
					values = params[param.Name]
				}
			} else {
				values = info.Values(param.In, param.Name)
			}

			if len(values) > 0 {
				if err := transformers.NewTransformerSuper(param.Transformer, &param.TransformerOption.CommonTransformOption).DecodeFrom(ctx, transformers.NewStringReaders(values), param.FieldValue(rv).Addr()); err != nil {
					errSet.AddErr(err, validator.Location(param.In), param.Name)
				}
			}
		}
	}

	if errSet.Err() == nil {
		return nil
	}

	return (&badRequest{errorFields: errSet.ToErrorFields()}).Err()
}

func (t *RequestTransformer) validate(v interface{}) error {
	if canValidate, ok := v.(interface{ Validate() error }); ok {
		if err := canValidate.Validate(); err != nil {
			if est := err.(interface {
				ToFieldErrors() statuserror.ErrorFields
			}); ok {
				if errorFields := est.ToFieldErrors(); len(errorFields) > 0 {
					return (&badRequest{errorFields: errorFields}).Err()
				}
			}
			return err
		}
		return nil
	}

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	errSet := validator.NewErrorSet()

	for in := range t.InParameters {
		parameters := t.InParameters[in]

		for i := range parameters {
			param := parameters[i]

			if param.Validator != nil {
				if err := param.Validator.Validate(param.FieldValue(rv)); err != nil {
					if param.In == "body" {
						errSet.AddErr(err, validator.Location(param.In))
					} else {
						errSet.AddErr(err, validator.Location(param.In), param.Name)
					}
				}
			}
		}
	}

	br := &badRequest{errorFields: errSet.ToErrorFields()}

	// TODO deprecated
	if postValidator, ok := rv.Interface().(PostValidator); ok {
		postValidator.PostValidate(br)
	}

	if errSet.Err() == nil {
		return nil
	}

	return br.Err()
}

type PostValidator interface {
	PostValidate(badRequest BadRequestError)
}

type BadRequestError interface {
	EnableErrTalk()
	SetMsg(msg string)
	AddErr(err error, nameOrIdx ...interface{})
}

type badRequest struct {
	errorFields statuserror.ErrorFields
	errTalk     bool
	msg         string
}

func (e *badRequest) EnableErrTalk() {
	e.errTalk = true
}

func (e *badRequest) SetMsg(msg string) {
	e.msg = msg
}

func (e *badRequest) AddErr(err error, nameOrIdx ...interface{}) {
	if len(nameOrIdx) > 1 {
		e.errorFields = append(e.errorFields, &statuserror.ErrorField{
			In:    nameOrIdx[0].(string),
			Field: validator.KeyPath(nameOrIdx[1:]).String(),
			Msg:   err.Error(),
		})
	}
}

func (e *badRequest) Err() error {
	if len(e.errorFields) == 0 {
		return nil
	}

	msg := e.msg
	if msg == "" {
		msg = "invalid parameters"
	}

	err := statuserror.Wrap(errors.New(""), http.StatusBadRequest, "badRequest").WithMsg(msg).AppendErrorFields(e.errorFields...)

	if e.errTalk {
		err = err.EnableErrTalk()
	}

	return err
}
