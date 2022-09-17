package model

import (
	"encoding/json"
	"reflect"
	"strings"

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

func NewProcedureItem(fieldType string, tagOption *utilities.TransactionExchangeTag) ProcedureItem {
	procedureItem := ProcedureItem{}
	if tagOption != nil {
		if tagOption.Require != nil {
			procedureItem.Require = *tagOption.Require
		}
		if tagOption.IsEmpty != nil {
			procedureItem.IsEmpty = *tagOption.IsEmpty
		}
	}
	procedureItem.Type = fieldType

	return procedureItem
}

type StructVariable map[string]interface{}
type SliceVariable []interface{}

type VariableAnalyser struct {
}

func (a VariableAnalyser) ItemAnalysis(variable interface{}) interface{} {
	typeOf := a.TypeOf(variable)
	valueOf := a.ValueOf(variable)
	if typeOf.Kind() == reflect.Ptr {
		valueOf = reflect.Indirect(valueOf)
		typeOf = valueOf.Type()
	}
	var output interface{}
	if typeOf.Kind() == reflect.Struct {
		structVariable := StructVariable{}
		for i := 0; i < valueOf.NumField(); i++ {
			fieldType := typeOf.Field(i)
			fieldValue := valueOf.Field(i)
			tagOption, _ := utilities.ParseTagToTransactionExchangeTag(string(fieldType.Tag))
			if fieldType.Type.Kind() == reflect.Slice {
				val := reflect.New(fieldType.Type.Elem())
				var out []interface{}
				if val.CanInterface() {
					out = append(out, a.ItemAnalysis(val.Interface()))
				} else {
					out = append(out, NewProcedureItem(val.Type().Name(), nil))
				}
				structVariable[tagOption.FieldRawname] = out
			} else {
				if fieldValue.CanInterface() {
					structVariable[tagOption.FieldRawname] = a.ItemAnalysis(fieldValue.Interface())
				} else {
					structVariable[tagOption.FieldRawname] = NewProcedureItem(typeOf.Name(), nil)
				}
			}
		}
		output = structVariable
	} else if typeOf.Kind() == reflect.Slice {
		sliceVariable := SliceVariable{}
		if valueOf.Len() == 0 {
			val := reflect.New(typeOf.Elem())
			if val.CanInterface() {
				sliceVariable = append(sliceVariable, a.ItemAnalysis(val.Interface()))
			} else {
				sliceVariable = append(sliceVariable, NewProcedureItem(val.Type().Name(), nil))
			}
		} else {
			for i := 0; i < valueOf.Len(); i++ {
				fieldType := typeOf.Field(i)
				tagOption, _ := utilities.ParseTagToTransactionExchangeTag(string(fieldType.Tag))
				sliceVariable = append(sliceVariable, NewProcedureItem(fieldType.Type.Name(), &tagOption))
			}
		}
		output = sliceVariable
	} else if typeOf.Kind() == reflect.String {
		output = NewProcedureItem("string", nil)
	} else if typeOf.Kind() == reflect.Func {
		output = NewProcedureItem("function", nil)
	} else if strings.Contains("int|int8|int16|int32|int64", typeOf.Kind().String()) {
		output = NewProcedureItem("number", nil)
	}

	return output
}

func (a VariableAnalyser) ValueOf(variable interface{}) reflect.Value {
	return reflect.ValueOf(variable)
}

func (a VariableAnalyser) TypeOf(variable interface{}) reflect.Type {
	return reflect.TypeOf(variable)
}
