package transformers

import (
	"go/ast"
	"go/types"
	"io"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"

	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/reflectx"
	"github.com/go-courier/reflectx/typesutil"
)

var (
	rtypeBytes = typesutil.FromRType(reflect.TypeOf([]byte("")))
	ttypeBytes = typesutil.FromTType(typesutil.NewTypesTypeFromReflectType(reflect.TypeOf([]byte(""))))
)

func PtrTo(typ typesutil.Type) typesutil.Type {
	switch t := typ.(type) {
	case *typesutil.RType:
		return typesutil.FromRType(reflect.PtrTo(t.Type))
	case *typesutil.TType:
		return typesutil.FromTType(types.NewPointer(t.Type))
	}
	return nil
}

func IsBytes(tpe typesutil.Type) bool {
	if tpe.Kind() == reflect.String {
		return false
	}
	switch tpe.(type) {
	case *typesutil.RType:
		return tpe.ConvertibleTo(rtypeBytes)
	case *typesutil.TType:
		return tpe.ConvertibleTo(ttypeBytes)
	}
	return false
}

func MIMEHeader(headers ...textproto.MIMEHeader) textproto.MIMEHeader {
	header := textproto.MIMEHeader{}
	for _, h := range headers {
		for k, values := range h {
			for _, v := range values {
				header.Add(k, v)
			}
		}
	}
	return header
}

func NamedStructFieldValueRange(rv reflect.Value, fn func(fieldValue reflect.Value, field *reflect.StructField), tags ...string) {
	typ := rv.Type()

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if !ast.IsExported(f.Name) {
			continue
		}

		fieldValue := rv.Field(i)

		fieldType := reflectx.Deref(f.Type)
		isStructType := fieldType.Kind() == reflect.Struct

		name, exists := f.Tag.Lookup(TagNameKey)
		if !exists {
			for i := range tags {
				_, exists = f.Tag.Lookup(tags[i])
				if exists {
					break
				}
			}
		}

		if isStructType && f.Anonymous && !exists {
			NamedStructFieldValueRange(fieldValue, fn)
			continue
		}

		if name != "-" {
			fn(fieldValue, &f)
		}
	}
}

func TagValueAndFlagsByTagString(tagString string) (string, map[string]bool) {
	valueAndFlags := strings.Split(tagString, ",")
	v := valueAndFlags[0]
	tagFlags := map[string]bool{}
	if len(valueAndFlags) > 1 {
		for _, flag := range valueAndFlags[1:] {
			tagFlags[flag] = true
		}
	}
	return v, tagFlags
}

func superWrite(w io.Writer, writeTo func(w io.Writer) error, contentType string) (string, error) {
	if rw, ok := w.(interface{ Header() http.Header }); ok {
		rw.Header().Set(httpx.HeaderContentType, contentType)
	}
	return contentType, writeTo(w)
}
