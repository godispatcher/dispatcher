package model

type Transaction interface {
	Transact() (interface{}, error)
	GetRequest() interface{}
	GetResponse() interface{}
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

type LicenceValidator func(licence string) (isValid bool)

type TransactionHolder struct {
	Name             string
	Slug             string
	Options          TransactionOptions
	InitTransaction  func(document Document) (Transaction, error)
	Type             Transaction
	LicenceValidator LicenceValidator
}
