package model

type Transaction interface {
	Transact() error
	SetRequest(req string)
	GetRequestType() interface{}
	GetResponse() interface{}
}
