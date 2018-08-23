package TeaGo

import (
	"net/http"
	"strings"
	"regexp"
	"net/url"
	"github.com/iwind/TeaGo/actions"
	"github.com/iwind/TeaGo/utils/string"
	"github.com/iwind/TeaGo/tasks"
	"sync"
	"os"
	"runtime"
	"time"
	"github.com/iwind/TeaGo/logs"
	"reflect"
	"github.com/iwind/TeaGo/Tea"
	"path/filepath"
	"mime"
	"github.com/iwind/TeaGo/files"
	"github.com/iwind/TeaGo/types"
)

type Server struct {
	directRoutes   map[string]func(writer http.ResponseWriter, request *http.Request)
	patternRoutes  []ServerRoutePattern
	staticDirs     []ServerStaticDir
	sessionManager interface{}
	lastModule     string
	config         *serverConfig
	logWriter      LogWriter
	accessLog      bool // 是否记录访问日志
}

type ServerRoutePattern struct {
	module  string
	reg     regexp.Regexp
	names   []string
	method  string
	runFunc func(writer http.ResponseWriter, request *http.Request)
}

type ServerStaticDir struct {
	prefix string
	dir    string
}

var textMimeMap = map[string]bool{
	"application/atom+xml":                true,
	"application/javascript":              true,
	"application/x-javascript":            true,
	"application/json":                    true,
	"application/rss+xml":                 true,
	"application/x-web-app-manifest+json": true,
	"application/xhtml+xml":               true,
	"application/xml":                     true,
	"image/svg+xml":                       true,
	"text/css":                            true,
	"text/plain":                          true,
	"text/javascript":                     true,
	"text/xml":                            true,
	"text/html":                           true,
	"text/xhtml":                          true,
	"text/sgml":                           true,
}

func NewServer() *Server {
	var server = &Server{
		accessLog: true,
	}
	server.Init()

	return server
}

func (this *Server) Init() {
	this.directRoutes = make(map[string]func(writer http.ResponseWriter, request *http.Request))
	this.patternRoutes = []ServerRoutePattern{}
	this.staticDirs = []ServerStaticDir{}

	// 配置
	this.config = &serverConfig{}
	this.config.Load()

	// 日志
	this.LogWriter(&DefaultLogWriter{})

	// 检查PID
	this.checkPid()

	// 记录PID
	this.writePid()
}

// 启动服务
func (this *Server) Start() {
	var address = this.config.Listen
	if len(address) > 0 {
		this.StartOn(address)
	} else {
		this.StartOn("0.0.0.0:8888")
	}
}

