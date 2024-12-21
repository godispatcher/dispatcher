package model

import "net/http"

type ServerOption struct {
	Header http.Header
}

type ServerInterface interface {
	Init(document Document) Document
	GetRequest() any
	GetResponse() any
	GetOptions() ServerOption
}
