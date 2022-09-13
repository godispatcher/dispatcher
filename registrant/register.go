package registrant

import (
	"net/http"

	"github.com/denizakturk/dispatcher/handling"
	"github.com/denizakturk/dispatcher/model"
)

func NewRegisterDispatch() RegisterDispatcher {
	dispatch := RegisterDispatcher{}

	dispatch.MainFunc = func(rw http.ResponseWriter, req *http.Request) {
		body, err := handling.RequestHandle(req)
		if err != nil {
			errDoc := &model.Document{}
			documentarist := model.NewDocumentarist(rw, errDoc)
			documentarist.WriteError(err)
			return
		}
		inputDoc, err := handling.RequestBodyToDocument(body)
		if err != nil {
			errDoc := &model.Document{}
			documentarist := model.NewDocumentarist(rw, errDoc)
			documentarist.WriteError(err)
			return
		}
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
