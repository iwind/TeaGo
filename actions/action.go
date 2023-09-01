package actions

import (
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

type Params map[string][]string

type ActionWriter interface {
	Write([]byte) (n int, err error)
}

type Action struct {
	ActionObject
}

type ActionParamError struct {
	Param    string   `json:"param"`
	Messages []string `json:"messages"`
}

// ActionWrapper 读取ActionObject对象的接口
type ActionWrapper interface {
	Object() *ActionObject
}

// RunAction 执行某个Action
func RunAction(actionPtr interface{},
	spec *ActionSpec,
	request *http.Request,
	responseWriter http.ResponseWriter,
	params Params,
	helpers []interface{},
	initData Data,
) interface{} {
	// 运行
	action := actionPtr.(ActionWrapper).Object()
	runActionCopy(spec, request, responseWriter, params, action.SessionManager, action.sessionCookieName, action.maxSize, helpers, initData)

	return actionPtr
}

// 执行Action副本（为了防止同一个Action多次调用会相互影响）
func runActionCopy(spec *ActionSpec,
	request *http.Request,
	responseWriter http.ResponseWriter,
	params Params,
	sessionManager interface{},
	sessionCookieName string,
	maxSize float64,
	helpers []interface{},
	initData Data,
) {
	var actionPtrValue = spec.NewPtrValue()
	var actionObject = actionPtrValue.Interface().(ActionWrapper).Object()
	var afterFuncs = []func(){}

	actionObject.Spec = spec

	// 执行helper.AfterAction()
	defer func() {
		if len(afterFuncs) > 0 {
			for i := len(afterFuncs) - 1; i >= 0; i-- {
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
		actionPkg = "default/" + actionPkg
		separatorIndex = strings.Index(actionPkg, "/")
	}

	actionObject.ViewDir("@" + actionPkg[:separatorIndex])
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
	actionObject.sessionCookieName = sessionCookieName

	// 设置最大文件上传尺寸
	actionObject.maxSize = maxSize

	// 读取上传的文件
	actionObject.Files = []*File{}
	if request.MultipartForm != nil {
		parseRequestFiles(actionObject)
	}

	// 初始化
	actionObject.init()
	actionObject.Data = Data{}
	if initData != nil {
		for k, v := range initData {
			actionObject.Data[k] = v
		}
	}

	// 执行Helpers
	for _, helper := range helpers {
		helperValue := reflect.ValueOf(helper)
		if helperValue.IsNil() || !helperValue.IsValid() {
			continue
		}

		field := reflect.StructField{}

		var helperMethod = helperValue.MethodByName("BeforeAction")
		if helperMethod.IsValid() {
			// 执行 BeforeAction() 方法
			goNext := runHelperMethodBefore(helperMethod, actionPtrValue, field, helperValue)
			if !goNext {
				// 自动调用结束 AfterAction() 方法
				afterMethod := helperValue.MethodByName("AfterAction")
				if afterMethod.IsValid() {
					afterFuncs = append(afterFuncs, func() {
						runHelperMethodAfter(afterMethod, actionPtrValue, field)
					})
				}

				return
			}

			// 自动调用结束 AfterAction() 方法
			afterMethod := helperValue.MethodByName("AfterAction")
			if afterMethod.IsValid() {
				if afterMethod.IsValid() {
					afterFuncs = append(afterFuncs, func() {
						runHelperMethodAfter(afterMethod, actionPtrValue, field)
					})
				}
			}
		}
	}

	// 执行Action
	var requestRun = "Run" + strings.ToUpper(string(request.Method[0])) + strings.ToLower(request.Method[1:])
	runFuncValue, found := spec.FuncMap[requestRun]
	if !found {
		runFuncValue, found = spec.FuncMap["Run"]

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
			http.Error(responseWriter, "500 Internal Error", http.StatusInternalServerError)
		}
		return
	}
	var argValue = reflect.Indirect(reflect.New(argType))
	var countFields = argValue.NumField()
	for i := 0; i < countFields; i++ {
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
		switch fieldValue.Interface().(type) {
		case *File: // 支持文件指针
			bindName, ok := field.Tag.Lookup("field")
			if ok {
				fieldValue.Set(reflect.ValueOf(actionObject.File(bindName)))
			} else {
				filePtr := actionObject.File(fieldName)
				if filePtr == nil {
					var lowerFirstName = strings.ToLower(string(fieldName[0])) + fieldName[1:]
					filePtr = actionObject.File(lowerFirstName)
				}
				fieldValue.Set(reflect.ValueOf(filePtr))
			}
			continue
		case File: // 支持文件
			bindName, ok := field.Tag.Lookup("field")
			if ok {
				filePtr := actionObject.File(bindName)
				if filePtr != nil {
					fieldValue.Set(reflect.ValueOf(*filePtr))
				}
			} else {
				filePtr := actionObject.File(fieldName)
				if filePtr == nil {
					var lowerFirstName = strings.ToLower(string(fieldName[0])) + fieldName[1:]
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
		var fieldParamValue []string
		var hasValue = false

		cookieName, ok := field.Tag.Lookup("cookie")
		var hasTagValue = false
		if ok && len(cookieName) > 0 {
			hasTagValue = true
			cookieValue, err := request.Cookie(cookieName)
			if err != nil {
				hasValue = false
			} else {
				fieldParamValue = []string{strings.TrimSpace(cookieValue.Value)}
				hasValue = len(fieldParamValue) > 0
			}
		}

		// session:"Session参数名"
		sessionName, ok := field.Tag.Lookup("session")
		if ok && len(sessionName) > 0 {
			hasTagValue = true

			session := actionPtrValue.Interface().(ActionWrapper).Object().Session()
			if session != nil {
				sessionValue := session.GetString(sessionName)
				fieldParamValue = []string{strings.TrimSpace(sessionValue)}
				hasValue = len(fieldParamValue) > 0
			}
		}

		// 从request参数中读取
		if !hasTagValue {
			fieldParamValue, hasValue = getActionParamFuzzy(&params, fieldName)
		}

		// default:"默认值"
		// 注意：DefaultValue不对字符串进行Trim()处理
		if !hasValue || (len(fieldParamValue) == 0 || len(fieldParamValue[0]) == 0) {
			defaultValue, ok := field.Tag.Lookup("default")
			if ok {
				hasValue = true
				fieldParamValue = []string{defaultValue}
			}
		}

		if hasValue {
			firstParamValue := strings.TrimFunc(fieldParamValue[0], unicode.IsSpace)
			switch field.Type.Kind() {
			case reflect.Int:
				fieldValue.Set(reflect.ValueOf(types.Int(firstParamValue)))
			case reflect.Int8:
				fieldValue.Set(reflect.ValueOf(types.Int8(firstParamValue)))
			case reflect.Int16:
				fieldValue.Set(reflect.ValueOf(types.Int16(firstParamValue)))
			case reflect.Int32:
				fieldValue.Set(reflect.ValueOf(types.Int32(firstParamValue)))
			case reflect.Int64:
				fieldValue.Set(reflect.ValueOf(types.Int64(firstParamValue)))
			case reflect.Uint:
				fieldValue.Set(reflect.ValueOf(types.Uint(firstParamValue)))
			case reflect.Uint8:
				fieldValue.Set(reflect.ValueOf(types.Uint8(firstParamValue)))
			case reflect.Uint16:
				fieldValue.Set(reflect.ValueOf(types.Uint16(firstParamValue)))
			case reflect.Uint32:
				fieldValue.Set(reflect.ValueOf(types.Uint32(firstParamValue)))
			case reflect.Uint64:
				fieldValue.Set(reflect.ValueOf(types.Uint64(firstParamValue)))
			case reflect.Bool:
				if lists.Contains([]string{"on", "true", "yes", "enabled"}, firstParamValue) {
					fieldValue.SetBool(true)
				} else {
					fieldValue.SetBool(types.Bool(firstParamValue))
				}
			case reflect.String:
				fieldValue.Set(reflect.ValueOf(firstParamValue))
			case reflect.Float32:
				fieldValue.Set(reflect.ValueOf(types.Float32(firstParamValue)))
			case reflect.Float64:
				fieldValue.Set(reflect.ValueOf(types.Float64(firstParamValue)))
			case reflect.Slice:
				elemKind := field.Type.Elem().Kind()
				if elemKind == reflect.Uint8 { // 字节
					fieldValue.SetBytes([]byte(firstParamValue))
				} else { // slice
					sliceValue, err := types.Slice(fieldParamValue, field.Type)
					if err == nil {
						// trim right spaces in string value
						if stringSlice, ok := sliceValue.([]string); ok {
							for k, v := range stringSlice {
								stringSlice[k] = strings.TrimFunc(v, unicode.IsSpace)
							}
						}
						fieldValue.Set(reflect.ValueOf(sliceValue))
					}
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
		var firstArg reflect.Value
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
	var returnValues []reflect.Value
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

func getActionParam(params *Params, name string) (value []string, has bool) {
	values, ok := (*params)[name]
	if !ok {
		return nil, false
	}
	if len(values) == 0 {
		return nil, false
	}
	return values, true
}

func getActionParamFuzzy(params *Params, name string) (value []string, has bool) {
	value, hasValue := getActionParam(params, name)
	if hasValue {
		return value, true
	}

	var lowerFirstName = strings.ToLower(string(name[0])) + name[1:]
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

// copy from golang source
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
		name := tag[:i]
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
		qvalue := tag[:i+1]
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			break
		}
		tags[name] = value
	}
	return tags
}
