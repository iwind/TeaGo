package types

import (
	"reflect"
	"strconv"
	"fmt"
	"errors"
)

func Byte(value interface{}) byte {
	return byte(Int(value))
}

func Int(value interface{}) int {
	return int(Int32(value))
}

func Int8(value interface{}) int8 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return int8(value.(int))
	case reflect.Int8:
		return int8(value.(int8))
	case reflect.Int16:
		return int8(value.(int16))
	case reflect.Int32:
		return int8(value.(int32))
	case reflect.Int64:
		return int8(value.(int64))
	case reflect.Uint:
		return int8(value.(uint))
	case reflect.Uint8:
		return int8(value.(uint8))
	case reflect.Uint16:
		return int8(value.(uint16))
	case reflect.Uint32:
		return int8(value.(uint32))
	case reflect.Uint64:
		return int8(value.(uint64))
	case reflect.Float32:
		return int8(value.(float32))
	case reflect.Float64:
		return int8(value.(float64))
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 64)
		if err == nil {
			return int8(result)
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 64)
			if err == nil {
				return int8(floatResult)
			}
		}
	}
	return 0
}

func Int16(value interface{}) int16 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return int16(value.(int))
	case reflect.Int8:
		return int16(value.(int8))
	case reflect.Int16:
		return int16(value.(int16))
	case reflect.Int32:
		return int16(value.(int32))
	case reflect.Int64:
		return int16(value.(int64))
	case reflect.Uint:
		return int16(value.(uint))
	case reflect.Uint8:
		return int16(value.(uint8))
	case reflect.Uint16:
		return int16(value.(uint16))
	case reflect.Uint32:
		return int16(value.(uint32))
	case reflect.Uint64:
		return int16(value.(uint64))
	case reflect.Float32:
		return int16(value.(float32))
	case reflect.Float64:
		return int16(value.(float64))
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 64)
		if err == nil {
			return int16(result)
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 64)
			if err == nil {
				return int16(floatResult)
			}
		}
	}
	return 0
}

func Int64(value interface{}) int64 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return int64(value.(int))
	case reflect.Int8:
		return int64(value.(int8))
	case reflect.Int16:
		return int64(value.(int16))
	case reflect.Int32:
		return int64(value.(int32))
	case reflect.Int64:
		return int64(value.(int64))
	case reflect.Uint:
		return int64(value.(uint))
	case reflect.Uint8:
		return int64(value.(uint8))
	case reflect.Uint16:
		return int64(value.(uint16))
	case reflect.Uint32:
		return int64(value.(uint32))
	case reflect.Uint64:
		return int64(value.(uint64))
	case reflect.Float32:
		return int64(value.(float32))
	case reflect.Float64:
		return int64(value.(float64))
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 64)
		if err == nil {
			return result
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 64)
			if err == nil {
				return int64(floatResult)
			}
		}
	}
	return 0
}

func Uint64(value interface{}) uint64 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return uint64(value.(int))
	case reflect.Int8:
		return uint64(value.(int8))
	case reflect.Int16:
		return uint64(value.(int16))
	case reflect.Int32:
		return uint64(value.(int32))
	case reflect.Int64:
		return uint64(value.(int64))
	case reflect.Uint:
		return uint64(value.(uint))
	case reflect.Uint8:
		return uint64(value.(uint8))
	case reflect.Uint16:
		return uint64(value.(uint16))
	case reflect.Uint32:
		return uint64(value.(uint32))
	case reflect.Uint64:
		return uint64(value.(uint64))
	case reflect.Float32:
		return uint64(value.(float32))
	case reflect.Float64:
		return uint64(value.(float64))
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 64)
		if err == nil {
			return uint64(result)
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 64)
			if err == nil {
				return uint64(floatResult)
			}
		}
	}
	return 0
}

func Int32(value interface{}) int32 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return int32(value.(int))
	case reflect.Int8:
		return int32(value.(int8))
	case reflect.Int16:
		return int32(value.(int16))
	case reflect.Int32:
		return int32(value.(int32))
	case reflect.Int64:
		return int32(value.(int64))
	case reflect.Uint:
		return int32(value.(uint))
	case reflect.Uint8:
		return int32(value.(uint8))
	case reflect.Uint16:
		return int32(value.(uint16))
	case reflect.Uint32:
		return int32(value.(uint32))
	case reflect.Uint64:
		return int32(value.(uint64))
	case reflect.Float32:
		return int32(value.(float32))
	case reflect.Float64:
		return int32(value.(float64))
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 32)
		if err == nil {
			return int32(result)
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 32)
			if err == nil {
				return int32(floatResult)
			}
		}
	}
	return 0
}

