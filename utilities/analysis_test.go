package utilities

import (
	"errors"
	"reflect"
	"testing"
)

type embeddedAnalysisFields struct {
	Embedded string `json:"embedded"`
}

type analysisDocument struct {
	embeddedAnalysisFields
	Untagged string
	Ignored  string         `json:"-"`
	Optional *int           `json:"optional,omitempty"`
	Bytes    []byte         `json:"bytes"`
	Array    [2]bool        `json:"array"`
	Lookup   map[string]int `json:"lookup"`
	Anything interface{}    `json:"anything"`
	hidden   string
}

type documentedRequest struct {
	Query   string            `json:"query" require:"true" is_empty:"false" example:"laptop"`
	Page    int               `json:"page" example:"2"`
	Enabled bool              `json:"enabled" is_empty:"false"`
	Tags    []string          `json:"tags" is_empty:"false"`
	Labels  map[string]string `json:"labels" is_empty:"false"`
}

type recursiveAnalysisNode struct {
	Children []*recursiveAnalysisNode             `json:"children"`
	Index    map[string]*recursiveAnalysisNode    `json:"index"`
	Other    *recursiveAnalysisNodeFromOtherScope `json:"other"`
}

type recursiveAnalysisNodeFromOtherScope struct {
	Value string `json:"value"`
}

type objectJSONMarshaler struct{}

func (objectJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`{"name":"example","enabled":true}`), nil
}

type nullJSONMarshaler struct{}

