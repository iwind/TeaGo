package actions

import (
	"reflect"
	"net/http"
	"strings"
	"strconv"
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/iwind/TeaGo/utils/string"
	"github.com/iwind/TeaGo/logs"
	"text/template"
	"runtime/debug"
	"github.com/iwind/TeaGo/Tea"
	"path/filepath"
	"github.com/iwind/TeaGo/types"
	"github.com/iwind/TeaGo/lists"
)

type Params map[string][]string

type ActionWriter interface {
	Write([]byte) (n int, err error)
}

type ActionObject struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	ParamsMap      Params

	Module string

	Code    int
	Data    Data
	Message string
	errors  []ActionParamError

	SessionManager interface{}
	session        *Session

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

type Action struct {
	ActionObject
}

type ActionParamError struct {
	Param    string   `json:"param"`
	Messages []string `json:"messages"`
}

// 读取ActionObject对象的接口
type ActionWrapper interface {
	Object() *ActionObject
}

// 取得内置的动作对象
func (this *ActionObject) Object() *ActionObject {
	return this
}

// 初始化动作
func (this *ActionObject) Init() {
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

// 设置头部信息
func (this *ActionObject) Header(name, value string) {
	this.ResponseWriter.Header().Add(name, value)
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
func (this *ActionObject) WriteString(output ... string) {
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
func (this *ActionObject) WriteFormat(format string, args ... interface{}) {
	if len(args) > 0 {
		format = fmt.Sprintf(format, args ...)
	}
	this.Write([]byte(format))
}

// 写入JSON
func (this *ActionObject) WriteJSON(value interface{}) {
	this.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	var jsonBytes, err = ffjson.Marshal(value)
	if err != nil {
		this.Write([]byte(err.Error()))
		return
	}
	this.Write(jsonBytes)
}

// 成功返回
func (this *ActionObject) Success() {
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
	var jsonBytes, err = ffjson.Marshal(respJson)
	if err != nil {
		jsonBytes, err = ffjson.Marshal(JSON{
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
func (this *ActionObject) Fail() {
	this.failWithoutPanic()
	panic(this)
}

// 不使用panic的返回，仅供内部使用
func (this *ActionObject) failWithoutPanic() {
	this.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	var code = this.Code
	if code == 0 {
		code = http.StatusBadRequest
	}
	var jsonBytes, err = ffjson.Marshal(JSON{
		"code":    code,
		"message": this.Message,
		"data":    this.Data,
		"errors":  this.errors,
	})
	if err != nil {
		jsonBytes, err = ffjson.Marshal(JSON{
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

// 带有失败消息的返回
func (this *ActionObject) FailMessage(code int, message string) {
	this.Code = code
	this.Message = message
	this.Fail()
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
	this.Header("Content-Type", "text/html; charset=utf-8")

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
func (this *ActionObject) Next(nextAction string, params map[string]interface{}, hash string) *ActionObject {
	this.next.Action = nextAction
	this.next.Params = params
	this.next.Hash = hash
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

// 执行某个Action
func RunAction(actionPtr interface{},
	spec *ActionSpec,
	request *http.Request,
	responseWriter http.ResponseWriter,
	params Params) interface{} {
	// 运行
	action := actionPtr.(ActionWrapper).Object()
	runActionCopy(spec, request, responseWriter, params, action.SessionManager, action.maxSize)

	return actionPtr
}

// 执行Action副本（为了防止同一个Action多次调用会相互影响）
func runActionCopy(spec *ActionSpec, request *http.Request,
	responseWriter http.ResponseWriter, params Params, sessionManager interface{}, maxSize float64) {
	var actionPtrValue = spec.NewPtrValue()
	var actionObject = actionPtrValue.Interface().(ActionWrapper).Object()
	var afterFuncs = []func(){}

	// 执行helper.AfterAction()
	defer func() {
		if len(afterFuncs) > 0 {
			for i := len(afterFuncs) - 1; i >= 0; i -- {
				afterFuncs[i]()
			}
		}
	}()

	// 执行After()
	if spec.AfterFunc != nil {
		defer spec.AfterFunc.Call([]reflect.Value{actionPtrValue})
	}

	// 捕获message
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		if _, ok := err.(*ActionObject); ok {
			return
		}

		// 如果有错误信息
		if errors, ok := err.([]ActionParamError); ok {
			actionObject.errors = errors
			actionObject.failWithoutPanic()
			return
		}

		// 如果是字符串
		message, ok := err.(string)
		if ok {
			actionObject.Message = message
			actionObject.failWithoutPanic()
		} else {
			errObject, ok := err.(error)
			if ok {
				logs.Errorf("%s\n~~~\n%s", errObject.Error(), string(debug.Stack()))
			}
		}
	}()

	// 设置模板
	pkgPath := spec.PkgPath
	className := spec.ClassName
	actionPkg := pkgPath[strings.LastIndex(pkgPath, "/actions/")+len("/actions/"):]
	actionClass := strings.TrimSuffix(className[strings.LastIndex(className, ".")+1:], "Action")

	separatorIndex := strings.Index(actionPkg, "/")

	// 如果没有指定模块，则认为@default模块
	if separatorIndex == -1 {
		actionPkg = "@default/" + actionPkg
		separatorIndex = strings.Index(actionPkg, "/")
	}

	actionObject.ViewDir(actionPkg[:separatorIndex])
	actionObject.View(actionPkg[separatorIndex+1:] + "/" + strings.ToLower(actionClass[0:1]) + actionClass[1:])

	// 设置变量
	actionObject.Module = spec.Module
	actionObject.Request = request
	if responseWriter != nil {
		actionObject.ResponseWriter = responseWriter
	}
	actionObject.ParamsMap = params
	actionObject.viewFuncMap = template.FuncMap{}

	// 设置Session
	actionObject.SessionManager = sessionManager

	// 设置最大文件上传尺寸
	actionObject.maxSize = maxSize

	// 读取上传的文件
	actionObject.Files = []*File{}
	if request.MultipartForm != nil {
		parseRequestFiles(actionObject)
	}

	// 初始化
	actionObject.Init()

	// 执行
	var requestRun = "Run" + strings.ToUpper(string(request.Method[0])) + strings.ToLower(string(request.Method[1:]))
	runFuncValue, found := spec.Funcs[requestRun]
	if !found {
		runFuncValue, found = spec.Funcs["Run"]

		if !found {
			logs.Errorf("'Action.Run()' or 'Action." + requestRun + "()' method should be implemented in '" + spec.Type.Name() + "' (at " + spec.Type.PkgPath() + "/" + spec.Type.Name() + ")")
			if responseWriter != nil {
				http.Error(responseWriter.(http.ResponseWriter), "500 Internal Error", http.StatusInternalServerError)
			}
			return
		}
	}

	var runMethodType = runFuncValue.Type()
	if runMethodType.NumIn() == 1 {
		runFuncValue.Call([]reflect.Value{actionPtrValue})
		return
	}
	if runMethodType.NumIn() > 2 {
		logs.Errorf("Action.Run() method should contains only one argument")
		if responseWriter != nil {
			http.Error(responseWriter.(http.ResponseWriter), "500 Internal Error", http.StatusInternalServerError)
		}
		return
	}

	var argType = runMethodType.In(1)

	if argType.Kind() != reflect.Struct {
		logs.Errorf("Action.Run() method should contains only struct argument")
		if responseWriter != nil {
			http.Error(responseWriter.(http.ResponseWriter), "500 Internal Error", http.StatusInternalServerError)
		}
		return
	}
	var argValue = reflect.Indirect(reflect.New(argType))
	var countFields = argValue.NumField()
	for i := 0; i < countFields; i ++ {
		var field = argType.Field(i)
		var fieldName = field.Name
		if len(fieldName) == 0 {
			continue
		}

		var fieldValue = argValue.Field(i)
		if !fieldValue.CanSet() {
			logs.Errorf("Action.Run(): field value '" + field.Name + "' can not be accessed")
			continue
		}

		// 初始化特殊类型的参数
		switch fieldValue.Type().String() {
		case "*actions.File": // 支持文件指针
			bindName, ok := field.Tag.Lookup("field")
			if ok {
				fieldValue.Set(reflect.ValueOf(actionObject.File(bindName)))
			} else {
				filePtr := actionObject.File(fieldName)
				if filePtr == nil {
					var lowerFirstName = strings.ToLower(string(fieldName[0])) + string(fieldName[1:])
					filePtr = actionObject.File(lowerFirstName)
				}
				fieldValue.Set(reflect.ValueOf(filePtr))
			}
			continue
		case "actions.File": // 支持文件
			bindName, ok := field.Tag.Lookup("field")
			if ok {
				filePtr := actionObject.File(bindName)
				if filePtr != nil {
					fieldValue.Set(reflect.ValueOf(*filePtr))
				}
			} else {
				filePtr := actionObject.File(fieldName)
				if filePtr == nil {
					var lowerFirstName = strings.ToLower(string(fieldName[0])) + string(fieldName[1:])
					filePtr = actionObject.File(lowerFirstName)
				}
				if filePtr != nil {
					fieldValue.Set(reflect.ValueOf(*filePtr))
				}
			}
			continue
		}

		// 执行方法 Helper 方法
		if fieldValue.NumMethod() > 0 {
			// 初始化
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}

			var helperMethod = fieldValue.MethodByName("BeforeAction")
			if helperMethod.IsValid() {
				// 执行 BeforeAction() 方法
				goNext := runHelperMethodBefore(helperMethod, actionPtrValue, field, fieldValue)
				if !goNext {
					// 自动调用结束 AfterAction() 方法
					afterMethod := fieldValue.MethodByName("AfterAction")
					if afterMethod.IsValid() {
						afterFuncs = append(afterFuncs, func() {
							runHelperMethodAfter(afterMethod, actionPtrValue, field)
						})
					}

					return
				}

				// 自动调用结束 AfterAction() 方法
				afterMethod := fieldValue.MethodByName("AfterAction")
				if afterMethod.IsValid() {
					if afterMethod.IsValid() {
						afterFuncs = append(afterFuncs, func() {
							runHelperMethodAfter(afterMethod, actionPtrValue, field)
						})
					}
				}
			}
		}

		// alias:"别名"
		bindName, ok := field.Tag.Lookup("alias")
		if ok && len(bindName) > 0 {
			fieldName = bindName
		}

		// cookie:"Cookie参数名"
		var fieldParamValue string
		var hasValue = false

		cookieName, ok := field.Tag.Lookup("cookie")
		var hasTagValue = false
		if ok && len(cookieName) > 0 {
			hasTagValue = true
			cookieValue, err := request.Cookie(cookieName)
			if err != nil {
				hasValue = false
			} else {
				fieldParamValue = strings.TrimSpace(cookieValue.Value)
				hasValue = len(fieldParamValue) > 0
			}
		}

		// session:"Session参数名"
		sessionName, ok := field.Tag.Lookup("session")
		if ok && len(sessionName) > 0 {
			hasTagValue = true

			session := actionPtrValue.Interface().(ActionWrapper).Object().Session()
			if session != nil {
				sessionValue := session.StringValue(sessionName)
				fieldParamValue = strings.TrimSpace(sessionValue)
				hasValue = len(fieldParamValue) > 0
			}
		}

		// 从request参数中读取
		if !hasTagValue {
			fieldParamValue, hasValue = getActionParamFuzzy(&params, fieldName)
		}

		// default:"默认值"
		// 注意：DefaultValue不对字符串进行Trim()处理
		if !hasValue || len(fieldParamValue) == 0 {
			defaultValue, ok := field.Tag.Lookup("default")
			if ok {
				hasValue = true
				fieldParamValue = defaultValue
			}
		}

		if hasValue {
			switch field.Type.Kind() {
			case reflect.Int:
				fieldValue.Set(reflect.ValueOf(types.Int(fieldParamValue)))
			case reflect.Int8:
				fieldValue.Set(reflect.ValueOf(types.Int8(fieldParamValue)))
			case reflect.Int16:
				fieldValue.Set(reflect.ValueOf(types.Int16(fieldParamValue)))
			case reflect.Int32:
				fieldValue.Set(reflect.ValueOf(types.Int32(fieldParamValue)))
			case reflect.Int64:
				fieldValue.Set(reflect.ValueOf(types.Int64(fieldParamValue)))
			case reflect.Uint:
				fieldValue.Set(reflect.ValueOf(types.Uint(fieldParamValue)))
			case reflect.Uint8:
				fieldValue.Set(reflect.ValueOf(types.Uint8(fieldParamValue)))
			case reflect.Uint16:
				fieldValue.Set(reflect.ValueOf(types.Uint16(fieldParamValue)))
			case reflect.Uint32:
				fieldValue.Set(reflect.ValueOf(types.Uint32(fieldParamValue)))
			case reflect.Uint64:
				fieldValue.Set(reflect.ValueOf(types.Uint64(fieldParamValue)))
			case reflect.Bool:
				if lists.Contains([]string{"on", "true", "yes", "enabled"}, fieldParamValue) {
					fieldValue.SetBool(true)
				} else {
					fieldValue.SetBool(types.Bool(fieldParamValue))
				}
			case reflect.String:
				fieldValue.Set(reflect.ValueOf(fieldParamValue))
			case reflect.Float32:
				fieldValue.Set(reflect.ValueOf(types.Float32(fieldParamValue)))
			case reflect.Float64:
				fieldValue.Set(reflect.ValueOf(types.Float64(fieldParamValue)))
			case reflect.Slice:
				if field.Type.Elem().Kind() == reflect.Uint8 {
					fieldValue.SetBytes([]byte(fieldParamValue))
				}
			}
		}
	}

	// Before
	if spec.BeforeFunc != nil {
		returnValues := spec.BeforeFunc.Call([]reflect.Value{actionPtrValue})
		if len(returnValues) > 0 {
			result := returnValues[0]

			if result.Interface() != nil && result.Type().Kind() == reflect.Bool && !result.Bool() {
				return
			}
		}
	}

	// 执行Run
	runFuncValue.Call([]reflect.Value{actionPtrValue, argValue})
}

// 调用参数中的 Helper BeforeAction
func runHelperMethodBefore(method reflect.Value, actionPtr reflect.Value, field reflect.StructField, fieldValue reflect.Value) (goNext bool) {
	paramName := field.Name

	// 设置属性
	if len(field.Tag) > 0 {
		tags := parseTagsFromString(string(field.Tag))
		for name, value := range tags {
			if len(name) < 1 {
				continue
			}

			newName := ""
			pieces := strings.Split(name, "_")
			for _, piece := range pieces {
				newPiece := strings.ToUpper(piece[0:1])
				if len(piece) > 1 {
					newPiece += piece[1:]
				}
				newName += newPiece
			}

			helperPtrFieldValue := reflect.Indirect(fieldValue).FieldByName(newName)

			if !helperPtrFieldValue.IsValid() || !helperPtrFieldValue.CanSet() {
				continue
			}
			switch helperPtrFieldValue.Type().Kind() {
			case reflect.Int:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Int(value)))
			case reflect.Int8:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Int8(value)))
			case reflect.Int16:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Int16(value)))
			case reflect.Int32:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Int32(value)))
			case reflect.Int64:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Int64(value)))
			case reflect.Uint:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Uint(value)))
			case reflect.Uint8:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Uint8(value)))
			case reflect.Uint16:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Uint16(value)))
			case reflect.Uint32:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Uint32(value)))
			case reflect.Uint64:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Uint64(value)))
			case reflect.String:
				helperPtrFieldValue.Set(reflect.ValueOf(types.String(value)))
			case reflect.Bool:
				helperPtrFieldValue.Set(reflect.ValueOf(types.Bool(value)))
			}
		}
	}

	methodType := method.Type()
	returnValues := []reflect.Value{}
	numIn := methodType.NumIn()
	if numIn == 0 {
		returnValues = method.Call([]reflect.Value{})
	} else if numIn == 1 {
		actionIn := methodType.In(0)

		if actionIn != reflect.TypeOf((*ActionWrapper)(nil)).Elem() {
			var actionInString = actionIn.String()
			if actionInString == "actions.ActionObject" {
				returnValues = method.Call([]reflect.Value{reflect.Indirect(reflect.ValueOf(actionPtr.Interface().(ActionWrapper).Object()))})
			} else if actionInString == "*actions.ActionObject" {
				returnValues = method.Call([]reflect.Value{reflect.ValueOf(actionPtr.Interface().(ActionWrapper).Object())})
			} else {
				logs.Errorf("[Action]invalid BeforeAction() method, should be 'BeforeAction(actionPtr ActionWrapper) (goNext bool)'")
				return true
			}
		} else {
			returnValues = method.Call([]reflect.Value{actionPtr})
		}
	} else if numIn == 2 {
		actionIn := methodType.In(0)
		firstArg := reflect.Value{}
		if actionIn != reflect.TypeOf((*ActionWrapper)(nil)).Elem() {
			var actionInString = actionIn.String()
			if actionInString == "actions.ActionObject" {
				firstArg = reflect.Indirect(reflect.ValueOf(actionPtr.Interface().(ActionWrapper).Object()))
			} else if actionInString == "*actions.ActionObject" {
				firstArg = reflect.ValueOf(actionPtr.Interface().(ActionWrapper).Object())
			} else {
				logs.Errorf("[Action]invalid BeforeAction() method, should be 'BeforeAction(actionPtr ActionWrapper) (goNext bool)'")
				return true
			}
		} else {
			firstArg = actionPtr
		}

		paramIn := methodType.In(1)
		if paramIn.Kind() != reflect.String {
			logs.Errorf("[Action]invalid BeforeAction() method, should be 'BeforeAction(actionPtr ActionWrapper) (goNext bool)'")
			return true
		}

		returnValues = method.Call([]reflect.Value{firstArg, reflect.ValueOf(paramName)})
	} else {
		logs.Errorf("[Action]invalid BeforeAction() method, should be 'BeforeAction(actionPtr ActionWrapper) (goNext bool)'")
		return true
	}

	if len(returnValues) > 0 {
		result := returnValues[0]

		if !result.IsValid() {
			return true
		}

		if result.Interface() == nil {
			return true
		}

		if result.Type().Kind() == reflect.Bool {
			return result.Bool()
		}
	}

	return true
}

