package registrant

import (
	"dispatcher/constants"
	"dispatcher/handling"
	"dispatcher/model"
	"encoding/json"
	"fmt"
	"net/http"
)

func NewRegisterDispatch() RegisterDispatcher {
	dispatch := RegisterDispatcher{}

	dispatch.MainFunc = func(rw http.ResponseWriter, req *http.Request) {
		body := handling.RequestHandle(req)
		doc := handling.RequestBodyToDocument(body)
		transaction, err := MatchDepartmentAndTransaction(*doc)
		responseDoc := model.Document{Department: doc.Department, Transaction: doc.Transaction, Type: doc.Type}
		responder := newResponder(rw, responseDoc)

		if err != nil {
			responder.writeError(err)
			return
		}
		if doc.Type == "procedure" {
			procedure := model.Procedure{}
			procedure.FromRequestType((*transaction).GetRequestType())
			responder.writeProcedure(procedure)
			return
		}
		err = RequestHandler(*doc, transaction)
		if err != nil {
			responder.writeError(err)
			return
		}
		if transaction != nil {
			responder.writeResponse(runningTransaction(transaction), constants.DOC_TYPE_RESULT)
			return
		}
	}
	dispatch.Port = "9000"
	return dispatch
}

type RegisterDispatcher struct {
	MainFunc func(http.ResponseWriter, *http.Request)
	Port     string
}

func runningTransaction(transaction *model.Transaction) interface{} {
	(*transaction).Transact()
	return (*transaction).GetResponse()
}

type Responder struct {
	rw  http.ResponseWriter
	doc model.Document
}

func newResponder(rw http.ResponseWriter, doc model.Document) Responder {
	responder := Responder{rw: rw, doc: doc}
	responder.rw.Header().Add(constants.HTTP_CONTENT_TYPE, constants.HTTP_CONTENT_JSON)

	return responder
}

func (r *Responder) writeResponse(output interface{}, documentType string) {
	r.doc.Type = documentType
	r.doc.Output = output
	response, _ := json.Marshal(r.doc)
	fmt.Fprint(r.rw, string(response))
}

func (r *Responder) writeError(err error) {
	r.doc.Type = constants.DOC_TYPE_ERROR
	r.doc.Error = err.Error()
	response, _ := json.Marshal(r.doc)
	fmt.Fprint(r.rw, string(response))
}

func (r *Responder) writeProcedure(procedure interface{}) {
	r.doc.Procedure = procedure
	response, _ := json.Marshal(r.doc)
	fmt.Fprint(r.rw, string(response))
}