func Uint(value interface{}) uint {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return uint(value.(int))
	case reflect.Int8:
		return uint(value.(int8))
	case reflect.Int16:
		return uint(value.(uint16))
	case reflect.Int32:
		return uint(value.(int32))
	case reflect.Int64:
		return uint(value.(int64))
	case reflect.Uint:
		return uint(value.(uint))
	case reflect.Uint8:
		return uint(value.(uint8))
	case reflect.Uint16:
		return uint(value.(uint16))
	case reflect.Uint32:
		return uint(value.(uint32))
	case reflect.Uint64:
		return uint(value.(uint64))
	case reflect.Float32:
		return uint(value.(float32))
	case reflect.Float64:
		return uint(value.(float64))
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 32)
		if err == nil {
			return uint(result)
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 32)
			if err == nil {
				return uint(floatResult)
			}
		}
	}
	return 0
}

func Uint8(value interface{}) uint8 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return uint8(value.(int))
	case reflect.Int8:
		return uint8(value.(int8))
	case reflect.Int16:
		return uint8(value.(int16))
	case reflect.Int32:
		return uint8(value.(int32))
	case reflect.Int64:
		return uint8(value.(int64))
	case reflect.Uint:
		return uint8(value.(uint))
	case reflect.Uint8:
		return uint8(value.(uint8))
	case reflect.Uint16:
		return uint8(value.(uint16))
	case reflect.Uint32:
		return uint8(value.(uint32))
	case reflect.Uint64:
		return uint8(value.(uint64))
	case reflect.Float32:
		return uint8(value.(float32))
	case reflect.Float64:
		return uint8(value.(float64))
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 32)
		if err == nil {
			return uint8(result)
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 32)
			if err == nil {
				return uint8(floatResult)
			}
		}
	}
	return 0
}

func Uint16(value interface{}) uint16 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return uint16(value.(int))
	case reflect.Int8:
		return uint16(value.(int8))
	case reflect.Int16:
		return uint16(value.(int16))
	case reflect.Int32:
		return uint16(value.(int32))
	case reflect.Int64:
		return uint16(value.(int64))
	case reflect.Uint:
		return uint16(value.(uint))
	case reflect.Uint8:
		return uint16(value.(uint8))
	case reflect.Uint16:
		return uint16(value.(uint16))
	case reflect.Uint32:
		return uint16(value.(uint32))
	case reflect.Uint64:
		return uint16(value.(uint64))
	case reflect.Float32:
		return uint16(value.(float32))
	case reflect.Float64:
		return uint16(value.(float64))
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 32)
		if err == nil {
			return uint16(result)
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 32)
			if err == nil {
				return uint16(floatResult)
			}
		}
	}
	return 0
}

func Uint32(value interface{}) uint32 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return uint32(value.(int))
	case reflect.Int8:
		return uint32(value.(int8))
	case reflect.Int16:
		return uint32(value.(int16))
	case reflect.Int32:
		return uint32(value.(int32))
	case reflect.Int64:
		return uint32(value.(int64))
	case reflect.Uint:
		return uint32(value.(uint))
	case reflect.Uint8:
		return uint32(value.(uint8))
	case reflect.Uint16:
		return uint32(value.(uint16))
	case reflect.Uint32:
		return uint32(value.(uint32))
	case reflect.Uint64:
		return uint32(value.(uint64))
	case reflect.Float32:
		return uint32(value.(float32))
	case reflect.Float64:
		return uint32(value.(float64))
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 32)
		if err == nil {
			return uint32(result)
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 32)
			if err == nil {
				return uint32(floatResult)
			}
		}
	}
	return 0
}

