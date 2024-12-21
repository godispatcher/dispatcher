package utilities

import (
	"encoding/json"
	"reflect"
	"strings"
)

type StructVariable map[string]interface{}
type SliceVariable []interface{}

func Analysis(variable interface{}, nestedTypes *[]string) interface{} {

	typeOf := reflect.TypeOf(variable)
	valueOf := reflect.ValueOf(variable)
	var output interface{}
	if typeOf == nil {
		return "any"
	}
	switch typeOf.Kind() {
	case reflect.Struct:
		methodVal := valueOf.MethodByName("MarshalJSON")
		if methodVal.IsValid() {
			byteData, _ := json.Marshal(variable)
			output = strings.Trim(string(byteData), "\"")
			break
		}
		structVar := StructVariable{}
		(*nestedTypes) = append(*nestedTypes, typeOf.Name())
	NEXTLOOP:
		for i := 0; i < valueOf.NumField(); i++ {
			f := valueOf.Field(i)
			ft := typeOf.Field(i)
			tagOption, _ := ParseTagToTransactionExchangeTag(string(ft.Tag))
			switch f.Type().Kind() {
			case reflect.Map, reflect.Slice, reflect.Ptr:
				backIcon := strings.Builder{}
				for i := len(*nestedTypes) - 1; i >= 0; i-- {
					item := (*nestedTypes)[i]
					backIcon.WriteString("<")
					if strings.Compare(ft.Type.Elem().Name(), item) == 0 {
						structVar[tagOption.FieldRawname] = backIcon.String() + "- Parent"
						continue NEXTLOOP
					}
				}
			}
			if f.CanInterface() {
				structVar[tagOption.FieldRawname] = Analysis(f.Interface(), nestedTypes)
			} else if f.CanAddr() {
				structVar[tagOption.FieldRawname] = Analysis(f.Addr().Interface(), nestedTypes)
			} else {
				structVar[tagOption.FieldRawname] = Analysis(reflect.New(f.Type()).Interface(), nestedTypes)
			}
		}
		output = structVar
		if len(*nestedTypes) > 0 {
			*nestedTypes = (*nestedTypes)[:len(*nestedTypes)-1]
		}
	case reflect.String:
		output = valueOf.Type().Kind().String()
	case reflect.Ptr:
		lastElemType := typeOf.Elem()
		for ok := lastElemType.Kind() == reflect.Ptr; ok; ok = lastElemType.Kind() == reflect.Ptr {
			lastElemType = lastElemType.Elem()
		}
		output = Analysis(reflect.New(lastElemType).Elem().Interface(), nestedTypes)
	case reflect.Slice:
		var sliceMap []interface{}
		sliceMap = append(sliceMap, Analysis(reflect.New(typeOf.Elem()).Interface(), nestedTypes))
		output = sliceMap
	case reflect.Map:
		mapData := make(map[string][]interface{})
		mapData[valueOf.Type().Key().String()] = append(mapData[valueOf.Type().Key().String()], Analysis(reflect.New(valueOf.Type().Elem()).Interface(), nestedTypes))
		output = mapData
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		output = "Number"
	default:
		output = typeOf.Kind().String()
	}
	return output
}