// 在某个地址上启动服务
func (this *Server) StartOn(address string) {
	var serverMux = http.NewServeMux()

	// 静态资源目录
	for _, staticDir := range this.staticDirs {
		var prefix = staticDir.prefix
		if len(prefix) > 0 {
			if !strings.HasPrefix(prefix, "/") {
				prefix = "/" + prefix
			}
			if !strings.HasSuffix(prefix, "/") {
				prefix += "/"
			}

			serverMux.HandleFunc(prefix, func(writer http.ResponseWriter, request *http.Request) {
				writer = newResponseWriter(writer)

				// 输出日志
				if this.accessLog {
					defer this.logWriter.Print(time.Now(), writer.(*responseWriter), request)
				}

				this.outputMimeType(writer, request.URL.Path)
				http.StripPrefix(strings.TrimSuffix(prefix, "/"), http.FileServer(http.Dir(staticDir.dir+"/"))).ServeHTTP(writer, request)
			})
		}
	}

	// 加载和动作一致的静态资源
	serverMux.HandleFunc("/_/", func(writer http.ResponseWriter, request *http.Request) {
		writer = newResponseWriter(writer)

		// 输出日志
		if this.accessLog {
			defer this.logWriter.Print(time.Now(), writer.(*responseWriter), request)
		}

		ext := strings.ToLower(filepath.Ext(request.URL.Path))
		if stringutil.Contains([]string{".html"}, ext) {
			http.Error(writer, "No permission to view page", http.StatusForbidden)
			return
		}
		request.URL.Path = strings.TrimPrefix(request.URL.Path, "/_/")

		this.outputMimeType(writer, request.URL.Path)
		http.FileServer(http.Dir(Tea.ViewsDir())).ServeHTTP(writer, request)
	})

	// 请求处理函数
	var moduleReg, err = stringutil.RegexpCompile("^/+@([\\w-]+)(/.*)$")
	if err != nil {
		panic(err)
		return
	}

	serverMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer = newResponseWriter(writer)

		// 输出日志
		if this.accessLog {
			defer this.logWriter.Print(time.Now(), writer.(*responseWriter), request)
		}

		var requestPath = request.URL.Path

		// 模块
		parsedResult := moduleReg.FindAllStringSubmatch(requestPath, -1)
		var module = ""
		if len(parsedResult) > 0 {
			module = parsedResult[0][1]
			requestPath = parsedResult[0][2]
		}

		// 路由key
		var key string
		if len(module) > 0 {
			key = module + "/" + requestPath + "__MELOY__" + request.Method
		} else {
			key = requestPath + "__MELOY__" + request.Method
		}

		runFunc, found := this.directRoutes[key]
		if found {
			runFunc(writer, request)
			return
		}

		// 支持 *
		key = requestPath + "__MELOY__*"
		runFunc, found = this.directRoutes[key]
		if found {
			runFunc(writer, request)
			return
		}

		// 查找pattern
		if len(this.patternRoutes) > 0 {
			for _, route := range this.patternRoutes {
				if route.method != "*" && route.method != request.Method {
					continue
				}

				if route.reg.MatchString(requestPath) {
					matches := route.reg.FindStringSubmatch(request.URL.Path)

					values := url.Values{}
					for index, name := range route.names {
						values.Add(name, matches[index+1])
					}

					if len(values) > 0 {
						if len(request.URL.RawQuery) == 0 {
							request.URL.RawQuery = values.Encode()
						} else {
							request.URL.RawQuery += "&" + values.Encode()
						}
					}

					route.runFunc(writer, request)
					return
				}
			}
		}

		// 试图读取静态文件
		publicFilePath := Tea.PublicFile(requestPath)
		stat, err := os.Stat(publicFilePath)
		if err == nil && !stat.IsDir() {
			this.outputMimeType(writer, requestPath)
			http.FileServer(http.Dir(Tea.PublicDir())).ServeHTTP(writer, request)
			return
		}

		// 处理404的情况
		this.config.processError(request, writer, http.StatusNotFound, "404 page not found")
		//http.Error(writer, "404 page not found", http.StatusNotFound)
		return

	})

	var wg = &sync.WaitGroup{}
	wg.Add(1)

	// http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = runtime.NumCPU() * 100
	// http.DefaultTransport.(*http.Transport).MaxIdleConns = runtime.NumCPU() * 2048
	go func() {
		logs.Println("start server on", address)
		err := http.ListenAndServe(address, serverMux)
		if err != nil {
			logs.Errorf(err.Error())
			time.Sleep(100 * time.Millisecond)
			wg.Add(-1)
		}
	}()

	// 启动任务管理器
	tasks.Start(runtime.NumCPU() * 4)

	// 等待
	wg.Wait()
}

func (this *Server) router(pattern string, method string, actionPtr interface{}) {
	if reflect.TypeOf(actionPtr).Kind().String() != "ptr" {
		actionPtr = reflect.New(reflect.TypeOf(actionPtr)).Interface()
	}
	actionPtr.(actions.ActionWrapper).Object().Module = this.lastModule
	method = strings.ToUpper(method)

	// 是否包含匹配参数
	reg, err := regexp.Compile(":(?:(\\w+)(\\s*(\\([^)]+\\))?))")
	if err != nil {
		logs.Errorf("%s", err.Error())
		return
	}

	matches := reg.FindAllStringSubmatch(pattern, 10)
	names := []string{}
	for _, match := range matches {
		names = append(names, match[1])
		if len(match[3]) > 0 {
			pattern = strings.Replace(pattern, match[0], match[3], 10)
		} else {
			pattern = strings.Replace(pattern, match[0], "([^/]+)", 10)
		}
	}

	reg, err = regexp.Compile("^" + pattern + "$")
	if err == nil && len(names) > 0 {
		var routePattern = ServerRoutePattern{
			module:  this.lastModule,
			reg:     *reg,
			names:   names,
			method:  method,
			runFunc: this.buildHandle(actionPtr),
		}
		this.patternRoutes = append(this.patternRoutes, routePattern)
		return
	}

	// 正常的路由
	var key string
	if len(this.lastModule) > 0 {
		key = this.lastModule + "/" + pattern + "__MELOY__" + method
	} else {
		key = pattern + "__MELOY__" + method
	}
	this.directRoutes[key] = this.buildHandle(actionPtr)
}

