package registrant

import (
	"github.com/denizakturk/dispatcher/constants"
	"github.com/denizakturk/dispatcher/handling"
	"github.com/denizakturk/dispatcher/model"
	"net/http"
)

func NewRegisterDispatch() RegisterDispatcher {
	dispatch := RegisterDispatcher{}

	dispatch.MainFunc = func(rw http.ResponseWriter, req *http.Request) {
		body, err := handling.RequestHandle(req)
		if err != nil {
			errDoc := &model.Document{}
			errDoc.Error = err.Error()
			errDoc.Type = constants.DOC_TYPE_ERROR
			documentarist := model.NewDocumentarist(rw, errDoc)
			documentarist.Write()
		}
		inputDoc := handling.RequestBodyToDocument(body)
		documentarist := model.NewDocumentarist(rw, inputDoc)
		documentation := NewDocumentation(&documentarist)

		documentation.DocumentEnforcer()
		documentarist.Write()
	}

	dispatch.Port = "9000"
	return dispatch
}

type RegisterDispatcher struct {
	MainFunc func(http.ResponseWriter, *http.Request)
	Port     string
}
