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

		var lastResponse interface{} = nil

		if doc.Type == constants.DOC_TYPE_PROCEDURE {
			responder.doc.Procedure = output
		} else {
			responder.doc.Output = output
			lastResponse = *output
		}

		if lastResponse != nil && doc.ChainRequestOption != nil {
			lastResponse = responseTransformer(lastResponse, doc.ChainRequestOption)
		}

		if doc.Dispatchings != nil {
			for _, val := range responder.doc.Dispatchings {
				val.Form.FromInterface(lastResponse)
				err := dispatchTracker(val)
				if err != nil {
					val.Error = err.Error()
					val.Type = constants.DOC_TYPE_ERROR
				}
			}
		}
		responder.writeDocument()
	}

	dispatch.Port = "9000"
	return dispatch
}

func DocumentHandler(doc model.Document) model.Document {
	transaction, err := MatchDepartmentAndTransaction(doc)
	resultDocument := doc
	if err != nil {
		resultDocument.Error = err.Error()
		return resultDocument
	}
	output, err := transactionRunner(transaction, &resultDocument)

	if err != nil {
		resultDocument.Error = err.Error()
		return resultDocument
	}

	var lastResponse interface{} = nil

	if doc.Type == constants.DOC_TYPE_PROCEDURE {
		resultDocument.Procedure = output
	} else {
		resultDocument.Output = output
		lastResponse = *output
	}

	if lastResponse != nil && doc.ChainRequestOption != nil {
		lastResponse = responseTransformer(lastResponse, doc.ChainRequestOption)
	}

	if doc.Dispatchings != nil {
		for _, val := range resultDocument.Dispatchings {
			val.Form.FromInterface(lastResponse)
			err := dispatchTracker(val)
			if err != nil {
				val.Error = err.Error()
				val.Type = constants.DOC_TYPE_ERROR
			}
		}
	}

	return resultDocument
}

func dispatchTracker(doc *model.Document) (err error) {
	transaction, err := MatchDepartmentAndTransaction(*doc)
	if err != nil {
		return err
	}
	output, err := transactionRunner(transaction, doc)
	if err != nil {
		return err
	}
	var lastResponse interface{} = nil
	if doc.Type == constants.DOC_TYPE_PROCEDURE {
		doc.Procedure = output
	} else {
		doc.Output = output
		lastResponse = output
	}

	if lastResponse != nil && doc.ChainRequestOption != nil {
		lastResponse = responseTransformer(lastResponse, doc.ChainRequestOption)
	}
	// TODO: lastResponse doc.ChainRequestOption a göre işlemden geçirildikten sonra val.Form.FromInterface fonksiyonuna aktarılacak
	// Ana blokta da bu işlemin aynı gerekiyor.
	if doc.Dispatchings != nil {
		for _, val := range doc.Dispatchings {
			val.Form.FromInterface(lastResponse)
			err = dispatchTracker(val)
			if err != nil {
				return err
			}
		}
	}

	return
}

func transactionRunner(transaction *model.Transaction, doc *model.Document) (output *interface{}, err error) {
	if doc.Type == constants.DOC_TYPE_PROCEDURE {
		inputProcedure := &model.Procedure{}
		outputProcedure := &model.Procedure{}
		inputProcedure.FromRequestType((*transaction).GetRequestType())
		outputProcedure.FromResponseType((*transaction).GetResponse())
		output := interface{}(inputProcedure)
		doc.Output = outputProcedure
		doc.Type = constants.DOC_TYPE_PROCEDURE
		return &output, nil
	}
	err = RequestHandler(*doc, transaction)
	if err != nil {
		return nil, err
	}

	if transaction != nil {
		err := (*transaction).Transact()
		output := (*transaction).GetResponse()
		doc.Type = constants.DOC_TYPE_RESULT
		return &output, err
	}

	return nil, errors.New("An unidentified error has occurred")
}

type RegisterDispatcher struct {
	MainFunc func(http.ResponseWriter, *http.Request)
	Port     string
}

func responseTransformer(response interface{}, chainRequestOption model.ChainRequestOption) interface{} {
	responseByte, _ := json.Marshal(response)
	var responseMap map[string]interface{}
	json.Unmarshal(responseByte, &responseMap)
	for key, val := range chainRequestOption {
		if _, ok := responseMap[key]; ok {
			responseMap[val.(string)] = responseMap[key]
		}
	}

	return responseMap
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
	if &r.doc.Type == nil || r.doc.Type == "" {
		r.doc.Type = constants.DOC_TYPE_RESULT
	}
	response, _ := json.Marshal(r.doc)
	fmt.Fprint(r.rw, string(response))
}
