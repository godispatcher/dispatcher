package coordinator

import (
	"dispatcher/model"
	"dispatcher/registrant"
)

func ExecuteTransaction(department string, transaction string, form map[string]interface{}) *model.Document {
	inputDoc := &model.Document{Department: department, Transaction: transaction, Form: model.DocumentForm(form)}
	documentarist := model.NewDocumentarist(nil, inputDoc)
	documentation := registrant.NewDocumentation(&documentarist)
	documentation.DocumentEnforcer()
	return documentation.Documentarist.Output
}
