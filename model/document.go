package model

import (
	"dispatcher/utilities"
	"encoding/json"
	"reflect"
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
	Department         string             `json:"department,omitempty"`
	Transaction        string             `json:"transaction,omitempty"`
	Type               string             `json:"type,omitempty"`
	Procedure          interface{}        `json:"procedure,omitempty"`
	Form               DocumentForm       `json:"form,omitempty"`
	Output             interface{}        `json:"output,omitempty"`
	Error              interface{}        `json:"error,omitempty"`
	Dispatchings       []*Document         `json:"dispatchings,omitempty"`
	ChainRequestOption ChainRequestOption `json:"chain_request_option,omitempty"`
}

type ProcedureItem struct {
	Require bool   `json:"require"`
	IsEmpty bool   `json:"is_empty"`
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
