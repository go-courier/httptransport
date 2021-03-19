package httptransport

import (
	"context"
	"io"
	"net/http"
	"reflect"

	"github.com/pkg/errors"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/metax"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/statuserror"
)

func NewHttpRouteHandler(serviceMeta *ServiceMeta, httpRoute *HttpRouteMeta, requestTransformerMgr *RequestTransformerMgr) *HttpRouteHandler {
	operatorFactories := httpRoute.OperatorFactoryWithRouteMetas

	if len(operatorFactories) == 0 {
		panic(errors.Errorf("missing valid operator"))
	}

	requestTransformers := make([]*RequestTransformer, len(operatorFactories))

	for i := range operatorFactories {
		opFactory := operatorFactories[i]
		rt, err := requestTransformerMgr.NewRequestTransformer(context.Background(), opFactory.Type)
		if err != nil {
			panic(err)
		}
		requestTransformers[i] = rt
	}

	return &HttpRouteHandler{
		RequestTransformerMgr: requestTransformerMgr,
		HttpRouteMeta:         httpRoute,

		serviceMeta:         serviceMeta,
		requestTransformers: requestTransformers,
	}
}

type HttpRouteHandler struct {
	*RequestTransformerMgr
	*HttpRouteMeta
	serviceMeta         *ServiceMeta
	requestTransformers []*RequestTransformer
}

type contextKeyOperationID int

func ContextWithOperationID(ctx context.Context, operationID string) context.Context {
	return context.WithValue(ctx, contextKeyOperationID(1), operationID)
}

func OperationIDFromContext(ctx context.Context) string {
	return ctx.Value(contextKeyOperationID(1)).(string)
}

type contextKeyOperatorFactory int

func ContextWithOperatorFactory(ctx context.Context, om *courier.OperatorFactory) context.Context {
	return context.WithValue(ctx, contextKeyOperatorFactory(1), om)
}

func OperatorFactoryFromContext(ctx context.Context) *courier.OperatorFactory {
	v, _ := ctx.Value(contextKeyOperatorFactory(1)).(*courier.OperatorFactory)
	return v
}

func (handler *HttpRouteHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	operationID := handler.OperatorFactoryWithRouteMetas[len(handler.OperatorFactoryWithRouteMetas)-1].ID

	ctx := r.Context()
	ctx = ContextWithHttpRequest(ctx, r)
	ctx = ContextWithServiceMeta(ctx, *handler.serviceMeta)
	ctx = ContextWithOperationID(ctx, operationID)

	spanName := handler.serviceMeta.String() + "/" + operationID

	ctx = metax.ContextWithMeta(ctx, metax.Meta{
		"operator": {spanName},
	})

	rw.Header().Set("X-Meta", spanName)

	requestInfo := NewRequestInfo(r)

	for i := range handler.OperatorFactoryWithRouteMetas {
		opFactory := handler.OperatorFactoryWithRouteMetas[i]

		if opFactory.NoOutput {
			continue
		}

		op := opFactory.New()

		ctx = ContextWithOperatorFactory(ctx, opFactory.OperatorFactory)

		rt := handler.requestTransformers[i]
		if rt != nil {
			err := rt.DecodeFrom(requestInfo, opFactory.OperatorFactory, op)
			if err != nil {
				handler.writeErr(rw, r, err)
				return
			}
		}

		result, err := op.Output(ctx)

		if err != nil {
			handler.writeErr(rw, r, err)
			return
		}

		if !opFactory.IsLast {
			if c, ok := result.(context.Context); ok {
				ctx = c
			} else {
				// set result in context with key of operator name
				ctx = context.WithValue(ctx, opFactory.ContextKey, result)
			}
			continue
		}

		handler.writeResp(rw, r, result)
	}
}

func (handler *HttpRouteHandler) resolveTransformer(response *httpx.Response) (string, httpx.Encode, error) {
	transformer, err := handler.TransformerMgr.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf(response.Value)), transformers.TransformerOption{
		MIME: response.ContentType,
	})
	if err != nil {
		return "", nil, err
	}

	return transformer.String(), func(w io.Writer, v interface{}) error {
		_, err := transformer.EncodeToWriter(w, v)
		return err
	}, nil
}

func (handler *HttpRouteHandler) writeResp(rw http.ResponseWriter, r *http.Request, resp interface{}) {
	err := httpx.ResponseFrom(resp).WriteTo(rw, r, handler.resolveTransformer)
	if err != nil {
		handler.writeErr(rw, r, err)
	}
}

func (handler *HttpRouteHandler) writeErr(rw http.ResponseWriter, r *http.Request, err error) {
	resp, ok := err.(*httpx.Response)
	if !ok {
		resp = httpx.ResponseFrom(err)
	}

	if statusErr, ok := statuserror.IsStatusErr(resp.Unwrap()); ok {
		err := statusErr.AppendSource(handler.serviceMeta.String())

		if rwe, ok := rw.(ResponseWithError); ok {
			rwe.WriteErrer(err)
		}

		resp.Value = err
	}

	errForWrite := resp.WriteTo(rw, r, handler.resolveTransformer)
	if errForWrite != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte("courier write err failed:" + errForWrite.Error()))
	}
}

type ResponseWithError interface {
	WriteErrer(err error)
}
