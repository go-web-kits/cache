package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Compress(map[string]string{"a": "b"}) => "map[string]string##{\"a\":\"b\"}"
func Compress(value interface{}) (string, error) {
	var valString string
	var err error
	indirectValue := reflect.Indirect(reflect.ValueOf(value))
	value = indirectValue.Interface()

	switch indirectValue.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
		bs, e := json.Marshal(value)
		valString, err = string(bs), e
	default:
		valString = fmt.Sprint(value)
	}

	return strings.Join([]string{indirectValue.Type().String(), valString}, "##"), err
}

// UnCompress("map[string]string##{\"a\":\"b\"}") => map[string]interface{}{"a": "b"}
// UnCompress("map[int]int##{\"1\":2}") => map[string]interface{}{"1": 2.000000}
func UnCompress(compressed string, obj ...interface{}) (interface{}, error) {
	info := strings.Split(compressed, "##")
	if len(info) != 2 {
		return compressed, nil
	}
	typeName, valString := info[0], info[1]
	var value interface{}
	var err error
	var isObj bool
	if len(obj) > 0 {
		value = obj[0]
		isObj = strings.Contains(typeName, reflect.TypeOf(value).String()[1:])
	}

	isSlice := typeName[0] == "["[0]
	toMap := typeName[0:3] == "map" || info[1][0:1] == "{"
	if toMap || isSlice || isObj {
		if value != nil && reflect.TypeOf(value).Kind() == reflect.Ptr {
			err = json.Unmarshal([]byte(valString), value)
		} else {
			err = json.Unmarshal([]byte(valString), &value)
		}
	} else {
		switch typeName {
		case "int":
			value, err = strconv.Atoi(valString)
		case "int32":
			value, err = strconv.ParseInt(valString, 10, 32)
		case "int64":
			value, err = strconv.ParseInt(valString, 10, 64)
		case "uint":
			value, err = strconv.Atoi(valString)
			value = uint(value.(int))
		case "uint8":
			value, err = strconv.Atoi(valString)
			value = byte(value.(int))
		case "uint64":
			value, err = strconv.ParseUint(valString, 10, 64)
		case "float32":
			value, err = strconv.ParseFloat(valString, 32)
		case "float64":
			value, err = strconv.ParseFloat(valString, 64)
		case "bool":
			if valString == "true" {
				value = true
			} else {
				value = false
			}
		default:
			value = valString
		}
	}

	return value, err
}

func encode(compressed string) string {
	return base64.StdEncoding.EncodeToString([]byte(compressed))
}

func decode(encoded string) (string, error) {
	bs, err := base64.StdEncoding.DecodeString(encoded)
	return string(bs), err
}
