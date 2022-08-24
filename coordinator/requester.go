package coordinator

import (
	"dispatcher/model"
	"dispatcher/registrant"
)

func ExecuteTransaction(department string, transaction string, form map[string]interface{}) model.Document {
	document := model.Document{Department: department, Transaction: transaction, Form: model.DocumentForm(form)}
	result := registrant.DocumentHandler(document)

	return result
}
