package coordinator

import (
	"github.com/denizakturk/dispatcher/model"
	"github.com/denizakturk/dispatcher/registrant"
)

type Coordinator struct {
	Security struct{ Licence string }
}

func (c Coordinator) ExecuteTransaction(department string, transaction string, form map[string]interface{}) *model.Document {
	inputDoc := &model.Document{
		Department:  department,
		Transaction: transaction,
		Form:        model.DocumentForm(form),
		Security: &model.Security{
			Licence: c.Security.Licence,
		},
	}
	documentarist := model.NewDocumentarist(nil, inputDoc)
	documentation := registrant.NewDocumentation(&documentarist)
	documentation.DocumentEnforcer()
	return documentation.Documentarist.Output
}
