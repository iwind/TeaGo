package dbs

import (
	"github.com/iwind/TeaGo/types"
	"reflect"
	"strings"
)

type Field struct {
	Name               string       // 字段名
	Type               string       // 字段类型
	FullType           string       // 完整的字段类型，包含长度等附加信息
	DataKind           reflect.Kind // 数据类型
	Length             int          // @TODO 暂未实现
	DefaultValueString string
	DefaultValue       interface{}
	AutoIncrement      bool
	IsPrimaryKey       bool   // 是否为主键
	IsNotNull          bool   // 是否为非NULL
	IsUnsigned         bool   // 是否为无符号
	Comment            string // 字段的注释说明
	Collation          string // 字符集信息
	ValueType          byte   // 值类型，ValueType*

	MappingName     string // 映射的名称
	MappingKindName string // 映射的数据类型字符串格式
}

const (
	ValueTypeInvalid byte = iota
	ValueTypeBool
	ValueTypeNumber
	ValueTypeString
	ValueTypeTime // @TODO 暂未实现
)

func (this *Field) parseDataKind() {
	pieces := strings.Split(this.FullType, "(")
	if len(pieces) == 0 {
		this.DataKind = reflect.Invalid
		this.ValueType = ValueTypeInvalid
		return
	}
	this.IsUnsigned = strings.Contains(this.FullType, "unsigned")

	var dbDataType = strings.ToLower(strings.TrimSpace(pieces[0]))
	switch dbDataType {
	// numbers
	case "bit":
		this.DataKind = reflect.Uint8
		this.ValueType = ValueTypeNumber
		this.DefaultValue = uint8(types.Int32(this.DefaultValueString))
	case "tinyint":
		if this.IsUnsigned {
			this.DataKind = reflect.Uint8
			this.DefaultValue = uint8(types.Int32(this.DefaultValueString))
		} else {
			this.DataKind = reflect.Int8
			this.DefaultValue = int8(types.Int32(this.DefaultValueString))
		}
		this.ValueType = ValueTypeNumber
	case "bool":
		if this.IsUnsigned {
			this.DataKind = reflect.Uint8
			this.DefaultValue = uint8(types.Int32(this.DefaultValueString))
		} else {
			this.DataKind = reflect.Int8
			this.DefaultValue = int8(types.Int32(this.DefaultValueString))
		}
		this.ValueType = ValueTypeBool
	case "boolean":
		if this.IsUnsigned {
			this.DataKind = reflect.Uint8
			this.DefaultValue = uint8(types.Int32(this.DefaultValueString))
		} else {
			this.DataKind = reflect.Int8
			this.DefaultValue = int8(types.Int32(this.DefaultValueString))
		}
		this.ValueType = ValueTypeBool
	case "smallint":
		if this.IsUnsigned {
			this.DataKind = reflect.Uint16
			this.DefaultValue = uint16(types.Uint32(this.DefaultValueString))
		} else {
			this.DataKind = reflect.Int16
			this.DefaultValue = int16(types.Int32(this.DefaultValueString))
		}
		this.ValueType = ValueTypeNumber
	case "mediumint":
		if this.IsUnsigned {
			this.DataKind = reflect.Uint32
			this.DefaultValue = types.Uint32(this.DefaultValueString)
		} else {
			this.DataKind = reflect.Int32
			this.DefaultValue = types.Int32(this.DefaultValueString)
		}
		this.ValueType = ValueTypeNumber
	case "int":
		if this.IsUnsigned {
			this.DataKind = reflect.Uint32
			this.DefaultValue = types.Uint32(this.DefaultValueString)
		} else {
			this.DataKind = reflect.Int32
			this.DefaultValue = types.Int32(this.DefaultValueString)
		}
		this.ValueType = ValueTypeNumber
	case "integer":
		if this.IsUnsigned {
			this.DataKind = reflect.Uint32
			this.DefaultValue = types.Uint32(this.DefaultValueString)
		} else {
			this.DataKind = reflect.Int32
			this.DefaultValue = types.Int32(this.DefaultValueString)
		}
		this.ValueType = ValueTypeNumber
	case "bigint":
		if this.IsUnsigned {
			this.DataKind = reflect.Uint64
			this.DefaultValue = types.Uint64(this.DefaultValueString)
		} else {
			this.DataKind = reflect.Int64
			this.DefaultValue = types.Int64(this.DefaultValueString)
		}
		this.ValueType = ValueTypeNumber
	case "decimal":
		if this.IsUnsigned {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		} else {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		}
		this.ValueType = ValueTypeNumber
	case "dec":
		if this.IsUnsigned {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		} else {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		}
		this.ValueType = ValueTypeNumber
	case "float":
		if this.IsUnsigned {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		} else {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		}
		this.ValueType = ValueTypeNumber
	case "double":
		if this.IsUnsigned {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		} else {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		}
		this.ValueType = ValueTypeNumber
	case "double precision":
		if this.IsUnsigned {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		} else {
			this.DataKind = reflect.Float64
			this.DefaultValue = types.Float64(this.DefaultValueString)
		}
		this.ValueType = ValueTypeNumber

		// date & time

	default:
		this.DataKind = reflect.String
		this.ValueType = ValueTypeString
		this.DefaultValue = this.DefaultValueString
	}
}

func (this *Field) ValueTypeName() string {
	var dataType = "string"
	switch this.ValueType {
	case ValueTypeBool:
		dataType = "bool"
	default:
		dataType = this.DataKind.String()
	}
	return dataType
}

func (this *Field) Definition() string {
	definition := this.FullType
	if len(this.DefaultValueString) > 0 {
		definition += " DEFAULT '" + this.DefaultValueString + "'"
	}
	if len(this.Comment) > 0 {
		definition += " COMMENT '" + this.Comment + "'"
	}
	return definition
}
