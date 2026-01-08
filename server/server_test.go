package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApiDocServer_HTML(t *testing.T) {
	// Not: Test çalışırken templates/help.html dosyasının varlığından emin olunmalı
	// server/server.go içindeki ParseFiles("templates/help.html") proje kök dizinine göre çalışıyor

	srv := ApiDocServer{}
	req, err := http.NewRequest("GET", "/help", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("handler returned wrong content type: got %v want text/html",
			contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("handler returned unexpected body: body does not contain <!DOCTYPE html>")
	}

	if strings.Contains(body, "Template error") {
		t.Errorf("handler returned template error: %v", body)
	}
}

func TestApiDocServer_JSON(t *testing.T) {
	srv := ApiDocServer{}
	req, err := http.NewRequest("GET", "/help?format=json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("handler returned wrong content type: got %v want application/json",
			contentType)
	}
}
