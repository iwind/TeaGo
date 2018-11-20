package actions

import (
	"github.com/iwind/TeaGo/maps"
	"io"
	"text/template"
)

// 模板定义
type Template struct {
	native *template.Template
	vars   maps.Map
}

// 创建新模板
func NewTemplate(name string) *Template {
	return &Template{
		native: template.New(name),
		vars:   maps.NewMap(),
	}
}

// 设置分隔符
func (this *Template) Delims(left, right string) *Template {
	this.native.Delims(left, right)
	return this
}

// 设置函数
func (this *Template) Funcs(funcMap template.FuncMap) *Template {
	this.native.Funcs(funcMap)
	return this
}

// 分析文本
func (this *Template) Parse(text string) (*Template, error) {
	_, err := this.native.Parse(text)
	return this, err
}

// 执行模板
func (this *Template) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	return this.native.ExecuteTemplate(wr, name, data)
}

// 执行模板
func (this *Template) Execute(wr io.Writer, data interface{}) error {
	return this.native.Execute(wr, data)
}

// 获取子模板
func (this *Template) NewChild(name string) *Template {
	childTemplate := this.native.New(name)
	return &Template{
		native: childTemplate,
		vars:   maps.NewMap(),
	}
}

// 设置变量
func (this *Template) SetVars(vars maps.Map) *Template {
	for name, value := range vars {
		oldValue, found := this.vars[name]
		if !found {
			this.vars[name] = value
			continue
		}
		oldValueString, ok := oldValue.(string)
		if !ok {
			this.vars[name] = value
			continue
		}

		valueString, ok := value.(string)
		if !ok {
			this.vars[name] = value
			continue
		}

		this.vars[name] = oldValueString + valueString
	}
	return this
}

// 添加变量
func (this *Template) AddVar(varName, value string) *Template {
	this.vars[varName] = value
	return this
}

// 取得变量值
func (this *Template) VarValue(varName string) string {
	return this.vars.GetString(varName)
}

// 判断是否有某个变量
func (this *Template) HasVar(varName string) bool {
	return this.vars.Has(varName)
}
