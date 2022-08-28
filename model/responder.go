package model

import (
	"dispatcher/constants"
	"encoding/json"
	"fmt"
	"net/http"
)

func New(rw http.ResponseWriter, doc Document) Responder {
	responder := Responder{rw: rw, Doc: doc}
	responder.rw.Header().Add(constants.HTTP_CONTENT_TYPE, constants.HTTP_CONTENT_JSON)

	return responder
}

type Responder struct {
	Doc Document
	rw  http.ResponseWriter
}

func (r *Responder) WriteResponse(output interface{}, documentType string) {
	r.Doc.Type = documentType
	r.Doc.Output = output
	response, _ := json.Marshal(r.Doc)
	fmt.Fprint(r.rw, string(response))
}

func (r *Responder) WriteError(err error) {
	r.Doc.Type = constants.DOC_TYPE_ERROR
	r.Doc.Error = err.Error()
	response, _ := json.Marshal(r.Doc)
	fmt.Fprint(r.rw, string(response))
}

func (r *Responder) WriteProcedure(procedure interface{}) {
	r.Doc.Procedure = procedure
	response, _ := json.Marshal(r.Doc)
	fmt.Fprint(r.rw, string(response))
}

func (r *Responder) WriteDocument() {
	if &r.Doc.Type == nil || r.Doc.Type == "" {
		r.Doc.Type = constants.DOC_TYPE_RESULT
	}
	response, _ := json.Marshal(r.Doc)
	fmt.Fprint(r.rw, string(response))
}