func Int32Value(value interface{}) (int32, error) {
	if value == nil {
		return 0, errors.New("value should not be nil")
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1, nil
		}
		return 0, nil
	case reflect.Int:
		return int32(value.(int)), nil
	case reflect.Int8:
		return int32(value.(int8)), nil
	case reflect.Int16:
		return int32(value.(int16)), nil
	case reflect.Int32:
		return int32(value.(int32)), nil
	case reflect.Int64:
		return int32(value.(int64)), nil
	case reflect.Uint:
		return int32(value.(uint)), nil
	case reflect.Uint8:
		return int32(value.(uint8)), nil
	case reflect.Uint16:
		return int32(value.(uint16)), nil
	case reflect.Uint32:
		return int32(value.(uint32)), nil
	case reflect.Uint64:
		return int32(value.(uint64)), nil
	case reflect.Float32:
		return int32(value.(float32)), nil
	case reflect.Float64:
		return int32(value.(float64)), nil
	case reflect.String:
		var result, err = strconv.ParseInt(value.(string), 10, 32)
		if err == nil {
			return int32(result), nil
		} else {
			floatResult, err := strconv.ParseFloat(value.(string), 32)
			if err == nil {
				return int32(floatResult), nil
			}
			return 0, err
		}
	}
	return 0, nil
}

func Float64(value interface{}) float64 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return float64(value.(int))
	case reflect.Int8:
		return float64(value.(int8))
	case reflect.Int16:
		return float64(value.(int16))
	case reflect.Int32:
		return float64(value.(int32))
	case reflect.Int64:
		return float64(value.(int64))
	case reflect.Uint:
		return float64(value.(uint))
	case reflect.Uint8:
		return float64(value.(uint8))
	case reflect.Uint16:
		return float64(value.(uint16))
	case reflect.Uint32:
		return float64(value.(uint32))
	case reflect.Uint64:
		return float64(value.(uint64))
	case reflect.Float32:
		return float64(value.(float32))
	case reflect.Float64:
		return float64(value.(float64))
	case reflect.String:
		floatResult, err := strconv.ParseFloat(value.(string), 64)
		if err == nil {
			return floatResult
		}
	}
	return 0
}

func Float32(value interface{}) float32 {
	if value == nil {
		return 0
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		if value.(bool) {
			return 1
		}
		return 0
	case reflect.Int:
		return float32(value.(int))
	case reflect.Int8:
		return float32(value.(int8))
	case reflect.Int16:
		return float32(value.(int16))
	case reflect.Int32:
		return float32(value.(int32))
	case reflect.Int64:
		return float32(value.(int64))
	case reflect.Uint:
		return float32(value.(uint))
	case reflect.Uint8:
		return float32(value.(uint8))
	case reflect.Uint16:
		return float32(value.(uint16))
	case reflect.Uint32:
		return float32(value.(uint32))
	case reflect.Uint64:
		return float32(value.(uint64))
	case reflect.Float32:
		return float32(value.(float32))
	case reflect.Float64:
		return float32(value.(float64))
	case reflect.String:
		floatResult, err := strconv.ParseFloat(value.(string), 32)
		if err == nil {
			return float32(floatResult)
		}
	}
	return 0
}

func Bool(value interface{}) bool {
	if value == nil {
		return false
	}

	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Bool:
		return value.(bool)
	}
	return Int64(value) > 0
}

func String(value interface{}) string {
	if value == nil {
		return ""
	}
	valueString, ok := value.(string)
	if ok {
		return valueString
	}
	return fmt.Sprintf("%#v", value)
}

func Compare(value1 interface{}, value2 interface{}) bool {
	if value1 == nil {
		return false
	}

	var kind = reflect.TypeOf(value1).Kind()
	switch kind {
	case reflect.Bool:
		return Int(value1) > Int(value2)
	case reflect.Int:
		return Int(value1) > Int(value2)
	case reflect.Int8:
		return Int8(value1) > Int8(value2)
	case reflect.Int16:
		return Int16(value1) > Int16(value2)
	case reflect.Int32:
		return Int32(value1) > Int32(value2)
	case reflect.Int64:
		return Int64(value1) > Int64(value2)
	case reflect.Uint:
		return Uint(value1) > Uint(value2)
	case reflect.Uint8:
		return Uint8(value1) > Uint8(value2)
	case reflect.Uint16:
		return Uint16(value1) > Uint16(value2)
	case reflect.Uint32:
		return Uint32(value1) > Uint32(value2)
	case reflect.Uint64:
		return Uint64(value1) > Uint64(value2)
	case reflect.Float32:
		return Float32(value1) > Float32(value2)
	case reflect.Float64:
		return Float64(value1) > Float64(value2)
	case reflect.String:
		return String(value1) > String(value2)
	}
	return String(value1) > String(value2)
}
