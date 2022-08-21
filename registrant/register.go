package registrant

import (
	"dispatcher/constants"
	"dispatcher/handling"
	"dispatcher/model"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func NewRegisterDispatch() RegisterDispatcher {
	dispatch := RegisterDispatcher{}

	dispatch.MainFunc = func(rw http.ResponseWriter, req *http.Request) {
		body := handling.RequestHandle(req)
		doc := handling.RequestBodyToDocument(body)
		transaction, err := MatchDepartmentAndTransaction(*doc)
		//responseDoc := model.Document{Department: doc.Department, Transaction: doc.Transaction, Type: doc.Type, Dispatchings: doc.Dispatchings}
		responder := newResponder(rw, *doc)

		if err != nil {
			responder.writeError(err)
			return
		}
		output, err := transactionRunner(transaction, &responder.doc)

		if err != nil {
			responder.writeError(err)
			return
		}

		responder.doc.Output = output
		lastResponse := output
		if doc.Dispatchings != nil {
			for _, val := range responder.doc.Dispatchings {
				val.Form.FromInterface(lastResponse)
				dispatchTracker(&responder, val)
			}
		}
		responder.writeDocument()
	}

	dispatch.Port = "9000"
	return dispatch
}

func dispatchTracker(responder *Responder, doc *model.Document) {
	transaction, err := MatchDepartmentAndTransaction(*doc)
	if err != nil {
		responder.writeError(err)
		return
	}
	output, err := transactionRunner(transaction, doc)
	if err != nil {
		responder.writeError(err)
		return
	}
	doc.Output = output
	if doc.Dispatchings != nil {
		for _, val := range doc.Dispatchings {
			dispatchTracker(responder, val)
		}
	}
}

func transactionRunner(transaction *model.Transaction, doc *model.Document) (output *interface{}, err error) {
	if doc.Type == constants.DOC_TYPE_PROCEDURE {
		procedure := &model.Procedure{}
		procedure.FromRequestType((*transaction).GetRequestType())
		output := interface{}(procedure)
		doc.Type = constants.DOC_TYPE_PROCEDURE
		return &output, nil
	}
	err = RequestHandler(*doc, transaction)
	if err != nil {
		return nil, err
	}

	if transaction != nil {
		output := runningTransaction(transaction)
		doc.Type = constants.DOC_TYPE_RESULT
		return &output, nil
	}

	return nil, errors.New("An unidentified error has occurred")
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

func (r *Responder) writeDocument() {
	r.doc.Type = constants.DOC_TYPE_RESULT
	response, _ := json.Marshal(r.doc)
	fmt.Fprint(r.rw, string(response))
}
