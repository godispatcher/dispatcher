package model

import (
	"strings"
	"testing"
)

type validationLine struct {
	SKU string `json:"sku" require:"true" isEmpty:"false"`
}

type validationReference struct {
	Code string `json:"code" require:"true"`
}

type validationRequest struct {
	Name       string                         `json:"name" require:"true" is_empty:"false"`
	Alias      string                         `json:",omitempty" is_empty:"false"`
	Untagged   string                         `require:"true"`
	Items      []validationLine               `json:"items" require:"true" is_empty:"false"`
	References map[string]validationReference `json:"references"`
	Ignored    string                         `json:"-" require:"true"`
}

func TestDocumentFormValidaterValidatesNestedValues(t *testing.T) {
	tests := []struct {
		name    string
		request string
		want    string
	}{
		{
			name:    "valid request",
			request: `{"name":"Ada","Untagged":"value","items":[{"sku":"A-1"}],"references":{"primary":{"code":"ref"}}}`,
		},
		{
			name:    "missing top-level field",
			request: `{"Untagged":"value","items":[{"sku":"A-1"}]}`,
			want:    "name",
		},
		{
			name:    "missing nested slice field",
			request: `{"name":"Ada","Untagged":"value","items":[{}]}`,
			want:    "items[0].sku",
		},
		{
			name:    "missing nested map field",
			request: `{"name":"Ada","Untagged":"value","items":[{"sku":"A-1"}],"references":{"primary":{}}}`,
			want:    `references["primary"].code`,
		},
		{
			name:    "wrong JSON type",
			request: `{"name":"Ada","Untagged":"value","items":"not-an-array"}`,
			want:    "cannot unmarshal",
		},
		{
			name:    "null for non-nullable field",
			request: `{"name":null,"Untagged":"value","items":[{"sku":"A-1"}]}`,
			want:    "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := DocumentFormValidater{Request: tt.request}
			err := validator.Validate(&validationRequest{})
			if tt.want == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestDocumentFormValidaterRejectsEmptyValues(t *testing.T) {
	type request struct {
		Text    string         `json:"text" is_empty:"false"`
		Count   int            `json:"count" is_empty:"false"`
		Enabled bool           `json:"enabled" is_empty:"false"`
		Values  []string       `json:"values" is_empty:"false"`
		Lookup  map[string]int `json:"lookup" is_empty:"false"`
		Pointer *string        `json:"pointer" is_empty:"false"`
	}

	tests := []struct {
		name    string
		request string
		want    string
	}{
		{name: "absent optional values", request: `{}`},
		{name: "blank string", request: `{"text":"  "}`, want: "text"},
		{name: "zero number", request: `{"count":0}`, want: "count"},
		{name: "false boolean", request: `{"enabled":false}`, want: "enabled"},
		{name: "empty slice", request: `{"values":[]}`, want: "values"},
		{name: "empty map", request: `{"lookup":{}}`, want: "lookup"},
		{name: "null pointer", request: `{"pointer":null}`, want: "pointer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := DocumentFormValidater{Request: tt.request}
			err := validator.Validate(request{})
			if tt.want == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestDocumentFormValidaterReturnsConfigurationAndInputErrors(t *testing.T) {
	tests := []struct {
		name    string
		request string
		target  any
		want    string
	}{
		{name: "malformed JSON", request: `{`, target: struct{}{}, want: "invalid request JSON"},
		{name: "non-object JSON", request: `[]`, target: struct{}{}, want: "cannot unmarshal"},
		{name: "null object", request: `null`, target: struct{}{}, want: "cannot be null"},
		{name: "nil target", request: `{}`, target: nil, want: "request type must be"},
		{name: "non-struct target", request: `{}`, target: "", want: "request type must be"},
		{
			name:    "invalid tag",
			request: `{}`,
			target: struct {
				Value string `json:"value" require:"invalid"`
			}{},
			want: "invalid require struct tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := DocumentFormValidater{Request: tt.request}
			err := validator.Validate(tt.target)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}
