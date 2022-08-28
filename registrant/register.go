package registrant

import (
	"dispatcher/handling"
	"dispatcher/model"
	"net/http"
)

func NewRegisterDispatch() RegisterDispatcher {
	dispatch := RegisterDispatcher{}

	dispatch.MainFunc = func(rw http.ResponseWriter, req *http.Request) {
		body, err := handling.RequestHandle(req)
		if err != nil {
			errDoc := model.Document{}
			errResponder := model.New(rw, errDoc)
			errResponder.WriteError(err)
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
