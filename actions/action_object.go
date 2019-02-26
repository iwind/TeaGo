package actions

import (
	"encoding/json"
	"fmt"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/caches"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	"github.com/iwind/TeaGo/utils/string"
	"github.com/pquerna/ffjson/ffjson"
	"net/http"
	"strings"
	"sync"
	"text/template"
)

type ActionObject struct {
	Spec *ActionSpec

	Request        *http.Request
	ResponseWriter http.ResponseWriter
	ParamsMap      Params

	Context *ActionContext

	Module string

	Code    int
	Data    Data
	Message string
	errors  []ActionParamError

	pretty bool // 格式化输出

	SessionManager interface{}
	session        *Session
	sessionLocker  sync.Mutex

	viewDir        string
	viewTemplate   string
	layoutTemplate string
	viewFuncMap    template.FuncMap

	maxSize float64

	Files []*File

	next struct {
		Action string                 `json:"action"`
		Params map[string]interface{} `json:"params"`
		Hash   string                 `json:"hash"`
	}

	writer ActionWriter
}

// 取得内置的动作对象
func (this *ActionObject) Object() *ActionObject {
	return this
}

// 初始化动作
func (this *ActionObject) init() {
	this.Data = map[string]interface{}{}
}

// 设置参数
func (this *ActionObject) SetParam(name, value string) {
	this.ParamsMap[name] = []string{value}
}

// 判断是否有参数
func (this *ActionObject) HasParam(name string) bool {
	values, ok := this.ParamsMap[name]
	return ok && len(values) > 0
}

// 获取参数
func (this *ActionObject) Param(name string) (value string, found bool) {
	values, ok := this.ParamsMap[name]
	if !ok {
		return "", false
	}
	if len(values) == 0 {
		return "", false
	}
	return values[0], true
}

// 获取字符串参数
func (this *ActionObject) ParamString(name string) string {
	v, _ := this.Param(name)
	return v
}

// 获取整型参数
func (this *ActionObject) ParamInt(name string) int {
	v, _ := this.Param(name)
	return types.Int(v)
}

// 获取参数数组
func (this *ActionObject) ParamArray(name string) []string {
	values, ok := this.ParamsMap[name]
	if !ok {
		return []string{}
	}
	return values
}

// 获取客户端的地址，有可能会包含端口
func (this *ActionObject) RequestRemoteAddr() string {
	value := this.Request.Header.Get("X-Real-IP")
	if len(value) > 0 {
		return value
	}

	value = this.Request.Header.Get("X-Forwarded-For")
	if len(value) > 0 {
		return value
	}

	value = this.Request.Header.Get("X-Forwarded-Host")
	if len(value) > 0 {
		return value
	}

	return this.Request.RemoteAddr
}

// 获取客户端的地址
func (this *ActionObject) RequestRemoteIP() string {
	addr := this.RequestRemoteAddr()
	if len(addr) == 0 {
		return ""
	}

	index := strings.Index(addr, ":")
	if index >= 0 {
		return addr[:index]
	}
	return addr
}

