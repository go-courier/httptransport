package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/enumeration"
	eg "github.com/go-courier/enumeration/generator"
	"github.com/go-courier/httptransport/openapi/generator"
	"github.com/go-courier/oas"
	"github.com/go-courier/packagesx"
)

func NewTypeGenerator(serviceName string, file *codegen.File) *TypeGenerator {
	return &TypeGenerator{
		ServiceName: serviceName,
		File:        file,
		Enums:       map[string][]enumeration.EnumOption{},
	}
}

type TypeGenerator struct {
	ServiceName string
	File        *codegen.File
	Enums       map[string][]enumeration.EnumOption
}

func (g *TypeGenerator) Scan(openapi *oas.OpenAPI) {
	ids := make([]string, 0)
	for id := range openapi.Components.Schemas {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		s := openapi.Components.Schemas[id]

		typ, ok := g.Type(s)

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

		e := eg.NewEnum(enumName, options)
		e.WriteToFile(g.File)
	}

}

func (g *TypeGenerator) Type(schema *oas.Schema) (codegen.SnippetType, bool) {
	tpe, alias := g.TypeIndirect(schema)
	if schema != nil && schema.Extensions[generator.XGoStarLevel] != nil {
		level := int(schema.Extensions[generator.XGoStarLevel].(float64))
		for level > 0 {
			tpe = codegen.Star(tpe)
			level--
		}
	}
	return tpe, alias
}

func (g *TypeGenerator) TypeIndirect(schema *oas.Schema) (codegen.SnippetType, bool) {
	if schema == nil {
		return codegen.Interface(), false
	}

	if schema.Refer != nil {
		return codegen.Type(schema.Refer.(*oas.ComponentRefer).ID), true
	}

	if schema.Extensions[generator.XGoVendorType] != nil {
		pkgImportPath, expose := packagesx.GetPkgImportPathAndExpose(schema.Extensions[generator.XGoVendorType].(string))
		return codegen.Type(g.File.Use(pkgImportPath, expose)), true
	}

	if schema.Enum != nil {
		if enumOptionsValues, ok := schema.Extensions[generator.XEnumOptions]; ok {
			name := codegen.UpperCamelCase(g.ServiceName) + schema.Extensions[generator.XID].(string)

			enumOptions := make([]enumeration.EnumOption, 0)
			buf := bytes.NewBuffer(nil)
			_ = json.NewEncoder(buf).Encode(enumOptionsValues)
			_ = json.NewDecoder(buf).Decode(&enumOptions)
			g.Enums[name] = enumOptions

			return codegen.Type(name), true
		}
	}

	if len(schema.AllOf) > 0 {
		if schema.AllOf[len(schema.AllOf)-1].Type == oas.TypeObject {
			return codegen.Struct(g.FieldsFrom(schema)...), false
		}
		return g.TypeIndirect(mayComposedAllOf(schema))
	}

	if schema.Type == oas.TypeObject {
		if schema.AdditionalProperties != nil {
			tpe, _ := g.Type(schema.AdditionalProperties.Schema)
			keyTyp := codegen.SnippetType(codegen.String)
			if schema.PropertyNames != nil {
				keyTyp, _ = g.Type(schema.PropertyNames)
			}
			return codegen.Map(keyTyp, tpe), false
		}
		return codegen.Struct(g.FieldsFrom(schema)...), false
	}

	if schema.Type == oas.TypeArray {
		if schema.Items != nil {
			tpe, _ := g.Type(schema.Items)
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

func (g *TypeGenerator) FieldsFrom(schema *oas.Schema) (fields []*codegen.SnippetField) {
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
		fields = append(fields, g.FieldOf(name, mayComposedAllOf(finalSchema.Properties[name]), requiredFieldSet))
	}
	return
}

func (g *TypeGenerator) FieldOf(name string, propSchema *oas.Schema, requiredFields map[string]bool) *codegen.SnippetField {
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

	typ, _ := g.Type(propSchema)

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

func writeEnumDefines(file *codegen.File, name string, options []enumeration.EnumOption) {
	_, _ = file.WriteString(`// openapi:enum
`)
	file.WriteBlock(
		codegen.DeclType(codegen.Var(codegen.Int, name)),
	)

	file.WriteString(`
const (
`)

	file.WriteString(codegen.UpperSnakeCase(name) + `_UNKNOWN ` + name + ` = iota
	`)

	sort.Slice(options, func(i, j int) bool {
		return options[i].ConstValue < options[j].ConstValue
	})

	index := 1
	for _, item := range options {
		v := item.ConstValue
		if v > index {
			file.WriteString(`)

	const (
	`)
			file.WriteString(codegen.UpperSnakeCase(name) + `__` + item.Value + fmt.Sprintf(" %s = iota + %d", name, v) + ` // ` + item.Label + `
	`)
			index = v + 1
			continue
		}
		index++
		file.WriteString(codegen.UpperSnakeCase(name) + `__` + item.Value + ` // ` + item.Label + `
	`)
	}

	file.WriteString(`)
`)
}
