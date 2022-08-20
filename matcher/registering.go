package matcher

import (
	"encoding/json"
	"errors"
	"dispatcher/constants"
	"dispatcher/converter"
	"dispatcher/document"
)

var DepartmentRegistering = []Department{}

func RegisterDepartment(department Department) {
	DepartmentRegistering = append(DepartmentRegistering, department)
}

func MatchDepartmentAndTransaction(document document.Document) (result *Transaction, err error) {
FirstLoop:
	for _, department := range DepartmentRegistering {
		if department.Name == document.Department {
			for name, transaction := range department.Transactions {
				if name == document.Transaction {
					result = &transaction
					break FirstLoop
				}
			}
		}
	}

	if result == nil {
		err = errors.New(constants.TRANSACTION_NOT_FOUND)
	}

	return result, err
}

func formToString(form document.DocumentForm) (string, error) {
	formByte, err := json.Marshal(form)

	return string(formByte), err
}

func RequestHandler(doc document.Document, transaction *Transaction) error {
	formString, _ := formToString(doc.Form)
	(*transaction).SetRequest(formString)
	documentFormValidater := converter.DocumentFormValidater{Request: formString}
	reqType := (*transaction).GetRequestType()

	return documentFormValidater.Validate(reqType)
}