// 调用参数中的 Helper BeforeAction
func runHelperMethodAfter(method reflect.Value, actionPtr reflect.Value, field reflect.StructField) (goNext bool) {
	paramName := field.Name

	methodType := method.Type()
	returnValues := []reflect.Value{}
	numIn := methodType.NumIn()
	if numIn == 0 {
		returnValues = method.Call([]reflect.Value{})
	} else if numIn == 1 {
		actionIn := methodType.In(0)

		if actionIn != reflect.TypeOf((*ActionWrapper)(nil)).Elem() {
			var actionInString = actionIn.String()
			if actionInString == "actions.ActionObject" {
				returnValues = method.Call([]reflect.Value{reflect.Indirect(reflect.ValueOf(actionPtr.Interface().(ActionWrapper).Object()))})
			} else if actionInString == "*actions.ActionObject" {
				returnValues = method.Call([]reflect.Value{reflect.ValueOf(actionPtr.Interface().(ActionWrapper).Object())})
			} else {
				logs.Errorf("[Action]invalid AfterAction() method, should be 'AfterAction(actionPtr ActionWrapper) (goNext bool)'")
				return true
			}
		} else {
			returnValues = method.Call([]reflect.Value{actionPtr})
		}
	} else if numIn == 2 {
		actionIn := methodType.In(0)
		firstArg := reflect.Value{}
		if actionIn != reflect.TypeOf((*ActionWrapper)(nil)).Elem() {
			var actionInString = actionIn.String()
			if actionInString == "actions.ActionObject" {
				firstArg = reflect.Indirect(reflect.ValueOf(actionPtr.Interface().(ActionWrapper).Object()))
			} else if actionInString == "*actions.ActionObject" {
				firstArg = reflect.ValueOf(actionPtr.Interface().(ActionWrapper).Object())
			} else {
				logs.Errorf("[Action]invalid AfterAction() method, should be 'AfterAction(actionPtr ActionWrapper) (goNext bool)'")
				return true
			}
		} else {
			firstArg = actionPtr
		}

		paramIn := methodType.In(1)
		if paramIn.Kind() != reflect.String {
			logs.Errorf("[Action]invalid BeforeAction() method, should be 'BeforeAction(actionPtr ActionWrapper) (goNext bool)'")
			return true
		}

		returnValues = method.Call([]reflect.Value{firstArg, reflect.ValueOf(paramName)})
	} else {
		logs.Errorf("[Action]invalid BeforeAction() method, should be 'BeforeAction(actionPtr ActionWrapper) (goNext bool)'")
		return true
	}

	if len(returnValues) > 0 {
		result := returnValues[0]

		if !result.IsValid() {
			return true
		}

		if result.Interface() == nil {
			return true
		}

		if result.Type().Kind() == reflect.Bool {
			return result.Bool()
		}
	}

	return true
}

