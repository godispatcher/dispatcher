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
	// Return the doc as is, but we also want to check the RequestContext
	ctx := model.GetRequestContext()
	if ctx != nil {
		// We can't easily return ctx in model.Document anymore, so we might just log or use a global for testing
		// but since we want to verify fields, let's use the Output field of document for verification in this mock
		verification := make(map[string]interface{})
		verification["header"] = ctx.Header
		verification["query_params"] = ctx.QueryParams
		verification["url_params"] = ctx.URLParams
		doc.Output = verification
	}
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

	// Check verification data from Output
	output, ok := responseDoc.Output.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected output to be a map, got %T", responseDoc.Output)
	}

	header := output["header"].(map[string]interface{})
	if header["X-Custom-Header"].([]interface{})[0].(string) != "custom-value" {
		t.Errorf("Expected X-Custom-Header to be 'custom-value', got '%v'", header["X-Custom-Header"])
	}

	queryParams := output["query_params"].(map[string]interface{})
	if queryParams["param1"].([]interface{})[0].(string) != "val1" {
		t.Errorf("Expected param1 to be 'val1', got '%v'", queryParams["param1"])
	}

	urlParams := output["url_params"].(map[string]interface{})
	if urlParams["department"] != "test-dept" {
		t.Errorf("Expected URLParams department to be 'test-dept', got '%v'", urlParams["department"])
	}
	if urlParams["transaction"] != "test-trans" {
		t.Errorf("Expected URLParams transaction to be 'test-trans', got '%v'", urlParams["transaction"])
	}
	if urlParams["segment_2"] != "extra" {
		t.Errorf("Expected URLParams segment_2 to be 'extra', got '%v'", urlParams["segment_2"])
	}
	if urlParams["segment_3"] != "param" {
		t.Errorf("Expected URLParams segment_3 to be 'param', got '%v'", urlParams["segment_3"])
	}
}
