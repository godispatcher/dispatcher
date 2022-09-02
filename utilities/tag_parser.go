package utilities

import (
	"strings"

	"github.com/denizakturk/dispatcher/constants"
)

type TransactionExchangeTag struct {
	Require      *bool  `json:"require"`
	IsEmpty      *bool  `json:"is_empty"`
	FieldRawname string `json:"field_raw_name"`
}

func ParseTagToTransactionExchangeTag(tag string) (result TransactionExchangeTag, err error) {
	options := strings.Split(tag, " ")
	for _, val := range options {
		optionDetail := strings.Split(val, ":")
		switch optionDetail[0] {
		case constants.OPTION_REQUIRE:
			{
				optionDetail[1] = strings.Trim(optionDetail[1], "\"")
				switch optionDetail[1] {
				case "true", "True", "TRUE":
					result.Require = &[]bool{true}[0]
				case "false", "False", "FALSE":
					result.Require = &[]bool{false}[0]
				}
			}
		case constants.OPTION_ISEMPTY:
			{
				optionDetail[1] = strings.Trim(optionDetail[1], "\"")
				switch optionDetail[1] {
				case "true", "True", "TRUE":
					result.IsEmpty = &[]bool{true}[0]
				case "false", "False", "FALSE":
					result.IsEmpty = &[]bool{false}[0]
				}
			}
		case constants.OPTION_JSON:
			{
				optionDetail[1] = strings.Trim(optionDetail[1], "\"")
				optionDetailArgument := strings.Split(optionDetail[1], ",")
				result.FieldRawname = optionDetailArgument[0]
			}
		}
	}

	return
}
