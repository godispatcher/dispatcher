package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
