package model

import "net/http"

type RegisterResponseModel struct {
	Header     http.Header
	StatusCode int
	Body       interface{}
}
