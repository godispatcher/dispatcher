package middleware

import (
	"encoding/json"
	"github.com/denizakturk/dispatcher/model"
)

type MiddlewareInterface interface {
	GetDepartmentName() string
	GetTransactionName() string
	InitTransaction(document model.Document) (err error)
	LicenceChecker(licence string) bool
	SetRequest(form model.DocumentForm) error
	Defaults()
	IsLicenceRequired() bool
	model.Transaction
}

func NewTransactionInit(department string, transaction string, transactionType MiddlewareInterface) TransactionInit {
	return TransactionInit{
		Department:  department,
		Transaction: transaction,
		Type:        transactionType,
	}
}

type Middleware[Req, Res any] struct {
	DepartmentName  string
	TransactionName string
	Request         Req
	Response        Res
	LicenceRequired bool
}

func (m *Middleware[Req, Res]) Defaults() {
	m.LicenceRequired = false
}

func (m Middleware[Req, Res]) IsLicenceRequired() bool {
	return m.LicenceRequired
}

func (m Middleware[Req, Res]) GetDepartmentName() string {
	return m.DepartmentName
}

func (m Middleware[Req, Res]) GetTransactionName() string {
	return m.TransactionName
}

func (t Middleware[Req, Res]) GetRequest() any {
	return t.Request
}

func (t Middleware[Req, Res]) GetResponse() any {
	return t.Response
}

func (t Middleware[Req, Res]) LicenceChecker(licence string) bool {
	return true
}

func (m *Middleware[Req, Res]) SetRequest(form model.DocumentForm) error {
	formByte, err := json.Marshal(form)
	if err != nil {
		return err
	}
	err = json.Unmarshal(formByte, &m.Request)
	return err
}
func (m *Middleware[Req, Res]) InitTransaction(document model.Document) (err error) {
	err = m.SetRequest(document.Form)
	if err != nil {
		return err
	}
	return err
}

type TransactionInit struct {
	Type        MiddlewareInterface
	Department  string
	Transaction string
}

func (m TransactionInit) Init(document model.Document) (instance model.Transaction, err error) {
	err = m.Type.InitTransaction(document)
	return m.Type, err
}
