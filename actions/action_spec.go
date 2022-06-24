package actions

import (
	"github.com/iwind/TeaGo/caches"
	"reflect"
	"strings"
)

// ActionSpec Action相关定义
type ActionSpec struct {
	Type reflect.Type

	BeforeFunc *reflect.Value
	AfterFunc  *reflect.Value

	FuncMap map[string]*reflect.Value

	Module    string
	PkgPath   string
	ClassName string

	caches.CacheFactory

	Context *ActionContext
}

// NewActionSpec 创建新定义
func NewActionSpec(actionPtr ActionWrapper) *ActionSpec {
	ptrValue := reflect.ValueOf(actionPtr)
	valueType := reflect.Indirect(ptrValue).Type()
	spec := &ActionSpec{
		Type:      valueType,
		PkgPath:   valueType.PkgPath(),
		ClassName: valueType.String(),
		FuncMap:   map[string]*reflect.Value{},
		Context:   NewActionContext(),
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
	for i := 0; i < countMethods; i++ {
		method := ptrType.Method(i)
		if !method.Func.IsValid() {
			continue
		}

		if strings.HasPrefix(method.Name, "Run") {
			spec.FuncMap[method.Name] = &method.Func
		}
	}

	return spec
}

// NewPtrValue 新建一个Action指针
func (this *ActionSpec) NewPtrValue() reflect.Value {
	actionPtr := reflect.New(this.Type)
	wrapper, ok := actionPtr.Interface().(ActionWrapper)
	if ok {
		wrapper.Object().Context = this.Context
	}
	return actionPtr
}

// HasClassPrefix class名是否包含任一前缀
func (this *ActionSpec) HasClassPrefix(prefix ...string) bool {
	for _, prefix1 := range prefix {
		if strings.HasPrefix(this.ClassName, prefix1) {
			return true
		}
	}
	return false
}
