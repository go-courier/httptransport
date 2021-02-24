package generator

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"sort"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/enumeration/scanner"
	"github.com/go-courier/httptransport/openapi/generator"
	"github.com/go-courier/oas"
	"github.com/go-courier/packagesx"
)

func NewTypeGenerator(serviceName string, file *codegen.File) *TypeGenerator {
	return &TypeGenerator{
		ServiceName: serviceName,
		File:        file,
		Enums:       map[string]scanner.Options{},
	}
}

type TypeGenerator struct {
	ServiceName string
	File        *codegen.File
	Enums       map[string]scanner.Options
}

func (g *TypeGenerator) Scan(ctx context.Context, openapi *oas.OpenAPI) {
	ids := make([]string, 0)
	for id := range openapi.Components.Schemas {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		s := openapi.Components.Schemas[id]

		typ, ok := g.Type(ctx, s)

		if ok {
			g.File.WriteBlock(
				codegen.DeclType(
					codegen.Var(typ, id).AsAlias(),
				),
			)
			continue
		}

		g.File.WriteBlock(
			codegen.DeclType(
				codegen.Var(typ, id),
			),
		)
	}

	enumNames := make([]string, 0)
	for id := range g.Enums {
		enumNames = append(enumNames, id)
	}
	sort.Strings(enumNames)

	for _, enumName := range enumNames {
		options := g.Enums[enumName]

		writeEnumDefines(g.File, enumName, options)
	}
}

func (g *TypeGenerator) Type(ctx context.Context, schema *oas.Schema) (codegen.SnippetType, bool) {
	tpe, alias := g.TypeIndirect(ctx, schema)
	if schema != nil && schema.Extensions[generator.XGoStarLevel] != nil {
		level := int(schema.Extensions[generator.XGoStarLevel].(float64))
		for level > 0 {
			tpe = codegen.Star(tpe)
			level--
		}
	}
	return tpe, alias
}

func paths(path string) []string {
	paths := make([]string, 0)

	d := path

	for {
		paths = append(paths, d)

		if !strings.Contains(d, "/") {
			break
		}

		d = filepath.Join(d, "../")
	}

	return paths
}

func (g *TypeGenerator) TypeIndirect(ctx context.Context, schema *oas.Schema) (codegen.SnippetType, bool) {
	if schema == nil {
		return codegen.Interface(), false
	}

	if schema.Refer != nil {
		return codegen.Type(schema.Refer.(*oas.ComponentRefer).ID), true
	}

	if schema.Extensions[generator.XGoVendorType] != nil {
		pkgImportPath, expose := packagesx.GetPkgImportPathAndExpose(schema.Extensions[generator.XGoVendorType].(string))

		vendorImports := VendorImportsFromContext(ctx)

		if len(vendorImports) > 0 {
			for _, p := range paths(pkgImportPath) {
				if _, ok := vendorImports[p]; ok {
					return codegen.Type(g.File.Use(pkgImportPath, expose)), true
				}
			}
		} else {
			return codegen.Type(g.File.Use(pkgImportPath, expose)), true
		}
	}

	if schema.Enum != nil {
		name := codegen.UpperCamelCase(g.ServiceName)

		if id, ok := schema.Extensions[generator.XID].(string); ok {
			name = name + id

			enumOptions := scanner.Options{}

			enumLabels := make([]string, len(schema.Enum))

			if xEnumLabels, ok := schema.Extensions[generator.XEnumLabels]; ok {
				if labels, ok := xEnumLabels.([]interface{}); ok {
					for i, l := range labels {
						if v, ok := l.(string); ok {
							enumLabels[i] = v
						}
					}
				}
			}

			if options, ok := schema.Extensions[generator.XEnumOptions]; ok {
				if list, ok := options.([]interface{}); ok {
					for i, l := range list {
						if opt, ok := l.(map[string]interface{}); ok {
							if s, ok := opt["label"]; ok {
								if v, ok := s.(string); ok {
									enumLabels[i] = v
								}
							}
						}
					}
				}
			}

			for i, e := range schema.Enum {
				o := scanner.Option{}

				switch v := e.(type) {
				case float64:
					o.Float = &v
				case int64:
					o.Int = &v
				case string:
					o.Str = &v
				}

				if len(enumLabels) > i {
					o.Label = enumLabels[i]
				}

				enumOptions = append(enumOptions, o)
			}

			g.Enums[name] = enumOptions

			return codegen.Type(name), true
		}
	}

	if len(schema.AllOf) > 0 {
		if schema.AllOf[len(schema.AllOf)-1].Type == oas.TypeObject {
			return codegen.Struct(g.FieldsFrom(ctx, schema)...), false
		}
		return g.TypeIndirect(ctx, mayComposedAllOf(schema))
	}

	if schema.Type == oas.TypeObject {
		if schema.AdditionalProperties != nil {
			tpe, _ := g.Type(ctx, schema.AdditionalProperties.Schema)
			keyTyp := codegen.SnippetType(codegen.String)
			if schema.PropertyNames != nil {
				keyTyp, _ = g.Type(ctx, schema.PropertyNames)
			}
			return codegen.Map(keyTyp, tpe), false
		}
		return codegen.Struct(g.FieldsFrom(ctx, schema)...), false
	}

	if schema.Type == oas.TypeArray {
		if schema.Items != nil {
			tpe, _ := g.Type(ctx, schema.Items)
			if schema.MaxItems != nil && schema.MinItems != nil && *schema.MaxItems == *schema.MinItems {
				return codegen.Array(tpe, int(*schema.MinItems)), false
			}
			return codegen.Slice(tpe), false
		}
	}

	return g.BasicType(string(schema.Type), schema.Format), false
}