func getActionParam(params *Params, name string) (value string, has bool) {
	values, ok := (*params)[name]
	if !ok {
		return "", false
	}
	if len(values) == 0 {
		return "", false
	}
	return values[0], true
}

func getActionParamFuzzy(params *Params, name string) (value string, has bool) {
	value, hasValue := getActionParam(params, name)
	if hasValue {
		return value, true
	}

	var lowerFirstName = strings.ToLower(string(name[0])) + string(name[1:])
	return getActionParam(params, lowerFirstName)
}

func parseRequestFiles(action *ActionObject) {
	err := action.Request.ParseMultipartForm(128 << 20)
	if err != nil {
		logs.Errorf("%s", err.Error())
		return
	}
	for field, headers := range action.Request.MultipartForm.File {
		for _, header := range headers {
			if err != nil {
				logs.Errorf("%s", err.Error())
				return
			}

			file := &File{}
			file.Field = field
			file.Size = header.Size
			file.Filename = header.Filename
			file.Ext = strings.ToLower(filepath.Ext(header.Filename))
			file.ContentType = header.Header.Get("Content-Type")
			file.OriginFile = header
			action.Files = append(action.Files, file)
		}
	}
}

func parseTagsFromString(tag string) map[string]string {
	tags := map[string]string{}

	// When modifying this code, also update the validateStructTag code
	// in cmd/vet/structtag.go.

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			break
		}
		tags[name] = value
	}
	return tags
}
