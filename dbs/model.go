package dbs

import (
	"sync"
	"reflect"
	"strings"
	"github.com/iwind/TeaGo/types"
)

var modelMapping = sync.Map{}

// @TODO 支持 pk, notNull, default 等tag
type Model struct {
	CountFields int
	Attrs       []string
	Fields      []string
	Kinds       []reflect.Kind
	KindsMap    map[string]reflect.Kind
	Type        reflect.Type
}

func NewModel(modelPointer interface{}) *Model {
	var sample = reflect.Indirect(reflect.ValueOf(modelPointer))
	var valueType = sample.Type()
	var modelName = valueType.Name()

	cachedModel, ok := modelMapping.Load(modelName)
	if ok {
		return cachedModel.(*Model)
	}
	var model = &Model{
		Type:     valueType,
		KindsMap: map[string]reflect.Kind{},
	}
	var countFields = valueType.NumField()
	for i := 0; i < countFields; i ++ {
		var field = valueType.Field(i)
		var kind = field.Type.Kind()
		if kind != reflect.Bool &&
			kind != reflect.Int &&
			kind != reflect.Int8 &&
			kind != reflect.Int16 &&
			kind != reflect.Int32 &&
			kind != reflect.Int64 &&
			kind != reflect.Uint &&
			kind != reflect.Uint8 &&
			kind != reflect.Uint16 &&
			kind != reflect.Uint32 &&
			kind != reflect.Uint64 &&
			kind != reflect.String &&
			kind != reflect.Float32 &&
			kind != reflect.Float64 {
			continue
		}

		// 查找 field:"字段名"，如果找不到则视为不是字段
		var originField = strings.TrimSpace(field.Tag.Get("field"))
		if len(originField) == 0 {
			continue
		}

		model.CountFields ++
		model.Attrs = append(model.Attrs, field.Name)
		model.Fields = append(model.Fields, originField)
		model.Kinds = append(model.Kinds, kind)
		model.KindsMap[originField] = kind
	}

	modelMapping.Store(modelName, model)
	return model
}

func (this *Model) convertValue(value interface{}, toKind reflect.Kind) interface{} {
	switch toKind {
	case reflect.Bool:
		return types.Bool(value)
	case reflect.Int:
		return types.Int(value)
	case reflect.Int8:
		return types.Int8(value)
	case reflect.Int16:
		return types.Int16(value)
	case reflect.Int32:
		return types.Int32(value)
	case reflect.Int64:
		return types.Int64(value)
	case reflect.Uint:
		return types.Uint(value)
	case reflect.Uint8:
		return types.Uint8(value)
	case reflect.Uint16:
		return types.Uint16(value)
	case reflect.Uint32:
		return types.Uint32(value)
	case reflect.Uint64:
		return types.Uint64(value)
	case reflect.String:
		return types.String(value)
	case reflect.Float32:
		return types.Float32(value)
	case reflect.Float64:
		return types.Float64(value)
	}
	return nil
}

func (this *Model) findAttrWithField(field string) (attr string, found bool) {
	for index, fieldName := range this.Fields {
		if fieldName == field {
			return this.Attrs[index], true
		}
	}
	return "", false
}
