package utilities

import (
	"encoding"
	"encoding/json"
	"reflect"
	"strings"
)

type StructVariable map[string]interface{}
type SliceVariable []interface{}

// hasCustomJSONMarshaling reports whether the type or its pointer implements
// json.Marshaler or encoding.TextMarshaler (or explicitly defines MarshalJSON).
// This helps us classify types like UUID as primitive JSON types (e.g., String)
// instead of falling back to their underlying Go kind (e.g., array/slice of bytes).
func hasCustomJSONMarshaling(t reflect.Type) bool {
	if t == nil {
		return false
	}
	jsonMarshaler := reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	textMarshaler := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

	// Check the type itself and also its pointer form to catch pointer receiver methods.
	if t.Implements(jsonMarshaler) || t.Implements(textMarshaler) {
		return true
	}
	if t.Kind() != reflect.Ptr {
		pt := reflect.PointerTo(t)
		if pt.Implements(jsonMarshaler) || pt.Implements(textMarshaler) {
			return true
		}
	}

	// Fallback explicit method presence check for MarshalJSON by name.
	if _, ok := t.MethodByName("MarshalJSON"); ok {
		return true
	}
	if t.Kind() != reflect.Ptr {
		pt := reflect.PointerTo(t)
		if _, ok := pt.MethodByName("MarshalJSON"); ok {
			return true
		}
	}
	return false
}

func Analysis(variable interface{}, nestedTypes *[]string) interface{} {

	typeOf := reflect.TypeOf(variable)
	valueOf := reflect.ValueOf(variable)
	var output interface{}
	if typeOf == nil {
		return "any"
	}

	// If the type has custom (JSON/Text) marshaling, prefer analyzing its marshaled JSON shape.
	if hasCustomJSONMarshaling(typeOf) {
		byteData, _ := json.Marshal(variable)
		output = strings.Trim(MarshalJSONAnalysis(byteData), "\"")
		return output
	}

	switch typeOf.Kind() {
	case reflect.Struct:
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

func MarshalJSONAnalysis(byteData []byte) string {
	var result interface{}
	err := json.Unmarshal(byteData, &result)
	if err != nil {
		return "Invalid JSON"
	}

	switch result.(type) {
	case string:
		return "String"
	case float64:
		return "Number"
	case bool:
		return "Boolean"
	case []interface{}:
		return "Array"
	case map[string]interface{}:
		return "Object"
	default:
		return "Unknown"
	}
}
