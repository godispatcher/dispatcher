package middleware

import "encoding/json"

type Middleware[Req any, Res any] struct {
	Request  Req
	Response Res
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
