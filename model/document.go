package model

import (
	"encoding/json"
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
