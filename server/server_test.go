package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/model"
	"github.com/godispatcher/dispatcher/transaction"
)

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

type mockServer struct {
	model.ServerInterface
	request  any
	response any
}

func (m mockServer) Init(document model.Document) model.Document { return model.Document{} }
func (m mockServer) GetRequest() any                             { return m.request }
func (m mockServer) GetResponse() any                            { return m.response }
func (m mockServer) GetOptions() model.ServerOption              { return model.ServerOption{} }
