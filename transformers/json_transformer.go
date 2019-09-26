package transformers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net/textproto"
	"reflect"
	"strconv"

	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/validator/errors"
)

func init() {
	TransformerMgrDefault.Register(&JSONTransformer{})
}

type JSONTransformer struct {
}

func (JSONTransformer) Names() []string {
	return []string{"application/json", "json"}
}

func (JSONTransformer) NamedByTag() string {
	return "json"
}

func (transformer *JSONTransformer) String() string {
	return transformer.Names()[0]
}

func (JSONTransformer) New(context.Context, typesutil.Type) (Transformer, error) {
	return &JSONTransformer{}, nil
}

func (transformer *JSONTransformer) EncodeToWriter(w io.Writer, v interface{}) (string, error) {
	if rv, ok := v.(reflect.Value); ok {
		v = rv.Interface()
	}

	return superWrite(w, func(w io.Writer) error {
		return json.NewEncoder(w).Encode(v)
	}, mime.FormatMediaType(transformer.String(), map[string]string{
		"charset": "utf-8",
	}))
}

func (JSONTransformer) DecodeFromReader(r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	if rv, ok := v.(reflect.Value); ok {
		if rv.Kind() != reflect.Ptr && rv.CanAddr() {
			rv = rv.Addr()
		}
		v = rv.Interface()
	}

	data, errForRead := ioutil.ReadAll(r)
	if errForRead != nil {
		return errForRead
	}

	dec := json.NewDecoder(bytes.NewBuffer(data))
	err := dec.Decode(v)
	if err != nil {
		switch e := err.(type) {
		case *json.UnmarshalTypeError:
			errSet := errors.NewErrorSet("")
			errSet.AddErr(e, location(data, int(e.Offset)))
			return errSet.Err()
		case *json.SyntaxError:
			return e
		default:
			offset := reflect.ValueOf(dec).Elem().Field(2 /*d*/).Field(1 /*off*/).Int()
			if offset > 0 {
				errSet := errors.NewErrorSet("")
				errSet.AddErr(e, location(data, int(offset-1)))
				return errSet.Err()
			}
			return e
		}
	}
	return nil
}

func location(data []byte, offset int) string {
	i := 0
	arrayPaths := map[string]bool{}
	arrayIdxSet := map[string]int{}
	pathWalker := &PathWalker{}

	markObjectKey := func() {
		jsonKey, l := nextString(data[i:])
		i += l

		if i < int(offset) && len(jsonKey) > 0 {
			key, _ := strconv.Unquote(string(jsonKey))
			pathWalker.Enter(key)
		}
	}

	markArrayIdx := func(path string) {
		if arrayPaths[path] {
			arrayIdxSet[path]++
		} else {
			arrayPaths[path] = true
		}
		pathWalker.Enter(arrayIdxSet[path])
	}

	for i < offset {
		i += nextToken(data[i:])
		char := data[i]

		switch char {
		case '"':
			_, l := nextString(data[i:])
			i += l
		case '[', '{':
			i++

			if char == '[' {
				markArrayIdx(pathWalker.String())
			} else {
				markObjectKey()
			}
		case '}', ']', ',':
			i++
			pathWalker.Exit()

			if char == ',' {
				path := pathWalker.String()

				if _, ok := arrayPaths[path]; ok {
					markArrayIdx(path)
				} else {
					markObjectKey()
				}
			}
		default:
			i++
		}
	}

	return pathWalker.String()
}

func nextToken(data []byte) int {
	for i, c := range data {
		switch c {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}
	return -1
}

func nextString(data []byte) (finalData []byte, l int) {
	quoteStartAt := -1
	for i, c := range data {
		switch c {
		case '"':
			if i > 0 && string(data[i-1]) == "\\" {
				continue
			}
			if quoteStartAt >= 0 {
				return data[quoteStartAt : i+1], i + 1
			} else {
				quoteStartAt = i
			}
		default:
			continue
		}
	}
	return nil, 0
}
