package department

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/godispatcher/dispatcher/model"
)

type MockTransactionItem struct {
	name string
}

func (m *MockTransactionItem) GetName() string { return m.name }
func (m *MockTransactionItem) GetTransaction() model.ServerInterface {
	return &MockServer{}
}

type MockServer struct{}

func (s *MockServer) Init(doc model.Document) model.Document {
	// Return the doc as is so we can inspect it in the response
	return doc
}
func (s *MockServer) GetRequest() any                { return struct{}{} }
func (s *MockServer) GetResponse() any               { return struct{}{} }
func (s *MockServer) GetOptions() model.ServerOption { return model.ServerOption{} }

func TestRegisterMainFunc_PopulatesFields(t *testing.T) {
	// Setup
	item := &MockTransactionItem{name: "test-trans"}
	DispatcherHolder.Add("test-dept", item)

	body := `{"department": "test-dept", "transaction": "test-trans", "form": {"key": "value"}}`
	req := httptest.NewRequest("POST", "/test-dept/test-trans/extra/param?param1=val1&param2=val2", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom-Header", "custom-value")

	rr := httptest.NewRecorder()

	// Execute
	RegisterMainFunc(rr, req)

	// Verify
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status OK, got %v. Body: %s", rr.Code, rr.Body.String())
	}

	var responseDoc model.Document
	err := json.Unmarshal(rr.Body.Bytes(), &responseDoc)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check Header
	if responseDoc.Header.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected X-Custom-Header to be 'custom-value', got '%v'", responseDoc.Header.Get("X-Custom-Header"))
	}

	// Check QueryParams
	if responseDoc.QueryParams.Get("param1") != "val1" {
		t.Errorf("Expected param1 to be 'val1', got '%v'", responseDoc.QueryParams.Get("param1"))
	}

	// Check URLParams
	if responseDoc.URLParams["department"] != "test-dept" {
		t.Errorf("Expected URLParams department to be 'test-dept', got '%v'", responseDoc.URLParams["department"])
	}
	if responseDoc.URLParams["transaction"] != "test-trans" {
		t.Errorf("Expected URLParams transaction to be 'test-trans', got '%v'", responseDoc.URLParams["transaction"])
	}
	if responseDoc.URLParams["segment_2"] != "extra" {
		t.Errorf("Expected URLParams segment_2 to be 'extra', got '%v'", responseDoc.URLParams["segment_2"])
	}
	if responseDoc.URLParams["segment_3"] != "param" {
		t.Errorf("Expected URLParams segment_3 to be 'param', got '%v'", responseDoc.URLParams["segment_3"])
	}
}
