package registrant

import (
	"dispatcher/constants"
	"dispatcher/handling"
	"dispatcher/model"
	"encoding/json"
	"errors"
)

var DepartmentRegistering = []model.Department{}

func RegisterDepartment(department model.Department) {
	DepartmentRegistering = append(DepartmentRegistering, department)
}

func MatchDepartmentAndTransaction(document model.Document) (result *model.Transaction, err error) {
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

func formToString(form model.DocumentForm) (string, error) {
	formByte, err := json.Marshal(form)

	return string(formByte), err
}

func RequestHandler(doc model.Document, transaction *model.Transaction) error {
	formString, _ := formToString(doc.Form)
	(*transaction).SetRequest(formString)
	documentFormValidater := handling.DocumentFormValidater{Request: formString}
	reqType := (*transaction).GetRequestType()

	return documentFormValidater.Validate(reqType)
}
