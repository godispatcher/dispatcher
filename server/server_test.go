package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/middleware"
	"github.com/godispatcher/dispatcher/model"
	"github.com/godispatcher/dispatcher/transaction"
	"github.com/godispatcher/dispatcher/utilities"
)

var validationTransactionCalled bool

type validationServerRequest struct {
	Age int `json:"age" require:"true"`
}

type validationServerResponse struct{}

type validationServerTransaction struct {
	middleware.Middleware[validationServerRequest, validationServerResponse]
}

func (*validationServerTransaction) SetSelfRunables() error  { return nil }
func (*validationServerTransaction) SetupTransaction() error { return nil }
func (*validationServerTransaction) Transact() error {
	validationTransactionCalled = true
	return nil
}

func TestServerInitRejectsRequestTypeMismatchBeforeTransact(t *testing.T) {
	validationTransactionCalled = false
	service := Server[validationServerTransaction, *validationServerTransaction]{}
	result := service.Init(model.Document{
		Department:  "Validation",
		Transaction: "typeMismatch",
		Form:        map[string]any{"age": "not-a-number"},
	})

	if result.Type != "Error" {
		t.Fatalf("result type = %q, want Error", result.Type)
	}
	if validationTransactionCalled {
		t.Fatal("Transact() was called for an invalid request")
	}
}

func TestApiDocServer_JSON(t *testing.T) {
	req, err := http.NewRequest("GET", "/help?format=json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := ApiDocServer{}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "application/json")
	}

	if !strings.Contains(rr.Body.String(), "departments") {
		t.Errorf("handler returned unexpected body: %v", rr.Body.String())
	}
}

func TestApiDocServer_HTML(t *testing.T) {
	req, err := http.NewRequest("GET", "/help", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := ApiDocServer{}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "text/html")
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("handler returned unexpected body: missing DOCTYPE")
	}
	if !strings.Contains(body, "GoDispatcher API Doc") {
		t.Errorf("handler returned unexpected body: missing title")
	}
	// Dark mode check
	if !strings.Contains(body, "prefers-color-scheme: dark") {
		t.Errorf("handler returned unexpected body: missing dark mode support")
	}
}

func TestApiDocServer_Toon(t *testing.T) {
	// Add dummy data to DispatcherHolder
	department.DispatcherHolder = nil
	department.DispatcherHolder.Add("Auth", transaction.TransactionBucketItem{
		Name: "login",
		Transaction: mockServer{
			request: struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}{},
			response: struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{},
		},
	})

	req, err := http.NewRequest("GET", "/help?format=toon", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := ApiDocServer{}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedHeader := "departments[1]:"
	if !strings.Contains(body, expectedHeader) {
		t.Errorf("Expected header %q not found in body", expectedHeader)
	}

	expectedDept := "name: Auth"
	if !strings.Contains(body, expectedDept) {
		t.Errorf("Expected department %q not found in body", expectedDept)
	}

	expectedTransHeader := "transactions[1]:"
	if !strings.Contains(body, expectedTransHeader) {
		t.Errorf("Expected transactions header %q not found in body", expectedTransHeader)
	}

	expectedTransName := "name: login"
	if !strings.Contains(body, expectedTransName) {
		t.Errorf("Expected transaction name %q not found in body", expectedTransName)
	}
}

func TestApiDocServer_YAML(t *testing.T) {
	req, err := http.NewRequest("GET", "/help?format=yaml", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := ApiDocServer{}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/x-yaml" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "application/x-yaml")
	}

	if !strings.Contains(rr.Body.String(), "departments:") {
		t.Errorf("handler returned unexpected body: %v", rr.Body.String())
	}
}

type mockServer struct {
	model.ServerInterface
	request  any
	response any
}

type failingResponseMarshaler struct{}

func (failingResponseMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("response marshal failed")
}

func (m mockServer) Init(document model.Document) model.Document { return model.Document{} }
func (m mockServer) GetRequest() any                             { return m.request }
func (m mockServer) GetResponse() any                            { return m.response }
func (m mockServer) GetOptions() model.ServerOption              { return model.ServerOption{} }

