package utilities

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type StructVariable map[string]interface{}
type SliceVariable []interface{}

type Schema struct {
	Type                 string             `json:"type,omitempty" yaml:"type,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required             []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Items                *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Nullable             bool               `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	MinLength            *int               `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MinItems             *int               `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MinProperties        *int               `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	AllowEmpty           *bool              `json:"x-allow-empty,omitempty" yaml:"x-allow-empty,omitempty"`
	Ref                  string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

type schemaAnalyzer struct {
	active map[reflect.Type]bool
}

type jsonField struct {
	field  reflect.StructField
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
		var required []string
		for name, candidates := range fields {
			field, ok := dominantJSONField(candidates)
			if !ok {
				continue
			}
			fieldSchema, err := analyzer.analyze(field.field.Type)
			if err != nil {
				return Schema{}, fmt.Errorf("analyze field %s.%s: %w", typeOf, name, err)
			}
			tagOptions, err := ParseTagToTransactionExchangeTag(string(field.field.Tag))
			if err != nil {
				return Schema{}, fmt.Errorf("analyze validation tags for %s.%s: %w", typeOf, name, err)
			}
			if tagOptions.Require != nil && *tagOptions.Require {
				required = append(required, name)
			}
			if tagOptions.IsEmpty != nil && !*tagOptions.IsEmpty {
				applyNotEmptyConstraints(&fieldSchema, field.field.Type)
			}
			properties[name] = &fieldSchema
		}
		sort.Strings(required)
		return Schema{Type: "object", Properties: properties, Required: required}, nil
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
		fields[name] = append(fields[name], jsonField{field: field, depth: depth, tagged: hasTag && strings.Split(tag, ",")[0] != ""})
	}
}

func applyNotEmptyConstraints(schema *Schema, typeOf reflect.Type) {
	allowEmpty := false
	schema.AllowEmpty = &allowEmpty
	for typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}
	one := 1
	switch typeOf.Kind() {
	case reflect.String:
		schema.MinLength = &one
	case reflect.Slice, reflect.Array:
		schema.MinItems = &one
	case reflect.Struct, reflect.Map:
		schema.MinProperties = &one
	}
}

// GenerateJSONExample creates a complete, directly serializable example for a
// Go request type. Explicit example tags take precedence over type defaults.
func GenerateJSONExample(typeOf reflect.Type) (interface{}, error) {
	return generateJSONExample(typeOf, make(map[reflect.Type]bool))
}

func generateJSONExample(typeOf reflect.Type, active map[reflect.Type]bool) (interface{}, error) {
	if typeOf == nil {
		return nil, nil
	}
	if hasCustomJSONMarshaling(typeOf) {
		return marshalExampleValue(typeOf)
	}
	for typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}

	switch typeOf.Kind() {
	case reflect.Struct:
		if active[typeOf] {
			return nil, nil
		}
		active[typeOf] = true
		defer delete(active, typeOf)

		fields := make(map[string][]jsonField)
		collectJSONFields(typeOf, 0, make(map[reflect.Type]bool), fields)
		result := make(map[string]interface{}, len(fields))
		for name, candidates := range fields {
			field, ok := dominantJSONField(candidates)
			if !ok {
				continue
			}
			value, exists, err := fieldExample(field.field, active)
			if err != nil {
				return nil, fmt.Errorf("generate example for %s.%s: %w", typeOf, name, err)
			}
			if exists {
				result[name] = value
			}
		}
		return result, nil
	case reflect.Slice:
		if typeOf.Elem().Kind() == reflect.Uint8 {
			return "string", nil
		}
		item, err := generateJSONExample(typeOf.Elem(), active)
		return []interface{}{item}, err
	case reflect.Array:
		item, err := generateJSONExample(typeOf.Elem(), active)
		return []interface{}{item}, err
	case reflect.Map:
		value, err := generateJSONExample(typeOf.Elem(), active)
		return map[string]interface{}{"key": value}, err
	case reflect.Interface:
		return map[string]interface{}{}, nil
	case reflect.String:
		return "string", nil
	case reflect.Bool:
		return true, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return float64(1), nil
	default:
		return nil, nil
	}
}

func fieldExample(field reflect.StructField, active map[reflect.Type]bool) (interface{}, bool, error) {
	if raw, ok := field.Tag.Lookup("example"); ok {
		value, err := parseExampleTag(raw, field.Type)
		return value, true, err
	}
	value, err := generateJSONExample(field.Type, active)
	return value, true, err
}

func parseExampleTag(raw string, typeOf reflect.Type) (interface{}, error) {
	value := reflect.New(typeOf)
	data := []byte(raw)
	baseType := typeOf
	for baseType.Kind() == reflect.Ptr {
		baseType = baseType.Elem()
	}
	if baseType.Kind() == reflect.String {
		data, _ = json.Marshal(raw)
	}
	if err := json.Unmarshal(data, value.Interface()); err != nil {
		return nil, fmt.Errorf("invalid example tag value %q for %s: %w", raw, typeOf, err)
	}
	encoded, err := json.Marshal(value.Elem().Interface())
	if err != nil {
		return nil, err
	}
	var decoded interface{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func marshalExampleValue(typeOf reflect.Type) (interface{}, error) {
	var value reflect.Value
	if typeOf.Kind() == reflect.Ptr {
		value = reflect.New(typeOf.Elem())
	} else {
		value = reflect.New(typeOf).Elem()
	}
	data, err := json.Marshal(value.Interface())
	if err != nil {
		return nil, err
	}
	var decoded interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
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
