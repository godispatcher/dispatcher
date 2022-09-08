package handling

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/denizakturk/dispatcher/constants"
	"github.com/denizakturk/dispatcher/utilities"
)

type DocumentFormValidater struct {
	Request string
}

func (v *DocumentFormValidater) Validate(TransactionRequestType interface{}) error {
	var incomingData map[string]interface{}
	json.Unmarshal([]byte(v.Request), &incomingData)
	var typeof reflect.Type
	var valueof reflect.Value
	valueof = reflect.ValueOf(TransactionRequestType)
	typeof = reflect.TypeOf(TransactionRequestType)
	if valueof.Kind().String() == "ptr" {
		valueof = reflect.Indirect(valueof)
		typeof = valueof.Type()
	}
	fmt.Println("Incoming Data")
	fmt.Printf("%+v\n", incomingData)

	for i := 0; i < valueof.NumField(); i++ {
		field := typeof.Field(i)
		tagOption, _ := utilities.ParseTagToTransactionExchangeTag(string(field.Tag))
		if tagOption.Require != nil && *tagOption.Require {

			if _, ok := incomingData[tagOption.FieldRawname]; !ok {
				return fmt.Errorf(constants.FIELD_NOT_FOUND, field.Name)
			}
			vll := valueof.FieldByName(field.Name)
			var val any
			switch vll.Type().Name() {
			case "string":
				{

					//val := vll.Interface()
					if v, ok := incomingData[tagOption.FieldRawname]; ok {
						val = v
					}
					fmt.Println("STRING_VALUE")
					fmt.Println(val)
					if tagOption.IsEmpty != nil && !*tagOption.IsEmpty && val == "" {
						return fmt.Errorf(constants.FIELD_CANNOT_BE_EMPTY, field.Name)
					}
				}
			case "int", "int32", "int64":
				{
					if v, ok := incomingData[tagOption.FieldRawname]; ok {
						val = v
					}
					fmt.Println("NUMBER_VALUE")
					fmt.Println(val)
					//val := vll.Interface()
					if tagOption.IsEmpty != nil && !*tagOption.IsEmpty && val == nil {
						return fmt.Errorf(constants.FIELD_CANNOT_BE_EMPTY, field.Name)
					}
				}
			}
		}

	}

	return nil
}
