package model

import (
	"encoding/json"
	"reflect"

	"github.com/denizakturk/dispatcher/utilities"
)

type DocumentForm map[string]interface{}

func (df *DocumentForm) FromInterface(data interface{}) error {
	byteData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(byteData, &df)
}

type ChainRequestOption map[string]interface{}

type Document struct {
	Department         string              `json:"department,omitempty"`
	Transaction        string              `json:"transaction,omitempty"`
	Type               string              `json:"type,omitempty"`
	Procedure          interface{}         `json:"procedure,omitempty"`
	Form               DocumentForm        `json:"form,omitempty"`
	Output             interface{}         `json:"output,omitempty"`
	Error              interface{}         `json:"error,omitempty"`
	Dispatchings       []*Document         `json:"dispatchings,omitempty"`
	ChainRequestOption ChainRequestOption  `json:"chain_request_option,omitempty"`
	Security           *Security           `json:"security,omitempty"`
	Options            *TransactionOptions `json:"options,omitempty"`
}

type Security struct {
	Licence string `json:"licence,omitempty"`
}

type ProcedureItem struct {
	Require bool   `json:"require,omitempty"`
	IsEmpty bool   `json:"is_empty,omitempty"`
	Type    string `json:"type"`
}

type Procedure map[string]ProcedureItem

func (p *Procedure) FromRequestType(requestType interface{}) {
	typeof := reflect.TypeOf(requestType)
	valueOf := reflect.ValueOf(requestType)

	for i := 0; i < valueOf.NumField(); i++ {
		field := typeof.Field(i)
		tagOption, _ := utilities.ParseTagToTransactionExchangeTag(string(field.Tag))

		procedureItem := ProcedureItem{}
		if tagOption.Require != nil {
			procedureItem.Require = *tagOption.Require
		}
		if tagOption.IsEmpty != nil {
			procedureItem.IsEmpty = *tagOption.IsEmpty
		}
		procedureItem.Type = field.Type.Name()
		(*p)[tagOption.FieldRawname] = procedureItem
	}
}

func (p *Procedure) FromResponseType(responseType interface{}) {
	typeof := reflect.TypeOf(responseType)
	valueOf := reflect.ValueOf(responseType)

	if responseType == nil || valueOf.NumField() < 1 {
		return
	}
	for i := 0; i < valueOf.NumField(); i++ {
		field := typeof.Field(i)
		tagOption, _ := utilities.ParseTagToTransactionExchangeTag(string(field.Tag))
		procedureItem := ProcedureItem{}
		procedureItem.Type = field.Type.Name()
		(*p)[tagOption.FieldRawname] = procedureItem
	}
}
