package model

import (
	"net/http"
)

type ServerOption struct {
	Header http.Header
}

type CORSOptions struct {
	AllowedOrigins    []string // e.g., ["*"] or ["https://example.com", "https://app.example.com"]
	AllowedMethods    []string // e.g., ["GET","POST","PUT","DELETE","OPTIONS"]
	AllowedHeaders    []string // e.g., ["Content-Type","Authorization"]
	ExposeHeaders     []string // headers exposed to browser
	AllowCredentials  bool     // set Access-Control-Allow-Credentials: true
	MaxAge            int      // seconds for Access-Control-Max-Age
	EnforceSameOrigin bool     // if true, only allow requests where Origin host matches request Host
}

// WithDefaults returns a copy of options where zero-values are filled with sensible defaults.
func (o *CORSOptions) WithDefaults() *CORSOptions {
	if o == nil {
		return (&CORSOptions{}).WithDefaults()
	}
	out := *o
	if len(out.AllowedOrigins) == 0 {
		out.AllowedOrigins = []string{"*"}
	}
	if len(out.AllowedMethods) == 0 {
		out.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	}
	if len(out.AllowedHeaders) == 0 {
		out.AllowedHeaders = []string{"Content-Type", "Authorization", "Accept", "Origin", "X-Requested-With"}
	}
	// ExposeHeaders default empty
	// AllowCredentials default false
	// MaxAge default 0 (browser default)
	// EnforceSameOrigin default false
	return &out
}

type ServerInterface interface {
	Init(document Document) Document
	GetRequest() any
	GetResponse() any
	GetOptions() ServerOption
}