func (g *TypeGenerator) BasicType(schemaType string, format string) codegen.SnippetType {
	switch format {
	case "binary":
		return codegen.Type(g.File.Use("mime/multipart", "FileHeader"))
	case "byte", "int", "int8", "int16", "int32", "int64", "rune", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64":
		return codegen.BuiltInType(format)
	case "float":
		return codegen.Float32
	case "double":
		return codegen.Float64
	default:
		switch schemaType {
		case "null":
			// type
			return nil
		case "integer":
			return codegen.Int
		case "number":
			return codegen.Float64
		case "boolean":
			return codegen.Bool
		default:
			return codegen.String
		}
	}
}

func (g *TypeGenerator) FieldsFrom(ctx context.Context, schema *oas.Schema) (fields []*codegen.SnippetField) {
	finalSchema := &oas.Schema{}

	if schema.AllOf != nil {
		for _, s := range schema.AllOf {
			if s.Refer != nil {
				fields = append(fields, codegen.Var(codegen.Type(s.Refer.(*oas.ComponentRefer).ID)))
			} else {
				finalSchema = s
				break
			}
		}
	} else {
		finalSchema = schema
	}

	if finalSchema.Properties == nil {
		return
	}

	names := make([]string, 0)
	for fieldName := range finalSchema.Properties {
		names = append(names, fieldName)
	}
	sort.Strings(names)

	requiredFieldSet := map[string]bool{}

	for _, name := range finalSchema.Required {
		requiredFieldSet[name] = true
	}

	for _, name := range names {
		fields = append(fields, g.FieldOf(ctx, name, mayComposedAllOf(finalSchema.Properties[name]), requiredFieldSet))
	}
	return
}

func (g *TypeGenerator) FieldOf(ctx context.Context, name string, propSchema *oas.Schema, requiredFields map[string]bool) *codegen.SnippetField {
	isRequired := requiredFields[name]

	if len(propSchema.AllOf) == 2 && propSchema.AllOf[1].Type != oas.TypeObject {
		propSchema = &oas.Schema{
			Reference:      propSchema.AllOf[0].Reference,
			SchemaObject:   propSchema.AllOf[1].SchemaObject,
			SpecExtensions: propSchema.AllOf[1].SpecExtensions,
		}
	}

	fieldName := codegen.UpperCamelCase(name)
	if propSchema.Extensions[generator.XGoFieldName] != nil {
		fieldName = propSchema.Extensions[generator.XGoFieldName].(string)
	}

	typ, _ := g.Type(ctx, propSchema)

	field := codegen.Var(typ, fieldName).WithComments(mayPrefixDeprecated(propSchema.Description, propSchema.Deprecated)...)

	tags := map[string][]string{}

	appendTag := func(key string, valuesOrFlags ...string) {
		tags[key] = append(tags[key], valuesOrFlags...)
	}

	appendNamedTag := func(key string, value string) {
		appendTag(key, value)
		if !isRequired && !strings.Contains(value, "omitempty") {
			appendTag(key, "omitempty")
		}
	}

	if propSchema.Extensions[generator.XTagJSON] != nil {
		appendNamedTag("json", propSchema.Extensions[generator.XTagJSON].(string))
	}

	if propSchema.Extensions[generator.XTagName] != nil {
		appendNamedTag("name", propSchema.Extensions[generator.XTagName].(string))
	}

	if propSchema.Extensions[generator.XTagXML] != nil {
		appendNamedTag("xml", propSchema.Extensions[generator.XTagXML].(string))
	}

	if propSchema.Extensions[generator.XTagMime] != nil {
		appendTag("mime", propSchema.Extensions[generator.XTagMime].(string))
	}

	if propSchema.Extensions[generator.XTagValidate] != nil {
		appendTag("validate", propSchema.Extensions[generator.XTagValidate].(string))
	}

	if propSchema.Default != nil {
		appendTag("default", fmt.Sprintf("%v", propSchema.Default))
	}

	field = field.WithTags(tags)
	return field
}

func mayComposedAllOf(schema *oas.Schema) *oas.Schema {
	// for named field
	if schema.AllOf != nil && len(schema.AllOf) == 2 && schema.AllOf[len(schema.AllOf)-1].Type != oas.TypeObject {
		nextSchema := &oas.Schema{
			Reference:    schema.AllOf[0].Reference,
			SchemaObject: schema.AllOf[1].SchemaObject,
		}

		for k, v := range schema.AllOf[1].SpecExtensions.Extensions {
			nextSchema.AddExtension(k, v)
		}

		for k, v := range schema.SpecExtensions.Extensions {
			nextSchema.AddExtension(k, v)
		}

		return nextSchema
	}

	return schema
}

func writeEnumDefines(file *codegen.File, name string, options scanner.Options) {
	if len(options) == 0 {
		return
	}

	switch options[0].Value().(type) {
	case int64:
		file.WriteBlock(
			codegen.DeclType(codegen.Var(codegen.Int64, name)),
		)
	case float64:
		file.WriteBlock(
			codegen.DeclType(codegen.Var(codegen.Float64, name)),
		)
	case string:
		file.WriteBlock(
			codegen.DeclType(codegen.Var(codegen.String, name)),
		)
	}

	file.WriteString(`
const (
`)

	sort.Sort(options)

	for _, item := range options {
		v := item.Value()
		value := v

		switch n := v.(type) {
		case string:
			value = strconv.Quote(n)
		}

		_, _ = fmt.Fprintf(file, `%s__%v %s = %v // %s
`, codegen.UpperSnakeCase(name), v, name, value, item.Label)
	}

	file.WriteString(`)
`)
}
