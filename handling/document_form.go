package handling

import (
	"encoding/json"
	"fmt"
	"github.com/denizakturk/dispatcher/constants"
	"github.com/denizakturk/dispatcher/utilities"
	"reflect"
)

type DocumentFormValidater struct {
	Request string
}

func (v *DocumentFormValidater) Validate(TransactionRequestType interface{}) error {
	var incomingData map[string]interface{}
	json.Unmarshal([]byte(v.Request), &incomingData)
	structMap := reflect.ValueOf(TransactionRequestType)

	typeof := reflect.TypeOf(TransactionRequestType)

	for i := 0; i < structMap.NumField(); i++ {
		field := typeof.Field(i)
		tagOption, _ := utilities.ParseTagToTransactionExchangeTag(string(field.Tag))
		if tagOption.Require != nil && *tagOption.Require {

			if _, ok := incomingData[tagOption.FieldRawname]; !ok {
				return fmt.Errorf(constants.FIELD_NOT_FOUND, field.Name)
			}
			vll := structMap.FieldByName(field.Name)
			switch vll.Type().Name() {
			case "string":
				{
					val := vll.Interface()
					if tagOption.IsEmpty != nil && !*tagOption.IsEmpty && val == "" {
						return fmt.Errorf(constants.FIELD_CANNOT_BE_EMPTY, field.Name)
					}
				}
			case "int", "int32", "int64":
				{
					val := vll.Interface()
					if tagOption.IsEmpty != nil && !*tagOption.IsEmpty && val == nil {
						return fmt.Errorf(constants.FIELD_CANNOT_BE_EMPTY, field.Name)
					}
				}
			}
		}

	}

	return nil
}