func TestApiDocServer_ResponseSchema(t *testing.T) {
	department.DispatcherHolder = nil
	department.DispatcherHolder.Add("Schema", transaction.TransactionBucketItem{
		Name: "show",
		Transaction: mockServer{
			request: struct {
				Query string `json:"query" require:"true" is_empty:"false" example:"laptop"`
			}{},
			response: struct {
				Name string `json:"name"`
			}{},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/help?format=json", nil)
	rr := httptest.NewRecorder()
	ApiDocServer{}.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var body struct {
		Departments []struct {
			Transactions []struct {
				Request struct {
					Schema  utilities.Schema `json:"schema"`
					Example model.Document   `json:"example"`
				} `json:"request"`
				Response struct {
					Schema utilities.Schema `json:"schema"`
				} `json:"response"`
			} `json:"transactions"`
		} `json:"departments"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	transactionDoc := body.Departments[0].Transactions[0]
	output := transactionDoc.Response.Schema
	if output.Type != "object" || output.Properties["name"].Type != "string" {
		t.Fatalf("unexpected response schema: %#v", output)
	}
	if request := transactionDoc.Request; request.Schema.Required[0] != "query" || request.Example.Department != "Schema" || request.Example.Transaction != "show" {
		t.Fatalf("unexpected request documentation: %#v", request)
	}
	form := transactionDoc.Request.Example.Form
	if form["query"] != "laptop" {
		t.Fatalf("unexpected request example form: %#v", transactionDoc.Request.Example.Form)
	}
}

func TestApiDocServer_HTMLSeparatesExampleAndSchemas(t *testing.T) {
	department.DispatcherHolder = nil
	department.DispatcherHolder.Add("Docs", transaction.TransactionBucketItem{
		Name: "search",
		Transaction: mockServer{
			request: struct {
				Query string `json:"query" require:"true" example:"laptop"`
			}{},
			response: struct{}{},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/help", nil)
	rr := httptest.NewRecorder()
	ApiDocServer{}.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	for _, label := range []string{"Gönderilebilir İstek", "Request Şeması", "Response Şeması", "İsteği JSON Kopyala"} {
		if !strings.Contains(rr.Body.String(), label) {
			t.Errorf("HTML output is missing %q", label)
		}
	}
}

func TestApiDocServer_ShortHTMLDoesNotRenderSchemaPanels(t *testing.T) {
	department.DispatcherHolder = nil
	department.DispatcherHolder.Add("Docs", transaction.TransactionBucketItem{
		Name:        "search",
		Transaction: mockServer{request: struct{}{}, response: struct{}{}},
	})

	req := httptest.NewRequest(http.MethodGet, "/help?short=1", nil)
	rr := httptest.NewRecorder()
	ApiDocServer{}.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "Gönderilebilir İstek") {
		t.Fatal("short documentation unexpectedly contains schema panels")
	}
}

func TestApiDocServer_AnalysisError(t *testing.T) {
	department.DispatcherHolder = nil
	department.DispatcherHolder.Add("Schema", transaction.TransactionBucketItem{
		Name: "broken",
		Transaction: mockServer{
			request:  struct{}{},
			response: failingResponseMarshaler{},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/help?format=json", nil)
	rr := httptest.NewRecorder()
	ApiDocServer{}.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestRateLimiter(t *testing.T) {
	doc := model.Document{
		Department:  "Test",
		Transaction: "Limited",
		Options: &model.TransactionOptions{
			RateLimiter: model.RateLimitOptions{
				Enabled: true,
				Limit:   2,
				Window:  60,
				Scope:   model.ScopeIP,
			},
		},
	}

	// 1. request
	key := utilities.GenerateKey(doc.Department, doc.Transaction, string(doc.Options.RateLimiter.Scope), "127.0.0.1", "", "")
	rl := utilities.GetRateLimiter(key, doc.Options.RateLimiter.Limit, doc.Options.RateLimiter.Window)
	res1 := rl.Allow()
	if !res1.Allowed {
		t.Errorf("First request should be allowed")
	}

	// 2. request
	res2 := rl.Allow()
	if !res2.Allowed {
		t.Errorf("Second request should be allowed")
	}

	// 3. request (should be limited)
	res3 := rl.Allow()
	if res3.Allowed {
		t.Errorf("Third request should be limited")
	}
	if res3.RetryAfter <= 0 {
		t.Errorf("RetryAfter should be positive")
	}
}
