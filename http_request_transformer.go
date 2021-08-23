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

	typex "github.com/go-courier/x/types"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/httptransport/validator"
	verrors "github.com/go-courier/httptransport/validator"
	"github.com/go-courier/statuserror"
	contextx "github.com/go-courier/x/context"
	reflectx "github.com/go-courier/x/reflect"
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

			if err := transformers.NewSupperTransformer(p.Transformer, &p.TransformerOption).EncodeTo(ctx, writers, fieldValue); err != nil {
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
	errSet := verrors.NewErrorSet("")

	if es, ok := err.(*verrors.ErrorSet); ok && in == "body" {
		errSet = es
	} else {
		errSet.AddErr(err, nameOrIdx...)
	}

	errSet.Flatten().Each(func(fieldErr *verrors.FieldError) {
		e.errorFields = append(e.errorFields, statuserror.NewErrorField(in, fieldErr.Path.String(), fieldErr.Error.Error()))
	})
}

func (e *BadRequest) Err() error {
	if e.errorFields == nil {
		return nil
	}

	msg := e.msg
	if msg == "" {
		msg = "invalid parameters"
	}

	err := statuserror.Wrap(errors.New(""), http.StatusBadRequest, "BadRequest").WithMsg(msg).AppendErrorFields(e.errorFields...)

	if e.errTalk {
		err = err.EnableErrTalk()
	}

	return err
}

func (t *RequestTransformer) DecodeFrom(ctx context.Context, info *httpx.RequestInfo, meta *courier.OperatorFactory, v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	typ := reflectx.Deref(rv.Type())
	if !typ.ConvertibleTo(t.Type) {
		return errors.Errorf("unmatched request transformer, need %s but got %s", t.Type, typ)
	}

	badRequestError := &BadRequest{}

	decodeFrom := func(param *transformers.RequestParameter, fieldValue reflect.Value) {
		switch param.In {
		case "body":
			if err := param.Transformer.DecodeFrom(ctx, info.Body(), fieldValue, textproto.MIMEHeader(info.Request.Header)); err != nil && err != io.EOF {
				badRequestError.AddErr(err, param.In, param.Name)
			}
		default:
			var values []string

			if param.In == "meta" {
				if meta.Params != nil {
					values = meta.Params[param.Name]
				}
			} else {
				values = info.Values(param.In, param.Name)
			}

			if len(values) > 0 {
				readers := transformers.NewStringReaders(values)

				if err := transformers.NewSupperTransformer(param.Transformer, &param.TransformerOption).DecodeFrom(context.Background(), readers, fieldValue); err != nil {
					badRequestError.AddErr(err, param.In, param.Name)
				}
			}
		}
	}

	for in := range t.InParameters {
		parameters := t.InParameters[in]

		for i := range parameters {
			param := parameters[i]

			fieldValue := param.FieldValue(rv)

			decodeFrom(&param, fieldValue)

			if param.Validator != nil {
				if err := param.Validator.Validate(fieldValue); err != nil {
					badRequestError.AddErr(err, param.In, param.Name)
				}
			}
		}
	}

	// TODO deprecated
	if postValidator, ok := rv.Interface().(PostValidator); ok {
		postValidator.PostValidate(badRequestError)
	}

	return badRequestError.Err()
}

type PostValidator interface {
	PostValidate(badRequest *BadRequest)
}
