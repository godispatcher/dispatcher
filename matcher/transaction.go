package matcher

type Transaction interface {
	Transact()
	SetRequest(req string)
	GetRequestType() interface{}
	GetResponse() interface{}
}
