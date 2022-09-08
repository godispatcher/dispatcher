package registrant

import (
	"encoding/json"

	"github.com/denizakturk/dispatcher/handling"
	"github.com/denizakturk/dispatcher/model"
)

var DepartmentRegistering = []model.Department{}

func RegisterDepartment(department model.Department) {
	DepartmentRegistering = append(DepartmentRegistering, department)
}

func formToString(form model.DocumentForm) (string, error) {
	formByte, err := json.Marshal(form)

	return string(formByte), err
}

func RequestHandler(inputDoc *model.Document, transaction *model.Transaction) error {
	formString, _ := formToString(inputDoc.Form)
	documentFormValidater := handling.DocumentFormValidater{Request: formString}
	reqType := (*transaction).GetRequest()

	return documentFormValidater.Validate(reqType)
}
