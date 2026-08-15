package utilities

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/godispatcher/dispatcher/constants"
)

type TransactionExchangeTag struct {
	Require      *bool  `json:"require"`
	IsEmpty      *bool  `json:"is_empty"`
	FieldRawname string `json:"field_raw_name"`
}

func ParseTagToTransactionExchangeTag(tag string) (result TransactionExchangeTag, err error) {
	structTag := reflect.StructTag(tag)
	if value, ok := structTag.Lookup(constants.OPTION_REQUIRE); ok {
		result.Require, err = parseBoolTag(constants.OPTION_REQUIRE, value)
		if err != nil {
			return result, err
		}
	}
	if value, ok := structTag.Lookup(constants.OPTION_ISEMPTY); ok {
		result.IsEmpty, err = parseBoolTag(constants.OPTION_ISEMPTY, value)
		if err != nil {
			return result, err
		}
	}
	if value, ok := structTag.Lookup(constants.OPTION_JSON); ok {
		result.FieldRawname = strings.Split(value, ",")[0]
	}

	return result, nil
}

func parseBoolTag(name, value string) (*bool, error) {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid %s struct tag value %q: %w", name, value, err)
	}
	return &parsed, nil
}