// 判断URL的Path部分是否包含任一前缀
func (this *ActionObject) HasPrefix(prefix ...string) bool {
	if len(prefix) == 0 {
		return false
	}
	path := this.Request.URL.Path
	for _, p := range prefix {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// 设置头部信息
func (this *ActionObject) AddHeader(name, value string) {
	this.ResponseWriter.Header().Add(name, value)
}

// 获取Header值
func (this *ActionObject) Header(name string) string {
	return this.Request.Header.Get(name)
}

// 设置cookie
func (this *ActionObject) AddCookie(cookie *http.Cookie) {
	http.SetCookie(this.ResponseWriter, cookie)
}

// 设置能接收的最大数据（字节）
func (this *ActionObject) SetMaxSize(maxSize float64) {
	this.maxSize = maxSize
}

// 输出错误信息
func (this *ActionObject) Error(error string, code int) {
	http.Error(this.ResponseWriter, error, code)
}

// 输出内容
func (this *ActionObject) WriteString(output ...string) {
	for _, outputArg := range output {
		this.Write([]byte(outputArg))
	}
}

// 输出二进制字节
func (this *ActionObject) Write(bytes []byte) {
	if this.writer != nil {
		this.writer.Write(bytes)
	} else {
		this.ResponseWriter.Write(bytes)
	}
}

// 输出可以格式化的内容
func (this *ActionObject) WriteFormat(format string, args ...interface{}) {
	if len(args) > 0 {
		format = fmt.Sprintf(format, args...)
	}
	this.Write([]byte(format))
}

// 写入JSON
func (this *ActionObject) WriteJSON(value interface{}) {
	this.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	var jsonBytes, err = this.marshal(value)
	if err != nil {
		this.Write([]byte(err.Error()))
		return
	}
	this.Write(jsonBytes)
}

// 成功返回
func (this *ActionObject) Success(message ...string) {
	if len(message) > 0 {
		this.Message = message[0]
	}

	this.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	var code = this.Code
	if code == 0 {
		code = 200
	}
	respJson := JSON{
		"code":    code,
		"message": this.Message,
		"data":    this.Data,
	}
	if len(this.next.Action) > 0 {
		respJson["next"] = this.next
	}
	var jsonBytes, err = this.marshal(respJson)
	if err != nil {
		jsonBytes, err = this.marshal(JSON{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
			"data":    nil,
		})
		if err != nil {
			this.Write([]byte(err.Error()))

			panic(this)
			return
		}
		this.Write(jsonBytes)

		panic(this)
		return
	}

	this.Write(jsonBytes)

	panic(this)
}

// 失败返回
func (this *ActionObject) Fail(message ...string) {
	if len(message) > 0 {
		this.Message = strings.Join(message, "")
	}

	this.failWithoutPanic()
	panic(this)
}

// 字段错误提示
func (this *ActionObject) FailField(field string, message ...string) {
	panic([]ActionParamError{{
		Param:    field,
		Messages: message,
	}})
}

// 不使用panic的返回，仅供内部使用
func (this *ActionObject) failWithoutPanic() {
	this.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	var code = this.Code
	if code == 0 {
		code = http.StatusBadRequest
	}
	var jsonBytes, err = this.marshal(JSON{
		"code":    code,
		"message": this.Message,
		"data":    this.Data,
		"errors":  this.errors,
	})
	if err != nil {
		jsonBytes, err = this.marshal(JSON{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
			"data":    nil,
			"errors":  this.errors,
		})
		if err != nil {
			this.Write([]byte(err.Error()))
			panic(this)
			return
		}
		this.Write(jsonBytes)
		panic(this)
		return
	}
	this.Write(jsonBytes)
}

// 设置Session管理器
func (this *ActionObject) SetSessionManager(sessionManager interface{}) {
	this.SessionManager = sessionManager
}

// 读取Session
func (this *ActionObject) Session() *Session {
	if this.session != nil {
		return this.session
	}

	this.sessionLocker.Lock()
	defer this.sessionLocker.Unlock()

	if this.session != nil {
		return this.session
	}

	if this.SessionManager == nil {
		return nil
	}

	var cookie, err = this.Request.Cookie("sid")
	var sid string
	if err != nil || cookie == nil || len(cookie.Value) != 32 {
		sid = stringutil.Rand(32)
		cookie = &http.Cookie{
			Name:  "sid",
			Value: sid,
			Path:  "/",
		}

		//@TODO 可以根据配置或者方法设置cookie超时时间

		this.AddCookie(cookie)
	} else {
		sid = cookie.Value
	}

	var session = &Session{
		Sid:     sid,
		Manager: this.SessionManager,
	}

	this.session = session
	return session
}

// 设置模板目录
func (this *ActionObject) ViewDir(viewDir string) {
	this.viewDir = viewDir
}

// 设置模板文件
func (this *ActionObject) View(viewTemplate string) {
	this.viewTemplate = viewTemplate
}

// 设置模板文件中可以使用的自定义函数
func (this *ActionObject) ViewFunc(name string, f interface{}) {
	this.viewFuncMap[name] = f
}

// 显示模板
func (this *ActionObject) Show() {
	this.AddHeader("Content-Type", "text/html; charset=utf-8")

	err := this.render(Tea.ViewsDir() + "/" + this.viewDir)
	if err != nil {
		logs.Errorf("%s", err.Error())
		this.Error(err.Error(), 500)
	}
}

// 取得单个文件
func (this *ActionObject) File(field string) *File {
	for _, file := range this.Files {
		if file.Field == field {
			return file
		}
	}
	return nil
}

// 设置下一个动作
func (this *ActionObject) Next(nextAction string, params map[string]interface{}, hash ...string) *ActionObject {
	this.next.Action = nextAction
	this.next.Params = params
	if len(hash) > 0 {
		this.next.Hash = strings.Join(hash, "&")
	}
	return this
}

// 设置刷新
func (this *ActionObject) Refresh() *ActionObject {
	this.next.Action = "*refresh"
	return this
}

// 跳转
func (this *ActionObject) RedirectURL(url string) {
	http.Redirect(this.ResponseWriter, this.Request, url, http.StatusTemporaryRedirect)
}

// 缓存
func (this *ActionObject) Cache() *caches.Factory {
	return this.Spec.Cache()
}

// 输出格式化后的JSON
func (this *ActionObject) Pretty() *ActionObject {
	this.pretty = true
	return this
}

// marshal json
func (this *ActionObject) marshal(value interface{}) ([]byte, error) {
	if this.pretty {
		return json.MarshalIndent(value, "", "   ")
	}
	return ffjson.Marshal(value)
}
