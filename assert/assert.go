package assert

import (
	"bytes"
	"fmt"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// 断言定义
type Assertion struct {
	t         *testing.T
	beginTime time.Time
	quiet     bool // 是否为静默模式，此模式下的测试通过的项不会提示
	isPassed  bool
}

// 取得一个新的断言
func NewAssertion(t *testing.T) *Assertion {
	return &Assertion{
		t:         t,
		beginTime: time.Now(),
		quiet:     true,
	}
}

// 是否开启静默模式，在此模式下成功的测试不会有提示
func (this *Assertion) Quiet(isQuiet ...bool) *Assertion {
	if len(isQuiet) == 0 {
		this.quiet = true
	} else {
		this.quiet = isQuiet[0]
	}
	return this
}

// 检查是否为true
func (this *Assertion) IsTrue(value bool, msg ...interface{}) *Assertion {
	if value {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否为false
func (this *Assertion) IsFalse(value bool, msg ...interface{}) *Assertion {
	if !value {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否为nil
func (this *Assertion) IsNil(value interface{}, msg ...interface{}) *Assertion {
	if value == nil || reflect.ValueOf(value).IsNil() {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否为非nil
func (this *Assertion) IsNotNil(value interface{}, msg ...interface{}) *Assertion {
	if value != nil && !reflect.ValueOf(value).IsNil() {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否为非error
func (this *Assertion) IsNotError(value interface{}, msg ...interface{}) *Assertion {
	if value == nil {
		this.Pass(msg...)
		return this
	}

	_, ok := value.(error)
	if ok {
		this.Fail(msg...)
	} else {
		this.Pass(msg...)
	}

	return this
}

// 检查是否为非空
func (this *Assertion) IsNotEmpty(value interface{}, msg ...interface{}) *Assertion {
	if value == nil {
		this.Fail(msg...)
		return this
	}

	s, ok := value.(string)
	if ok {
		if len(s) > 0 {
			this.Pass(msg...)
		} else {
			this.Fail(msg...)
		}
		return this
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Map {
		if v.Len() > 0 {
			this.Pass(msg...)
		} else {
			this.Fail(msg...)
		}
		return this
	}

	this.Pass(msg...)

	return this
}

// 检查是否为数字
func (this *Assertion) IsNumber(value interface{}, msg ...interface{}) *Assertion {
	if types.IsNumber(value) {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否为非数字
func (this *Assertion) IsNaN(value interface{}, msg ...interface{}) *Assertion {
	if !types.IsNumber(value) {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否相等
func (this *Assertion) Equals(value1 interface{}, value2 interface{}, msg ...interface{}) *Assertion {
	if value1 == value2 {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}
	return this
}

// 检查是否不相等
func (this *Assertion) NotEquals(value1 interface{}, value2 interface{}, msg ...interface{}) *Assertion {
	if value1 != value2 {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}
	return this
}

// 检查是否包含某个条目，目前只支持slice
func (this *Assertion) Contains(container interface{}, item interface{}, msg ...interface{}) *Assertion {
	if container == nil {
		this.Fail("'container' should not be nil")
		return this
	}

	value := reflect.ValueOf(container)
	kind := value.Kind()
	if kind == reflect.Slice || kind == reflect.Array {
		size := value.Len()
		for i := 0; i < size; i++ {
			item1 := value.Index(i).Interface()
			if item1 == item {
				this.Pass(msg...)
				return this
			}
		}

		this.Fail(msg...)
	} else if kind == reflect.Map {
		for _, keyValue := range value.MapKeys() {
			item1 := value.MapIndex(keyValue).Interface()
			if item1 == item {
				this.Pass(msg...)
				return this
			}
		}

		this.Fail(msg...)
	} else {
		this.Fail("'container' should be slice, array or map")
	}

	return this
}

// 执行正则匹配
func (this *Assertion) Match(pattern string, value string, msg ...interface{}) *Assertion {
	reg, err := regexp.Compile(pattern)
	if err != nil {
		this.Fail(err.Error())
	} else if reg.MatchString(value) {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}
	return this
}

// 检查类型
func (this *Assertion) IsKind(value interface{}, kind reflect.Kind, msg ...interface{}) *Assertion {
	v := reflect.TypeOf(value)
	if v == nil {
		this.Fail(msg...)
		return this
	}

	if v.Kind() == kind {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否为bool类型
func (this *Assertion) IsBool(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Bool, msg...)
}

// 检查是否为int类型
func (this *Assertion) IsInt(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Int, msg...)
}

// 检查是否为int8类型
func (this *Assertion) IsInt8(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Int8, msg...)
}

// 检查是否为int16类型
func (this *Assertion) IsInt16(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Int16, msg...)
}

// 检查是否为int32类型
func (this *Assertion) IsInt32(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Int32, msg...)
}

// 检查是否为int64类型
func (this *Assertion) IsInt64(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Int64, msg...)
}

// 检查是否为uint类型
func (this *Assertion) IsUint(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Uint, msg...)
}

// 检查是否为uint8类型
func (this *Assertion) IsUint8(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Uint8, msg...)
}

// 检查是否为uint16类型
func (this *Assertion) IsUint16(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Uint16, msg...)
}

// 检查是否为uint32类型
func (this *Assertion) IsUint32(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Uint32, msg...)
}

// 检查是否为uint64类型
func (this *Assertion) IsUint64(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Uint64, msg...)
}

// 检查是否为float32类型
func (this *Assertion) IsFloat32(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Float32, msg...)
}

// 检查是否为float64类型
func (this *Assertion) IsFloat64(value interface{}, msg ...interface{}) *Assertion {
	return this.IsKind(value, reflect.Float64, msg...)
}

// 检查是否为整数类型（int, int8, ...）
func (this *Assertion) IsInteger(value interface{}, msg ...interface{}) *Assertion {
	if types.IsInteger(value) {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否为浮点数类型（float32, float64）
func (this *Assertion) IsFloat(value interface{}, msg ...interface{}) *Assertion {
	if types.IsFloat(value) {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否大于某个数字
func (this *Assertion) Gt(value interface{}, compare interface{}, msg ...interface{}) *Assertion {
	if !types.IsNumber(value) {
		this.Fail("'value' should be a number")
		return this
	}

	if !types.IsNumber(compare) {
		this.Fail("'compare' should be a number")
		return this
	}

	if types.Float64(value) > types.Float64(compare) {
		this.Pass(msg...)
		return this
	}

	this.Fail(msg...)

	return this
}

// 检查是否大于等于某个数字
func (this *Assertion) Gte(value interface{}, compare interface{}, msg ...interface{}) *Assertion {
	if !types.IsNumber(value) {
		this.Fail("'value' should be a number")
		return this
	}

	if !types.IsNumber(compare) {
		this.Fail("'compare' should be a number")
		return this
	}

	if types.Float64(value) >= types.Float64(compare) {
		this.Pass(msg...)
		return this
	}

	this.Fail(msg...)

	return this
}

// 检查是否小于某个数字
func (this *Assertion) Lt(value interface{}, compare interface{}, msg ...interface{}) *Assertion {
	if !types.IsNumber(value) {
		this.Fail("'value' should be a number")
		return this
	}

	if !types.IsNumber(compare) {
		this.Fail("'compare' should be a number")
		return this
	}

	if types.Float64(value) < types.Float64(compare) {
		this.Pass(msg...)
		return this
	}

	this.Fail(msg...)

	return this
}

// 检查是否小于等于某个数字
func (this *Assertion) Lte(value interface{}, compare interface{}, msg ...interface{}) *Assertion {
	if !types.IsNumber(value) {
		this.Fail("'value' should be a number")
		return this
	}

	if !types.IsNumber(compare) {
		this.Fail("'compare' should be a number")
		return this
	}

	if types.Float64(value) <= types.Float64(compare) {
		this.Pass(msg...)
		return this
	}

	this.Fail(msg...)

	return this
}

// 检查是否在两个数字之间
func (this *Assertion) Between(value interface{}, min interface{}, max interface{}, msg ...interface{}) *Assertion {
	if !types.IsNumber(value) {
		this.Fail("'value' should be a number")
		return this
	}

	if !types.IsNumber(min) {
		this.Fail("'min' should be a number")
		return this
	}

	if !types.IsNumber(max) {
		this.Fail("'max' should be a number")
		return this
	}

	minFloat := types.Float64(min)
	maxFloat := types.Float64(max)
	if minFloat > maxFloat {
		this.Fail("'min' should not be greater than 'max'")
		return this
	}

	valueFloat := types.Float64(value)
	if valueFloat >= minFloat && valueFloat <= maxFloat {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否为字符串
func (this *Assertion) IsString(value interface{}, msg ...interface{}) *Assertion {
	_, ok := value.(string)
	if ok {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}
	return this
}

// 检查是否为map
func (this *Assertion) IsMap(value interface{}, msg ...interface{}) *Assertion {
	if value == nil {
		this.Fail(msg...)
		return this
	}

	if reflect.TypeOf(value).Kind() == reflect.Map {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 检查是否为slice
func (this *Assertion) IsSlice(value interface{}, msg ...interface{}) *Assertion {
	if value == nil {
		this.Fail(msg...)
		return this
	}

	if reflect.TypeOf(value).Kind() == reflect.Slice {
		this.Pass(msg...)
	} else {
		this.Fail(msg...)
	}

	return this
}

// 成功后执行
func (this *Assertion) Then(f func()) *Assertion {
	if f != nil {
		f()
	}
	return this
}

// 检查是否有panic
func (this *Assertion) Panic(f func(), msg ...interface{}) *Assertion {
	defer func() {
		err := recover()
		if err != nil {
			this.Pass(msg...)
		} else {
			this.Fail(msg...)
		}
	}()

	f()
	return this
}

// 检查是否没有panic
func (this *Assertion) NoPanic(f func(), msg ...interface{}) *Assertion {
	defer func() {
		err := recover()
		if err == nil {
			this.Pass(msg...)
		} else {
			this.Fail(msg...)
		}
	}()

	f()
	return this
}

// 检查执行时间是否超时
func (this *Assertion) NotTimeout(duration time.Duration, f func(), msg ...interface{}) *Assertion {
	before := time.Now()
	f()
	cost := time.Since(before)

	if len(msg) == 0 {
		msg = []interface{}{fmt.Sprintf("cost:%.6f seconds", cost.Seconds())}
	}

	if cost > duration {
		this.Fail(msg...)
	} else {
		this.Pass(msg...)
	}
	return this
}

// 输出日志
func (this *Assertion) Log(msg ...interface{}) *Assertion {
	this.output("", msg...)

	return this
}

// 格式化输出日志
func (this *Assertion) Logf(format string, args ...interface{}) *Assertion {
	this.Log(fmt.Sprintf(format, args...))

	return this
}

// 输出JSON格式的日志
func (this *Assertion) LogJSON(msg ...interface{}) *Assertion {
	for _, m := range msg {
		this.t.Log(stringutil.JSONEncodePretty(m))
	}
	return this
}

// 返回失败
func (this *Assertion) Fail(msg ...interface{}) *Assertion {
	this.output("[FAIL]", msg...)
	this.isPassed = false
	return this
}

// 返回致命错误
func (this *Assertion) Fatal(msg ...interface{}) *Assertion {
	this.output("[FATAL]", msg...)
	this.isPassed = false
	return this
}

// 返回成功
func (this *Assertion) Pass(msg ...interface{}) *Assertion {
	this.output("[PASS]", msg...)
	this.isPassed = true
	return this
}

// 打印测试花销(ms）
func (this *Assertion) Cost() *Assertion {
	this.t.Logf("cost:%.6f %s", time.Since(this.beginTime).Seconds()*1000, "ms")
	return this
}

func (this *Assertion) output(tag string, msg ...interface{}) *Assertion {
	var filename string
	var lineNo int

	_, currentFilename, _, currentOk := runtime.Caller(0)
	if currentOk {
		for i := 1; i < 32; i++ {
			_, filename1, lineNo1, ok := runtime.Caller(i)
			if !ok {
				break
			}

			if filename1 == currentFilename {
				continue
			}

			filename = filename1
			lineNo = lineNo1

			break
		}
	}

	// source
	source := ""
	if len(filename) > 0 {
		data, err := ioutil.ReadFile(filename)
		if err == nil {
			lines := bytes.Split(data, []byte("\n"))
			if lineNo-1 < len(lines) {
				source = strings.TrimSpace(string(lines[lineNo-1]))

				if len(source) > 80 {
					source = source[:80] + " ..."
				}
			}
		}
	}

	goPath := os.Getenv("GOPATH")
	goAbsPath, err := filepath.Abs(goPath)
	if err == nil {
		goPath = goAbsPath
		filename = strings.TrimPrefix(filename, goAbsPath)[1:]
	}

	if len(msg) > 0 {
		msgStrings := []string{}
		for _, msgItem := range msg {
			if tag == "[FAIL]" {
				f, ok := msgItem.(func() string)
				if ok {
					msgItem = f()
				}
			}
			msgStrings = append(msgStrings, types.String(msgItem))
		}

		output := tag + strings.Join(msgStrings, " ") + "\n"
		if len(source) > 0 {
			output += "| " + source + "\n"
		}
		output += "| " + filename + ":" + fmt.Sprintf("%d", lineNo)
		msg = []interface{}{output}
	} else {
		output := tag + "\n"
		if len(source) > 0 {
			output += "| " + source + "\n"
		}
		output += "| " + filename + ":" + fmt.Sprintf("%d", lineNo)
		msg = []interface{}{output}
	}

	if tag == "[PASS]" {
		if !this.quiet {
			this.t.Log(msg...)
		}
	} else if tag == "[FAIL]" {
		this.t.Log(msg...)
		this.t.Fail()
	} else if tag == "[FATAL]" {
		this.t.Fatal(msg...)
	} else {
		this.t.Log(msg...)
	}

	return this
}
