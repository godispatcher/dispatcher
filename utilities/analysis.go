package utilities

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type StructVariable map[string]interface{}
type SliceVariable []interface{}

type Schema struct {
	Type                 string             `json:"type,omitempty" yaml:"type,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Nullable             bool               `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Ref                  string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

type schemaAnalyzer struct {
	active map[reflect.Type]bool
}

type jsonField struct {
	typeOf reflect.Type
	depth  int
	tagged bool
}

var (
	jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

// AnalyzeJSONSchema creates a stable schema from a Go type without depending on
// the runtime contents of a response value.
func AnalyzeJSONSchema(typeOf reflect.Type) (Schema, error) {
	analyzer := schemaAnalyzer{active: make(map[reflect.Type]bool)}
	return analyzer.analyze(typeOf)
}

// Analysis is kept for compatibility with existing callers. New callers that
// need error details should use AnalyzeJSONSchema.
func Analysis(variable interface{}, _ *[]string) interface{} {
	schema, err := AnalyzeJSONSchema(reflect.TypeOf(variable))
	if err != nil {
		return Schema{Type: "invalid"}
	}
	return schema
}

func (analyzer *schemaAnalyzer) analyze(typeOf reflect.Type) (Schema, error) {
	if typeOf == nil {
		return Schema{Type: "any"}, nil
	}
	if hasCustomJSONMarshaling(typeOf) {
		schema, err := analyzeMarshaledType(typeOf)
		if typeOf.Kind() == reflect.Ptr {
			schema.Nullable = true
		}
		return schema, err
	}

	switch typeOf.Kind() {
	case reflect.Ptr:
		schema, err := analyzer.analyze(typeOf.Elem())
		schema.Nullable = true
		return schema, err
	case reflect.Interface:
		return Schema{Type: "any"}, nil
	case reflect.Struct:
		if analyzer.active[typeOf] {
			return Schema{Ref: typeReference(typeOf)}, nil
		}
		analyzer.active[typeOf] = true
		defer delete(analyzer.active, typeOf)

		fields := make(map[string][]jsonField)
		collectJSONFields(typeOf, 0, make(map[reflect.Type]bool), fields)
		properties := make(map[string]*Schema)
		for name, candidates := range fields {
			field, ok := dominantJSONField(candidates)
			if !ok {
				continue
			}
			fieldSchema, err := analyzer.analyze(field.typeOf)
			if err != nil {
				return Schema{}, fmt.Errorf("analyze field %s.%s: %w", typeOf, name, err)
			}
			properties[name] = &fieldSchema
		}
		return Schema{Type: "object", Properties: properties}, nil
	case reflect.Slice:
		if typeOf.Elem().Kind() == reflect.Uint8 {
			return Schema{Type: "string"}, nil
		}
		fallthrough
	case reflect.Array:
		item, err := analyzer.analyze(typeOf.Elem())
		return Schema{Type: "array", Items: &item}, err
	case reflect.Map:
		if !validJSONMapKey(typeOf.Key()) {
			return Schema{}, fmt.Errorf("unsupported JSON map key type %s", typeOf.Key())
		}
		value, err := analyzer.analyze(typeOf.Elem())
		return Schema{Type: "object", AdditionalProperties: &value}, err
	case reflect.String:
		return Schema{Type: "string"}, nil
	case reflect.Bool:
		return Schema{Type: "boolean"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return Schema{Type: "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return Schema{Type: "number"}, nil
	default:
		return Schema{Type: typeOf.Kind().String()}, nil
	}
}

func collectJSONFields(typeOf reflect.Type, depth int, visiting map[reflect.Type]bool, fields map[string][]jsonField) {
	if visiting[typeOf] {
		return
	}
	visiting[typeOf] = true
	defer delete(visiting, typeOf)

	for index := 0; index < typeOf.NumField(); index++ {
		field := typeOf.Field(index)
		tag, hasTag := field.Tag.Lookup("json")
		name := strings.Split(tag, ",")[0]
		if hasTag && name == "-" {
			continue
		}
		embeddedType := field.Type
		if embeddedType.Kind() == reflect.Ptr {
			embeddedType = embeddedType.Elem()
		}
		if field.Anonymous && name == "" && embeddedType.Kind() == reflect.Struct {
			collectJSONFields(embeddedType, depth+1, visiting, fields)
			continue
		}
		if field.PkgPath != "" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		fields[name] = append(fields[name], jsonField{typeOf: field.Type, depth: depth, tagged: hasTag && strings.Split(tag, ",")[0] != ""})
	}
}

func dominantJSONField(fields []jsonField) (jsonField, bool) {
	minimumDepth := fields[0].depth
	for _, field := range fields[1:] {
		if field.depth < minimumDepth {
			minimumDepth = field.depth
		}
	}
	var candidates []jsonField
	for _, field := range fields {
		if field.depth == minimumDepth {
			candidates = append(candidates, field)
		}
	}
	if len(candidates) == 1 {
		return candidates[0], true
	}
	var tagged []jsonField
	for _, field := range candidates {
		if field.tagged {
			tagged = append(tagged, field)
		}
	}
	if len(tagged) == 1 {
		return tagged[0], true
	}
	return jsonField{}, false
}

func hasCustomJSONMarshaling(typeOf reflect.Type) bool {
	if typeOf.Implements(jsonMarshalerType) || typeOf.Implements(textMarshalerType) {
		return true
	}
	return typeOf.Kind() != reflect.Ptr &&
		(reflect.PointerTo(typeOf).Implements(jsonMarshalerType) || reflect.PointerTo(typeOf).Implements(textMarshalerType))
}

func analyzeMarshaledType(typeOf reflect.Type) (Schema, error) {
	var value reflect.Value
	if typeOf.Kind() == reflect.Ptr {
		value = reflect.New(typeOf.Elem())
	} else {
		value = reflect.New(typeOf)
		if !value.Type().Implements(jsonMarshalerType) && !value.Type().Implements(textMarshalerType) {
			value = value.Elem()
		}
	}
	data, err := json.Marshal(value.Interface())
	if err != nil {
		return Schema{}, err
	}
	var decoded interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return Schema{}, err
	}
	return schemaFromJSONValue(decoded), nil
}

func schemaFromJSONValue(value interface{}) Schema {
	switch typed := value.(type) {
	case nil:
		return Schema{Type: "null"}
	case string:
		return Schema{Type: "string"}
	case bool:
		return Schema{Type: "boolean"}
	case float64:
		return Schema{Type: "number"}
	case []interface{}:
		item := Schema{Type: "any"}
		if len(typed) > 0 {
			item = schemaFromJSONValue(typed[0])
		}
		return Schema{Type: "array", Items: &item}
	case map[string]interface{}:
		properties := make(map[string]*Schema, len(typed))
		for name, item := range typed {
			schema := schemaFromJSONValue(item)
			properties[name] = &schema
		}
		return Schema{Type: "object", Properties: properties}
	default:
		return Schema{Type: "any"}
	}
}

func validJSONMapKey(typeOf reflect.Type) bool {
	if typeOf.Kind() == reflect.String || typeOf.Implements(textMarshalerType) ||
		(typeOf.Kind() != reflect.Ptr && reflect.PointerTo(typeOf).Implements(textMarshalerType)) {
		return true
	}
	return typeOf.Kind() >= reflect.Int && typeOf.Kind() <= reflect.Int64 ||
		typeOf.Kind() >= reflect.Uint && typeOf.Kind() <= reflect.Uintptr
}

func typeReference(typeOf reflect.Type) string {
	if typeOf.PkgPath() == "" {
		return typeOf.String()
	}
	return typeOf.PkgPath() + "." + typeOf.Name()
}

func MarshalJSONAnalysis(byteData []byte) string {
	var result interface{}
	if err := json.Unmarshal(byteData, &result); err != nil {
		return "invalid"
	}
	return schemaFromJSONValue(result).Type
}
