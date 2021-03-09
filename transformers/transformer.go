package transformers

import (
	"bytes"
	"context"
	"io"
	"net/textproto"
	"net/url"
	"reflect"
	"sync"

	"github.com/pkg/errors"

	"github.com/go-courier/reflectx"
	"github.com/go-courier/reflectx/typesutil"
)

type TransformerMgr interface {
	NewTransformer(context.Context, typesutil.Type, TransformerOption) (Transformer, error)
}

type contextKeyTransformerMgr int

func ContextWithTransformerMgr(ctx context.Context, mgr TransformerMgr) context.Context {
	return context.WithValue(ctx, contextKeyTransformerMgr(1), mgr)
}

func TransformerMgrFromContext(ctx context.Context) TransformerMgr {
	return ctx.Value(contextKeyTransformerMgr(1)).(TransformerMgr)
}

type Transformer interface {
	// name or alias of transformer
	// prefer using some keyword about content-type
	Names() []string
	// create transformer new transformer instance by type
	// in this step will to check transformer is valid for type
	New(context.Context, typesutil.Type) (Transformer, error)

	// named by tag
	NamedByTag() string

	// encode to writer
	EncodeToWriter(w io.Writer, v interface{}) (mediaType string, err error)
	// decode from reader
	DecodeFromReader(r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error

	// Content-Type
	String() string
}

var TagNameKey = "name"
var TagMIMEKey = "mime"

func TransformerOptionFromStructField(field typesutil.StructField) TransformerOption {
	tags := field.Tag()

	tagName := tags.Get(TagNameKey)

	opt := TransformerOption{
		MIME: tags.Get(TagMIMEKey),
	}

	if tagName != "" && tagName != "-" {
		name, flags := TagValueAndFlagsByTagString(tagName)
		opt.FieldName = name
		if flags["omitempty"] {
			opt.Omitempty = true
		}
	}

	if opt.FieldName == "" {
		opt.FieldName = field.Name()
	}

	return opt
}

type TransformerOption struct {
	FieldName string
	MIME      string
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

	if op.FieldName != "" {
		values.Add("FieldName", op.FieldName)
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

func (c *TransformerFactory) NewTransformer(ctx context.Context, typ typesutil.Type, opt TransformerOption) (Transformer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	key := typesutil.FullTypeName(typ) + opt.String()

	if v, ok := c.cache.Load(key); ok {
		return v.(Transformer), nil
	}

	if opt.MIME == "" {
		indirectType := typesutil.Deref(typ)

		if IsBytes(indirectType) {
			opt.MIME = "plain"
		} else {
			switch indirectType.Kind() {
			case reflect.Struct, reflect.Slice, reflect.Map, reflect.Array:
				opt.MIME = "json"
			default:
				opt.MIME = "plain"
			}
		}

		if _, ok := typesutil.EncodingTextMarshalerTypeReplacer(typ); ok {
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

type Adder interface {
	Add(key string, value string)
}

func NewMaybeTransformer(transformer Transformer, opt *CommonTransformOption) *MaybeTransformer {
	return &MaybeTransformer{
		transformer: transformer,
		opt:         opt,
	}
}

type MaybeTransformer struct {
	transformer Transformer
	opt         *CommonTransformOption
}

func (t *MaybeTransformer) Add(key string, v interface{}, adder Adder) error {
	if !t.opt.Explode && t.opt.Omitempty && reflectx.IsEmptyValue(v) {
		return nil
	}
	buf := bytes.NewBuffer(nil)
	if _, err := t.transformer.EncodeToWriter(buf, v); err != nil {
		return err
	}
	adder.Add(key, buf.String())
	return nil
}

func (t *MaybeTransformer) DecodeFromReader(r io.Reader, v interface{}) error {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, r); err != nil {
		return err
	}
	if !t.opt.Explode && t.opt.Omitempty && buf.Len() == 0 {
		return nil
	}
	return t.transformer.DecodeFromReader(buf, v)
}

func (t *MaybeTransformer) EncodeToWriter(w io.Writer, v interface{}) (string, error) {
	if !t.opt.Explode && t.opt.Omitempty && reflectx.IsEmptyValue(v) {
		return "", nil
	}
	return t.transformer.EncodeToWriter(w, v)
}
