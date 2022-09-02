package model

import (
	"github.com/denizakturk/dispatcher/constants"
	"encoding/json"
	"fmt"
	"net/http"
)

type Documentarist struct {
	Input          *Document
	Output         *Document
	ResponseWriter http.ResponseWriter
}

func NewDocumentarist(rw http.ResponseWriter, Input *Document) Documentarist {
	documentarist := Documentarist{Input: Input, ResponseWriter: rw, Output: &Document{}}
	documentarist.ResponseWriter.Header().Add(constants.HTTP_CONTENT_TYPE, constants.HTTP_CONTENT_JSON)

	return documentarist
}

func (m *Documentarist) WriteResponse(output interface{}, documentType string) {
	m.Output.Type = documentType
	m.Output.Output = output
	response, _ := json.Marshal(m.Output)
	fmt.Fprint(m.ResponseWriter, string(response))
}

func (m *Documentarist) WriteError(err error) {
	m.Output.Type = constants.DOC_TYPE_ERROR
	m.Output.Error = err.Error()
	response, _ := json.Marshal(m.Output)
	fmt.Fprint(m.ResponseWriter, string(response))
}

func (m *Documentarist) WriteProcedure(procedure interface{}) {
	m.Output.Procedure = procedure
	response, _ := json.Marshal(m.Output)
	fmt.Fprint(m.ResponseWriter, string(response))
}

func (m *Documentarist) WriteDocument() {
	if &m.Output.Type == nil || m.Output.Type == "" {
		m.Output.Type = constants.DOC_TYPE_RESULT
	}
	response, _ := json.Marshal(m.Output)
	fmt.Fprint(m.ResponseWriter, string(response))
}

func (m *Documentarist) Write() {
	response, _ := json.Marshal(m.Output)
	fmt.Fprint(m.ResponseWriter, string(response))
}
