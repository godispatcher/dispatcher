package utilities

import "testing"

func TestParseTagToTransactionExchangeTag(t *testing.T) {
	tag, err := ParseTagToTransactionExchangeTag(`json:"display_name,omitempty" require:"TRUE" isEmpty:"false"`)
	if err != nil {
		t.Fatal(err)
	}
	if tag.FieldRawname != "display_name" {
		t.Errorf("field name = %q, want display_name", tag.FieldRawname)
	}
	if tag.Require == nil || !*tag.Require {
		t.Errorf("require = %#v, want true", tag.Require)
	}
	if tag.IsEmpty == nil || *tag.IsEmpty {
		t.Errorf("isEmpty = %#v, want false", tag.IsEmpty)
	}
}

func TestParseTagToTransactionExchangeTagMalformedInput(t *testing.T) {
	if _, err := ParseTagToTransactionExchangeTag(`require:"not-a-bool"`); err == nil {
		t.Fatal("expected invalid boolean error")
	}

	if _, err := ParseTagToTransactionExchangeTag(`require`); err != nil {
		t.Fatalf("malformed struct tag should be ignored safely: %v", err)
	}
}
