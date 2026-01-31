package middleware

import (
	"encoding/json"

	"github.com/godispatcher/dispatcher/model"
)

type MiddlewareRunable func(document model.Document) error

type Middleware[Req any, Res any] struct {
	Request  Req
	Response Res
	Runables []MiddlewareRunable
}

func (m Middleware[Req, Res]) GetRunables() []MiddlewareRunable {
	return m.Runables
}

func (m *Middleware[Req, Res]) AddRunable(runable MiddlewareRunable) {
	m.Runables = append(m.Runables, runable)
}

func (m *Middleware[Req, Res]) SetRunables(runables []MiddlewareRunable) {
	m.Runables = runables
}

func (m *Middleware[Req, Res]) SetRequest(data []byte) error {
	m.Request = *new(Req)
	return json.Unmarshal(data, &m.Request)
}
func (c Middleware[Req, Res]) GetRequest() any {
	return c.Request
}
func (c Middleware[Req, Res]) GetResponse() any {
	return c.Response
}

/*
func (m *Middleware[Req, Res]) SetupTransaction() error {
	return nil
}
*/