func (this *Server) buildHandle(actionPtr interface{}) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		// URI Query
		var params = actions.Params{}
		for key, values := range request.URL.Query() {
			params[key] = values
		}

		// POST参数
		if request.Method == "POST" {
			maxSize := int64(this.config.MaxSize())
			if maxSize <= 0 {
				maxSize = 2 << 10
			}
			err := request.ParseMultipartForm(maxSize)
			if err != nil {
				err := request.ParseForm()
				if err != nil {
					logs.Error(err)
				}
			}
			for key, values := range request.Form {
				params[key] = values
			}
		}

		actionWrapper, ok := actionPtr.(actions.ActionWrapper)
		if !ok {
			logs.Errorf("actionPtr for path '%s' should be pointer", request.URL.Path)
			return
		}
		actionWrapper.Object().SetMaxSize(this.config.MaxSize())
		actionWrapper.Object().SetSessionManager(this.sessionManager)
		actions.RunAction(actionPtr, request, writer, params)
	}
}

// 设置模块定义开始
func (this *Server) Module(module string) *Server {
	this.lastModule = module
	return this
}

// 设置模块定义结束
func (this *Server) End() *Server {
	this.lastModule = ""
	return this
}

// 设置 GET 方法路由映射
func (this *Server) Get(path string, actionPtr interface{}) *Server {
	this.router(path, "get", actionPtr)
	return this
}

// 设置 POST 方法路由映射
func (this *Server) Post(path string, actionPtr interface{}) *Server {
	this.router(path, "post", actionPtr)
	return this
}

// 设置 GET 和 POST 方法路由映射
func (this *Server) GetPost(path string, actionPtr interface{}) *Server {
	return this.Any([]string{"get", "post"}, path, actionPtr)
}

// 设置 HEAD 方法路由映射
func (this *Server) Head(path string, actionPtr interface{}) *Server {
	this.router(path, "head", actionPtr)
	return this
}

func (this *Server) Delete(path string, actionPtr interface{}) *Server {
	this.router(path, "delete", actionPtr)
	return this
}

func (this *Server) Purge(path string, actionPtr interface{}) *Server {
	this.router(path, "purge", actionPtr)
	return this
}

func (this *Server) Put(path string, actionPtr interface{}) *Server {
	this.router(path, "put", actionPtr)
	return this
}

func (this *Server) Options(path string, actionPtr interface{}) *Server {
	this.router(path, "options", actionPtr)
	return this
}

func (this *Server) Trace(path string, actionPtr interface{}) *Server {
	this.router(path, "trace", actionPtr)
	return this
}

func (this *Server) Connect(path string, actionPtr interface{}) *Server {
	this.router(path, "connect", actionPtr)
	return this
}

func (this *Server) Any(methods []string, path string, actionPtr interface{}) *Server {
	for _, method := range methods {
		this.router(path, method, actionPtr)
	}
	return this
}

func (this *Server) All(path string, actionPtr interface{}) *Server {
	this.router(path, "*", actionPtr)
	return this
}

func (this *Server) Static(prefix string, dir string) *Server {
	this.staticDirs = append(this.staticDirs, ServerStaticDir{
		prefix: prefix,
		dir:    dir,
	})
	return this
}

func (this *Server) Session(sessionManager interface{}) *Server {
	this.sessionManager = sessionManager
	return this
}

func (this *Server) LogWriter(logWriter LogWriter) *Server {
	if this.logWriter != nil {
		this.logWriter.Close()
	}
	logWriter.Init()
	this.logWriter = logWriter
	logs.SetWriter(logWriter)
	return this
}

func (this *Server) AccessLog(bool bool) *Server {
	this.accessLog = bool
	return this
}

func (this *Server) outputMimeType(writer http.ResponseWriter, path string) {
	ext := filepath.Ext(path)
	if len(ext) == 0 {
		return
	}

	mimeType := mime.TypeByExtension(ext)
	if len(mimeType) == 0 {
		return
	}

	_, found := textMimeMap[mimeType]
	if !found {
		return
	}

	writer.Header().Set("Content-Type", mimeType+"; charset="+this.config.Charset)
}

// 检查PID
func (this *Server) checkPid() {
	pidFile := files.NewFile(Tea.Root + "/bin/pid")
	if !pidFile.IsFile() {
		return
	}

	pidString, err := pidFile.ReadAllString()
	if err != nil {
		return
	}

	pid := types.Int(pidString)
	if pid <= 0 {
		return
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	logs.Println("kill previous process", pid)
	process.Kill()

	// 等待sock关闭
	time.Sleep(100 * time.Millisecond)
}

// 写入当前程序的PID，以便后续的管理
func (this *Server) writePid() {
	pid := os.Getpid()
	pidDir := files.NewFile(Tea.Root + "/bin")
	if !pidDir.IsDir() {
		err := pidDir.Mkdir()
		if err == nil {
			pidFile := pidDir.Child("pid")
			pidFile.WriteFormat("%d", pid)
		}
	} else {
		pidFile := pidDir.Child("pid")
		pidFile.WriteFormat("%d", pid)
	}
}
