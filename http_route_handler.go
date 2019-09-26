package httptransport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/metax"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/statuserror"
)

func NewHttpRouteHandler(serviceMeta *ServiceMeta, httpRoute *HttpRouteMeta, requestTransformerMgr *RequestTransformerMgr) *HttpRouteHandler {
	operatorFactories := httpRoute.OperatorFactories()
	if len(operatorFactories) == 0 {
		panic(fmt.Errorf("missing valid operator"))
	}

	requestTransformers := make([]*RequestTransformer, len(operatorFactories))
	for i := range operatorFactories {
		opFactory := operatorFactories[i]
		rt, err := requestTransformerMgr.NewRequestTransformer(nil, opFactory.Type)
		if err != nil {
			panic(err)
		}
		requestTransformers[i] = rt
	}

	return &HttpRouteHandler{
		RequestTransformerMgr: requestTransformerMgr,
		HttpRouteMeta:         httpRoute,

		serviceMeta:         serviceMeta,
		operatorFactories:   operatorFactories,
		requestTransformers: requestTransformers,
	}
}

type HttpRouteHandler struct {
	*RequestTransformerMgr
	*HttpRouteMeta

	serviceMeta         *ServiceMeta
	operatorFactories   []*courier.OperatorMeta
	requestTransformers []*RequestTransformer
}

func ContextWithOperationID(ctx context.Context, operationID string) context.Context {
	return context.WithValue(ctx, "courier.OperationID", operationID)
}

func OperationIDFromContext(ctx context.Context) string {
	return ctx.Value("courier.OperationID").(string)
}

func ContextWithOperatorMeta(ctx context.Context, om *courier.OperatorMeta) context.Context {
	return context.WithValue(ctx, "courier.OperatorMeta", om)
}

func OperatorMetaFromContext(ctx context.Context) *courier.OperatorMeta {
	v, _ := ctx.Value("courier.OperatorMeta").(*courier.OperatorMeta)
	return v
}

func (handler *HttpRouteHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	operationID := handler.operatorFactories[len(handler.operatorFactories)-1].Type.Name()

	ctx := r.Context()
	ctx = ContextWithHttpRequest(ctx, r)
	ctx = ContextWithServiceMeta(ctx, *handler.serviceMeta)
	ctx = ContextWithOperationID(ctx, operationID)
	ctx = metax.ContextWithMeta(ctx, metax.Meta{
		"operator": {handler.serviceMeta.String() + "#" + operationID},
	})

	rw.Header().Set("X-Meta", metax.MetaFromContext(ctx).String())

	requestInfo := NewRequestInfo(r)

	for i := range handler.operatorFactories {
		opFactory := handler.operatorFactories[i]

		op := opFactory.New()

		ctx = ContextWithOperatorMeta(ctx, opFactory)

		rt := handler.requestTransformers[i]
		if rt != nil {
			err := rt.DecodeFrom(requestInfo, opFactory, op)
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

func (handler *HttpRouteHandler) writeResp(rw http.ResponseWriter, r *http.Request, resp interface{}) {
	if err := httpx.ResponseFrom(resp).WriteTo(rw, r, func(w io.Writer, response *httpx.Response) error {
		transformer, err := handler.TransformerMgr.NewTransformer(nil, typesutil.FromRType(reflect.TypeOf(response.Value)), transformers.TransformerOption{
			MIME: response.ContentType,
		})
		if err != nil {
			return err
		}
		if _, err := transformer.EncodeToWriter(w, response.Value); err != nil {
			return err
		}
		return nil
	}); err != nil {
		handler.writeErr(rw, r, err)
	}
}

func (handler *HttpRouteHandler) writeErr(rw http.ResponseWriter, r *http.Request, err error) {
	if _, ok := err.(httpx.RedirectDescriber); !ok {
		err = statuserror.FromErr(err).AppendSource(handler.serviceMeta.String())
	}
	handler.writeResp(rw, r, err)
}
