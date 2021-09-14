package transformers

import (
	"context"
	"io"
	"net/textproto"
	"net/url"
	"reflect"
	"sync"

	contextx "github.com/go-courier/x/context"

	typesx "github.com/go-courier/x/types"
	"github.com/pkg/errors"
)

func NewTransformer(ctx context.Context, tpe typesx.Type, opt TransformerOption) (Transformer, error) {
	return TransformerMgrFromContext(ctx).NewTransformer(ctx, tpe, opt)
}

type TransformerMgr interface {
	NewTransformer(context.Context, typesx.Type, TransformerOption) (Transformer, error)
}

type contextKeyTransformerMgr struct{}

func ContextWithTransformerMgr(ctx context.Context, mgr TransformerMgr) context.Context {
	return contextx.WithValue(ctx, contextKeyTransformerMgr{}, mgr)
}

func TransformerMgrFromContext(ctx context.Context) TransformerMgr {
	if mgr, ok := ctx.Value(contextKeyTransformerMgr{}).(TransformerMgr); ok {
		return mgr
	}
	return TransformerMgrDefault
}

type Transformer interface {
	// name or alias of transformer
	// prefer using some keyword about content-type
	// first must validate content-type
	Names() []string
	// create transformer new transformer instance by type
	// in this step will to check transformer is valid for type
	New(context.Context, typesx.Type) (Transformer, error)

	// EncodeTo
	// if w implement interface { Header() http.Header }
	// Content-Type will be set
	EncodeTo(ctx context.Context, w io.Writer, v interface{}) (err error)

	// DecodeFrom
	DecodeFrom(ctx context.Context, r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error
}

type TransformerOption struct {
	Name string
	MIME string
	CommonTransformOption
}

type CommonTransformOption struct {
	// when enable
	// should ignore value when value is empty
	Omitempty bool
	Explode   bool
}

func (op TransformerOption) String() string {
	values := url.Values{}

	if op.Name != "" {
		values.Add("Name", op.Name)
	}

	if op.MIME != "" {
		values.Add("MIME", op.MIME)
	}

	if op.Omitempty {
		values.Add("Omitempty", "true")
	}

	if op.Explode {
		values.Add("Explode", "true")
	}

	return values.Encode()
}

var TransformerMgrDefault = &TransformerFactory{}

type TransformerFactory struct {
	transformerSet map[string]Transformer
	cache          sync.Map
}

func (c *TransformerFactory) Register(transformers ...Transformer) {
	if c.transformerSet == nil {
		c.transformerSet = map[string]Transformer{}
	}
	for i := range transformers {
		transformer := transformers[i]
		for _, name := range transformer.Names() {
			c.transformerSet[name] = transformer
		}
	}
}

func (c *TransformerFactory) NewTransformer(ctx context.Context, typ typesx.Type, opt TransformerOption) (Transformer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	key := typesx.FullTypeName(typ) + opt.String()

	if v, ok := c.cache.Load(key); ok {
		return v.(Transformer), nil
	}

	if opt.MIME == "" {
		indirectType := typesx.Deref(typ)

		switch indirectType.Kind() {
		case reflect.Slice:
			if indirectType.Elem().PkgPath() == "" && indirectType.Elem().Kind() == reflect.Uint8 {
				// bytes
				opt.MIME = "plain"
			} else {
				opt.MIME = "json"
			}
		case reflect.Struct:
			// *mime/multipart.FileHeader
			if indirectType.PkgPath() == "mime/multipart" && indirectType.Name() == "FileHeader" {
				opt.MIME = "octet-stream"
			} else {
				opt.MIME = "json"
			}
		case reflect.Map, reflect.Array:
			opt.MIME = "json"
		default:
			opt.MIME = "plain"
		}

		if _, ok := typesx.EncodingTextMarshalerTypeReplacer(typ); ok {
			opt.MIME = "plain"
		}
	}

	if ct, ok := c.transformerSet[opt.MIME]; ok {
		contentTransformer, err := ct.New(ContextWithTransformerMgr(ctx, c), typ)
		if err != nil {
			return nil, err
		}
		c.cache.Store(key, contentTransformer)
		return contentTransformer, nil
	}

	return nil, errors.Errorf("fmt %s is not supported for content transformer", key)
}
