package actions

import (
	"fmt"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	"github.com/iwind/TeaGo/utils/string"
	"reflect"
	"regexp"
	"strings"
)

type Must struct {
	action      *ActionObject
	field       string
	value       interface{}
	valueString string
	valueFloat  float64

	hasErrors bool
	code      int
	errors    []ActionParamError
}

func (this *Must) BeforeAction(actionPtr ActionWrapper, paramName string) (goNext bool) {
	this.action = actionPtr.Object()
	return true
}

func (this *Must) Trim() *Must {
	var reflectValue = reflect.ValueOf(this.value)
	var kind = reflectValue.Kind()

	if kind == reflect.Ptr {
		realValue := reflect.Indirect(reflectValue)
		kind = realValue.Kind()

		if kind == reflect.String {
			realValue.Set(reflect.ValueOf(strings.TrimSpace(realValue.Interface().(string))))
			this.valueString = realValue.Interface().(string)
		}
	}

	return this
}

func (this *Must) Value(value interface{}) *Must {
	this.value = value

	var reflectValue = reflect.ValueOf(value)
	var kind = reflectValue.Kind()

	if kind == reflect.Ptr {
		realValue := reflect.Indirect(reflectValue)
		kind = realValue.Kind()

		if kind == reflect.String {
			this.valueString = realValue.Interface().(string)
		} else {
			this.valueString = fmt.Sprintf("%v", realValue.Interface())
			this.valueFloat = types.Float64(value)
		}
	} else if kind == reflect.String {
		this.valueString = value.(string)
	} else {
		this.valueString = fmt.Sprintf("%v", value)
		this.valueFloat = types.Float64(value)
	}

	return this
}

func (this *Must) Field(field string, value interface{}) *Must {
	this.field = field
	this.Value(value)
	return this
}

func (this *Must) Code(code int) *Must {
	this.code = code
	return this
}

func (this *Must) Require(message string) *Must {
	if this.HasErrors() {
		return this
	}
	if len(this.valueString) == 0 {
		this.addError(message)
	}
	return this
}

// 判断是否在一个值列表中
func (this *Must) In(values interface{}, message string) *Must {
	if this.HasErrors() {
		return this
	}

	if reflect.TypeOf(values).Kind() == reflect.Slice {
		if !lists.Contains(values, this.value) {
			this.addError(message)
		}
	}

	return this
}

func (this *Must) Mobile(message string) *Must {
	if this.HasErrors() {
		return this
	}
	reg, err := stringutil.RegexpCompile("^1[34578]\\d{9}$")
	if err != nil {
		logs.Errorf("%s", err.Error())
		this.addError(err.Error())
		return this
	}

	if !reg.MatchString(this.valueString) {
		this.addError(message)
	}
	return this
}

// 最小长度
func (this *Must) MinLength(length int, message string) *Must {
	if this.HasErrors() {
		return this
	}
	if len(this.valueString) < length {
		this.addError(message)
	}
	return this
}

//  最大字符长度
func (this *Must) MaxCharacters(charactersLength int, message string) *Must {
	if this.HasErrors() {
		return this
	}
	if len([]rune(this.valueString)) > charactersLength {
		this.addError(message)
	}
	return this
}

// 最小字符长度
func (this *Must) MinCharacters(charactersLength int, message string) *Must {
	if this.HasErrors() {
		return this
	}
	if len([]rune(this.valueString)) < charactersLength {
		this.addError(message)
	}
	return this
}

//  最大长度
func (this *Must) MaxLength(length int, message string) *Must {
	if this.HasErrors() {
		return this
	}
	if len(this.valueString) > length {
		this.addError(message)
	}
	return this
}

func (this *Must) Match(expr string, message string) *Must {
	if this.HasErrors() {
		return this
	}
	reg, err := regexp.Compile(expr)
	if err != nil {
		logs.Errorf("Must.Match():%s", err.Error())
		this.addError(err.Error())
		return this
	}

	if !reg.MatchString(this.valueString) {
		this.addError(message)
	}
	return this
}

func (this *Must) Equal(value string, message string) *Must {
	if this.HasErrors() {
		return this
	}
	if this.valueString != value {
		this.addError(message)
	}
	return this
}

func (this *Must) Email(message string) *Must {
	if this.HasErrors() {
		return this
	}
	return this.Match("(?i)^[a-z0-9]+([\\._\\-\\+]*[a-z0-9]+)*@([a-z0-9]+[\\-a-z0-9]*[a-z0-9]+\\.)+[a-z0-9]+$", message)
}

func (this *Must) Gt(value int64, message string) *Must {
	if this.HasErrors() {
		return this
	}

	if this.valueFloat <= float64(value) {
		this.addError(message)
	}
	return this
}

func (this *Must) Gte(value int64, message string) *Must {
	if this.HasErrors() {
		return this
	}

	if this.valueFloat < float64(value) {
		this.addError(message)
	}
	return this
}

func (this *Must) Lt(value int64, message string) *Must {
	if this.HasErrors() {
		return this
	}

	if this.valueFloat >= float64(value) {
		this.addError(message)
	}
	return this
}

func (this *Must) Lte(value int64, message string) *Must {
	if this.HasErrors() {
		return this
	}

	if this.valueFloat > float64(value) {
		this.addError(message)
	}
	return this
}

func (this *Must) Expect(fn func() (message string, success bool)) *Must {
	if this.HasErrors() {
		return this
	}

	message, success := fn()
	if !success {
		this.addError(message)
	}
	return this
}

func (this *Must) HasErrors() bool {
	return this.hasErrors
}

func (this *Must) addError(message string) *Must {
	var found = false
	for index, errorObject := range this.errors {
		if errorObject.Param == this.field {
			this.errors[index].Messages = append(errorObject.Messages, message)
			found = true
			break
		}
	}

	if !found {
		this.errors = append(this.errors, ActionParamError{
			Param:    this.field,
			Messages: []string{message},
		})
	}

	this.hasErrors = true
	panic(this.errors)

	return this
}

func (this *Must) Errors() []ActionParamError {
	return this.errors
}
