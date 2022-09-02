package coordinator

import (
	"github.com/denizakturk/dispatcher/model"
	"github.com/denizakturk/dispatcher/registrant"
)

func ExecuteTransaction(department string, transaction string, form map[string]interface{}) *model.Document {
	inputDoc := &model.Document{Department: department, Transaction: transaction, Form: model.DocumentForm(form)}
	documentarist := model.NewDocumentarist(nil, inputDoc)
	documentation := registrant.NewDocumentation(&documentarist)
	documentation.DocumentEnforcer()
	return documentation.Documentarist.Output
}