func (nullJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

type failingJSONMarshaler struct{}

func (failingJSONMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal failed")
}

func TestAnalyzeJSONSchemaStruct(t *testing.T) {
	schema, err := AnalyzeJSONSchema(reflect.TypeOf(analysisDocument{}))
	if err != nil {
		t.Fatal(err)
	}

	if schema.Type != "object" {
		t.Fatalf("type = %q, want object", schema.Type)
	}
	wantFields := []string{"embedded", "Untagged", "optional", "bytes", "array", "lookup", "anything"}
	for _, field := range wantFields {
		if _, ok := schema.Properties[field]; !ok {
			t.Errorf("missing property %q", field)
		}
	}
	for _, field := range []string{"Ignored", "-", "hidden", ""} {
		if _, ok := schema.Properties[field]; ok {
			t.Errorf("unexpected property %q", field)
		}
	}

	assertSchemaType(t, schema.Properties["bytes"], "string")
	assertSchemaType(t, schema.Properties["array"], "array")
	assertSchemaType(t, schema.Properties["array"].Items, "boolean")
	assertSchemaType(t, schema.Properties["lookup"], "object")
	assertSchemaType(t, schema.Properties["lookup"].AdditionalProperties, "integer")
	assertSchemaType(t, schema.Properties["anything"], "any")
	if !schema.Properties["optional"].Nullable {
		t.Error("pointer property should be nullable")
	}
}

func TestAnalyzeJSONSchemaIncludesValidationRules(t *testing.T) {
	schema, err := AnalyzeJSONSchema(reflect.TypeOf(documentedRequest{}))
	if err != nil {
		t.Fatal(err)
	}

	if len(schema.Required) != 1 || schema.Required[0] != "query" {
		t.Fatalf("required = %#v, want [query]", schema.Required)
	}
	if query := schema.Properties["query"]; query.MinLength == nil || *query.MinLength != 1 || query.AllowEmpty == nil || *query.AllowEmpty {
		t.Fatalf("query constraints = %#v", query)
	}
	if tags := schema.Properties["tags"]; tags.MinItems == nil || *tags.MinItems != 1 {
		t.Fatalf("tags constraints = %#v", tags)
	}
	if labels := schema.Properties["labels"]; labels.MinProperties == nil || *labels.MinProperties != 1 {
		t.Fatalf("labels constraints = %#v", labels)
	}
	if enabled := schema.Properties["enabled"]; enabled.AllowEmpty == nil || *enabled.AllowEmpty {
		t.Fatalf("enabled constraints = %#v", enabled)
	}
}

func TestGenerateJSONExampleUsesTagsAndValidTypeDefaults(t *testing.T) {
	example, err := GenerateJSONExample(reflect.TypeOf(documentedRequest{}))
	if err != nil {
		t.Fatal(err)
	}

	object, ok := example.(map[string]interface{})
	if !ok {
		t.Fatalf("example type = %T, want map", example)
	}
	if object["query"] != "laptop" || object["page"] != float64(2) {
		t.Fatalf("tagged examples = %#v", object)
	}
	if object["enabled"] != true {
		t.Fatalf("enabled = %#v, want true", object["enabled"])
	}
	if tags, ok := object["tags"].([]interface{}); !ok || len(tags) != 1 || tags[0] != "string" {
		t.Fatalf("tags = %#v, want one example item", object["tags"])
	}
}

func TestGenerateJSONExampleRejectsInvalidExampleTag(t *testing.T) {
	type request struct {
		Page int `json:"page" example:"not-a-number"`
	}

	if _, err := GenerateJSONExample(reflect.TypeOf(request{})); err == nil {
		t.Fatal("expected invalid example tag error")
	}
}

func TestGenerateJSONExampleHandlesRecursiveTypes(t *testing.T) {
	example, err := GenerateJSONExample(reflect.TypeOf(recursiveAnalysisNode{}))
	if err != nil {
		t.Fatal(err)
	}
	if example == nil {
		t.Fatal("expected recursive object example")
	}
}

func TestAnalyzeJSONSchemaRecursiveCollections(t *testing.T) {
	schema, err := AnalyzeJSONSchema(reflect.TypeOf(recursiveAnalysisNode{}))
	if err != nil {
		t.Fatal(err)
	}

	children := schema.Properties["children"]
	if children.Items == nil || children.Items.Ref == "" {
		t.Fatalf("recursive slice item = %#v, want reference", children.Items)
	}
	index := schema.Properties["index"]
	if index.AdditionalProperties == nil || index.AdditionalProperties.Ref == "" {
		t.Fatalf("recursive map value = %#v, want reference", index.AdditionalProperties)
	}
	other := schema.Properties["other"]
	if other.Ref != "" || other.Type != "object" {
		t.Fatalf("different named type treated as recursive: %#v", other)
	}
}

func TestAnalyzeJSONSchemaCustomMarshaler(t *testing.T) {
	schema, err := AnalyzeJSONSchema(reflect.TypeOf(objectJSONMarshaler{}))
	if err != nil {
		t.Fatal(err)
	}
	assertSchemaType(t, &schema, "object")
	assertSchemaType(t, schema.Properties["name"], "string")
	assertSchemaType(t, schema.Properties["enabled"], "boolean")
	pointerSchema, err := AnalyzeJSONSchema(reflect.TypeOf((*objectJSONMarshaler)(nil)))
	if err != nil {
		t.Fatal(err)
	}
	if !pointerSchema.Nullable {
		t.Error("pointer custom marshaler should be nullable")
	}

	nullSchema, err := AnalyzeJSONSchema(reflect.TypeOf(nullJSONMarshaler{}))
	if err != nil {
		t.Fatal(err)
	}
	assertSchemaType(t, &nullSchema, "null")

	if _, err := AnalyzeJSONSchema(reflect.TypeOf(failingJSONMarshaler{})); err == nil {
		t.Fatal("expected custom marshaling error")
	}
}

func TestMarshalJSONAnalysisUsesStandardTypeNames(t *testing.T) {
	tests := map[string]string{
		`"value"`: "string",
		`1`:       "number",
		`true`:    "boolean",
		`[]`:      "array",
		`{}`:      "object",
		`null`:    "null",
		`invalid`: "invalid",
	}
	for input, want := range tests {
		if got := MarshalJSONAnalysis([]byte(input)); got != want {
			t.Errorf("MarshalJSONAnalysis(%q) = %q, want %q", input, got, want)
		}
	}
}

func assertSchemaType(t *testing.T, schema *Schema, want string) {
	t.Helper()
	if schema == nil || schema.Type != want {
		t.Fatalf("schema = %#v, want type %q", schema, want)
	}
}
