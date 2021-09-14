package transformers

import (
	"context"
	"go/ast"
	"reflect"
	"strconv"
	"strings"
	"sync"

	contextx "github.com/go-courier/x/context"
	reflectx "github.com/go-courier/x/reflect"
	typesx "github.com/go-courier/x/types"
)

type Parameter struct {
	In    string
	Name  string
	Field typesx.StructField
	Type  typesx.Type
	Tags  map[string]Tag
	Loc   []int
}

func (p *Parameter) FieldValue(structReflectValue reflect.Value) reflect.Value {
	structReflectValue = reflectx.Indirect(structReflectValue)

	n := len(p.Loc)

	fieldValue := structReflectValue

	for i := 0; i < n; i++ {
		loc := p.Loc[i]
		fieldValue = fieldValue.Field(loc)

		// last loc should keep ptr value
		if i < n-1 {
			for fieldValue.Kind() == reflect.Ptr {
				// notice the ptr struct ensure only for Ptr Anonymous Field
				if fieldValue.IsNil() {
					fieldValue.Set(reflectx.New(fieldValue.Type()))
				}
				fieldValue = fieldValue.Elem()
			}
		}
	}

	return fieldValue
}

type Tag string

func (t Tag) Name() string {
	s := string(t)

	if i := strings.Index(s, ","); i >= 0 {
		if i == 0 {
			return ""
		}
		return s[0:i]
	}

	return s
}

func (t Tag) HasFlag(flag string) bool {
	if i := strings.Index(string(t), flag); i > 0 {
		return true
	}
	return false
}

type ParameterValue struct {
	Parameter
	Value reflect.Value
}

type GroupedParameters = map[string][]Parameter

type contextKeyGroupedParametersSet struct{}

var defaultGroupedParametersSet = &sync.Map{}

func GroupedParametersSetFromContext(ctx context.Context) *sync.Map {
	if m, ok := ctx.Value(contextKeyGroupedParametersSet{}).(*sync.Map); ok {
		return m
	}
	return defaultGroupedParametersSet
}

func WithGroupedParametersSet(ctx context.Context, m *sync.Map) context.Context {
	return contextx.WithValue(ctx, contextKeyGroupedParametersSet{}, m)
}

func EachParameter(ctx context.Context, tpe typesx.Type, each func(p *Parameter) bool) {
	var walk func(tpe typesx.Type, parents ...int)

	walk = func(tpe typesx.Type, parents ...int) {
		for i := 0; i < tpe.NumField(); i++ {
			f := tpe.Field(i)

			if !ast.IsExported(f.Name()) {
				continue
			}

			loc := append(parents, i)

			tags := ParseTags(string(f.Tag()))

			tagIn, hasIn := tags["in"]

			displayName := f.Name()

			tagName, hasName := tags["name"]
			if hasName {
				if name := tagName.Name(); name == "-" {
					// skip name:"-"
					continue
				} else {
					if name != "" {
						displayName = name
					}
				}
			}

			if f.Anonymous() && (!hasIn && !hasName) {
				fieldType := f.Type()

				_, ok := typesx.EncodingTextMarshalerTypeReplacer(fieldType)

				if !ok {
					for fieldType.Kind() == reflect.Ptr {
						fieldType = fieldType.Elem()
					}

					if fieldType.Kind() == reflect.Struct {
						walk(fieldType, loc...)
						continue
					}
				}
			}

			p := &Parameter{}
			p.Field = f
			p.Type = f.Type()
			p.Tags = tags
			p.In = tagIn.Name()
			p.Name = displayName
			p.Loc = append([]int{}, loc...)

			if !each(p) {
				break
			}
		}
	}

	walk(tpe)
}

func CollectGroupedParameters(ctx context.Context, tpe typesx.Type) GroupedParameters {
	if tpe.Kind() != reflect.Struct {
		return nil
	}

	m := GroupedParametersSetFromContext(ctx)

	if tp, ok := m.Load(tpe); ok {
		return tp.(GroupedParameters)
	}

	gp := GroupedParameters{}

	defer func() {
		m.Store(tpe, gp)
	}()

	EachParameter(ctx, tpe, func(p *Parameter) bool {
		gp[p.In] = append(gp[p.In], *p)
		return true
	})

	return gp
}

func ParseTags(tag string) map[string]Tag {
	tagFlags := map[string]Tag{}

	for tag != "" {
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := tag[:i+1]
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			break
		}
		tagFlags[name] = Tag(value)
	}

	return tagFlags
}
