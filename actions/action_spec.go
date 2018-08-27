package actions

import (
	"reflect"
	"strings"
)

// Action相关定义
type ActionSpec struct {
	Type reflect.Type

	BeforeFunc *reflect.Value
	AfterFunc  *reflect.Value

	Funcs map[string]*reflect.Value

	Module    string
	PkgPath   string
	ClassName string
}

// 创建新定义
func NewActionSpec(actionPtr ActionWrapper) *ActionSpec {
	ptrValue := reflect.ValueOf(actionPtr)
	valueType := reflect.Indirect(ptrValue).Type()
	spec := &ActionSpec{
		Type:      valueType,
		PkgPath:   valueType.PkgPath(),
		ClassName: valueType.String(),
		Funcs:     map[string]*reflect.Value{},
	}

	ptrType := ptrValue.Type()

	beforeMethod, found := ptrType.MethodByName("Before")
	if found {
		spec.BeforeFunc = &beforeMethod.Func
	}

	afterMethod, found := ptrType.MethodByName("After")
	if found {
		spec.AfterFunc = &afterMethod.Func
	}

	countMethods := ptrType.NumMethod()
	for i := 0; i < countMethods; i ++ {
		method := ptrType.Method(i)
		if !method.Func.IsValid() {
			continue
		}

		if strings.HasPrefix(method.Name, "Run") {
			spec.Funcs[method.Name] = &method.Func
		}
	}

	return spec
}

// 新建一个Action指针
func (this *ActionSpec) NewPtrValue() reflect.Value {
	return reflect.New(this.Type)
}
