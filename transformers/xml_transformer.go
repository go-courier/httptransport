package transformers

import (
	"encoding/xml"
	"io"
	"mime"
	"net/textproto"
	"reflect"

	"github.com/go-courier/reflectx/typesutil"
)

func init() {
	TransformerMgrDefault.Register(&XMLTransformer{})
}

type XMLTransformer struct {
}

func (XMLTransformer) Names() []string {
	return []string{"application/xml", "xml"}
}

func (t *XMLTransformer) String() string {
	return t.Names()[0]
}

func (XMLTransformer) New(typesutil.Type, TransformerMgr) (Transformer, error) {
	return &XMLTransformer{}, nil
}

func (t *XMLTransformer) EncodeToWriter(w io.Writer, v interface{}) (string, error) {
	if rv, ok := v.(reflect.Value); ok {
		v = rv.Interface()
	}
	if err := xml.NewEncoder(w).Encode(v); err != nil {
		return "", err
	}
	return mime.FormatMediaType(t.String(), map[string]string{
		"charset": "utf-8",
	}), nil
}

func (XMLTransformer) DecodeFromReader(r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	if rv, ok := v.(reflect.Value); ok {
		if rv.Kind() != reflect.Ptr && rv.CanAddr() {
			rv = rv.Addr()
		}
		v = rv.Interface()
	}
	d := xml.NewDecoder(r)
	err := d.Decode(v)
	if err != nil {
		// todo resolve field path by InputOffset()
		return err
	}
	return nil
}
