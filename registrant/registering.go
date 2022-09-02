package registrant

import (
	"github.com/denizakturk/dispatcher/constants"
	"github.com/denizakturk/dispatcher/handling"
	"github.com/denizakturk/dispatcher/model"
	"encoding/json"
	"errors"
)

var DepartmentRegistering = []model.Department{}

func RegisterDepartment(department model.Department) {
	DepartmentRegistering = append(DepartmentRegistering, department)
}

func MatchDepartmentAndTransaction(inputDoc *model.Document) (transaction *model.Transaction, err error) {
FirstLoop:
	for _, department := range DepartmentRegistering {
		if department.Name == inputDoc.Department {
			for name, findedTransaction := range department.Transactions {
				if name == inputDoc.Transaction {
					transaction = &findedTransaction
					break FirstLoop
				}
			}
		}
	}

	if transaction == nil {
		err = errors.New(constants.TRANSACTION_NOT_FOUND)
	}

	return transaction, err
}

func formToString(form model.DocumentForm) (string, error) {
	formByte, err := json.Marshal(form)

	return string(formByte), err
}

func RequestHandler(inputDoc *model.Document, transaction *model.Transaction) error {
	formString, _ := formToString(inputDoc.Form)
	(*transaction).SetRequest(formString)
	documentFormValidater := handling.DocumentFormValidater{Request: formString}
	reqType := (*transaction).GetRequestType()

	return documentFormValidater.Validate(reqType)
}
