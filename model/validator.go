package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/godispatcher/dispatcher/constants"
	"github.com/godispatcher/dispatcher/utilities"
)

type DocumentFormValidater struct {
	Request string
}

func (v *DocumentFormValidater) Validate(transactionRequestType interface{}) error {
	targetType := reflect.TypeOf(transactionRequestType)
	if targetType == nil {
		return errorsInvalidRequestType()
	}
	for targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
		if targetType == nil {
			return errorsInvalidRequestType()
		}
	}
	if targetType.Kind() != reflect.Struct {
		return errorsInvalidRequestType()
	}

	raw := json.RawMessage(v.Request)
	typedTarget := reflect.New(targetType).Interface()
	if err := json.Unmarshal(raw, typedTarget); err != nil {
		if !json.Valid(raw) {
			return fmt.Errorf("invalid request JSON: %w", err)
		}
		return fmt.Errorf("request JSON does not match %s: %w", targetType, err)
	}

	return validateJSONValue(raw, targetType, "")
}

func errorsInvalidRequestType() error {
	return fmt.Errorf("request type must be a struct or pointer to a struct")
}

func validateJSONValue(raw json.RawMessage, targetType reflect.Type, path string) error {
	for targetType.Kind() == reflect.Pointer {
		if isJSONNull(raw) {
			return nil
		}
		targetType = targetType.Elem()
	}
	if implementsJSONUnmarshaler(targetType) {
		return nil
	}
	if isJSONNull(raw) {
		switch targetType.Kind() {
		case reflect.Interface, reflect.Map, reflect.Slice:
			return nil
		default:
			if path == "" {
				path = "request"
			}
			return fmt.Errorf("the field named %s cannot be null", path)
		}
	}

	switch targetType.Kind() {
	case reflect.Struct:
		return validateJSONObject(raw, targetType, path)
	case reflect.Slice:
		if targetType.Elem().Kind() == reflect.Uint8 {
			return nil
		}
		fallthrough
	case reflect.Array:
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return err
		}
		for i, item := range items {
			if err := validateJSONValue(item, targetType.Elem(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		var items map[string]json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return err
		}
		keys := make([]string, 0, len(items))
		for key := range items {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			itemPath := fmt.Sprintf(`%s[%q]`, path, key)
			if err := validateJSONValue(items[key], targetType.Elem(), itemPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateJSONObject(raw json.RawMessage, targetType reflect.Type, path string) error {
	var incomingData map[string]json.RawMessage
	if err := json.Unmarshal(raw, &incomingData); err != nil {
		return err
	}

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		jsonTag, hasJSONTag := field.Tag.Lookup(constants.OPTION_JSON)
		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "-" {
			continue
		}
		if field.Anonymous && (!hasJSONTag || fieldName == "") {
			if err := validateJSONValue(raw, field.Type, path); err != nil {
				return err
			}
			continue
		}
		if fieldName == "" {
			fieldName = field.Name
		}
		fieldPath := joinFieldPath(path, fieldName)

		tagOption, err := utilities.ParseTagToTransactionExchangeTag(string(field.Tag))
		if err != nil {
			return fmt.Errorf("invalid validation tags for %s: %w", fieldPath, err)
		}
		value, exists := incomingData[fieldName]
		if tagOption.Require != nil && *tagOption.Require && !exists {
			return fmt.Errorf(constants.FIELD_NOT_FOUND, fieldPath)
		}
		if !exists {
			continue
		}
		if tagOption.IsEmpty != nil && !*tagOption.IsEmpty && isEmptyJSONValue(value) {
			return fmt.Errorf(constants.FIELD_CANNOT_BE_EMPTY, fieldPath)
		}
		if err := validateJSONValue(value, field.Type, fieldPath); err != nil {
			return err
		}
	}

	return nil
}

func joinFieldPath(parent, field string) string {
	if parent == "" {
		return field
	}
	return parent + "." + field
}

func isEmptyJSONValue(raw json.RawMessage) bool {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case float64:
		return typed == 0
	case bool:
		return !typed
	case []interface{}:
		return len(typed) == 0
	case map[string]interface{}:
		return len(typed) == 0
	default:
		return false
	}
}

func isJSONNull(raw json.RawMessage) bool {
	return strings.TrimSpace(string(raw)) == "null"
}

func implementsJSONUnmarshaler(targetType reflect.Type) bool {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	return targetType.Implements(unmarshalerType) || reflect.PointerTo(targetType).Implements(unmarshalerType)
}
