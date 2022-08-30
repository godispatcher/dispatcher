package model

type Transaction interface {
	Transact() error
	SetRequest(req string)
	GetRequestType() interface{}
	GetResponse() interface{}
	GetOptions() TransactionOptions
	LicenceChecker(licence string) bool
	SetToken(token string)
}

type TransactionOptions struct {
	Security SecurityOptions `json:"security,omitempty"`
}

func (m *TransactionOptions) GetOptions() TransactionOptions {
	return *m
}

type SecurityOptions struct {
	LicenceChecker bool `json:"licence_checker,omitempty"`
}
