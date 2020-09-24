package TeaGo

import (
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/utils/string"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

// 服务配置
type ServerConfig struct {
	Http struct {
		On     bool     `yaml:"on" json:"on"`
		Listen []string `yaml:"listen" json:"listen"` // 监听地址，带端口
	} `yaml:"http" json:"http"`
	Https struct {
		On     bool     `yaml:"on" json:"on"`
		Listen []string `yaml:"listen" json:"listen"`
		Cert   string   `yaml:"cert" json:"cert"`
		Key    string   `yaml:"key" json:"key"`
	} `yaml:"https" json:"https"`
	Env     string `yaml:"env" json:"env"`         // 环境，dev、test或prod
	Charset string `yaml:"charset" json:"charset"` // 字符集
	Upload  struct {
		MaxSize      string `yaml:"maxSize" json:"maxSize"` // 允许上传的最大尺寸
		maxSizeFloat float64
	} `yaml:"upload" json:"upload"`                             // 上传配置
	Errors map[string]interface{} `yaml:"errors" json:"errors"` // 错误配置
}

func (this *ServerConfig) Load() {
	// 先查找yaml
	configFile := Tea.ConfigFile("server.yaml")
	_, err := os.Stat(configFile)
	if err != nil {
		// 查找server.conf
		configFile = Tea.ConfigFile("server.conf")
		_, err = os.Stat(configFile)
		if err != nil {
			logs.Errorf("%s", err.Error())
			return
		}
	}

	fileBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		logs.Errorf("%s", err.Error())
		return
	}
	err = yaml.Unmarshal(fileBytes, this)
	if err != nil {
		logs.Errorf("%s", err.Error())
	} else {
		// maxSize
		maxSize, err := stringutil.ParseFileSize(this.Upload.MaxSize)
		if err != nil {
			logs.Errorf("%s", err.Error())
		} else {
			this.Upload.maxSizeFloat = maxSize
		}

		// env
		if this.Env == "" {
			this.Env = "dev"
		}
		if this.Env != Tea.EnvDev && this.Env != Tea.EnvTest && this.Env != Tea.EnvProd {
			logs.Errorf("'env' should be 'dev', 'test' or 'prod'")
		}
		Tea.Env = this.Env

		// 字符集
		if len(this.Charset) == 0 {
			this.Charset = "utf-8"
		}
	}
}

func (this *ServerConfig) MaxSize() float64 {
	return this.Upload.maxSizeFloat
}

func (this *ServerConfig) processError(request *http.Request, response io.Writer, code int, message string) {
	if this.Errors == nil {
		http.Error(response.(http.ResponseWriter), message, code)
		return
	}
	errorConfig, found := this.Errors[strconv.Itoa(code)]
	if !found {
		http.Error(response.(http.ResponseWriter), message, code)
		return
	}
	mapConfig, ok := errorConfig.(map[string]interface{})
	if !ok {
		http.Error(response.(http.ResponseWriter), message, code)
		return
	}
	url, ok := mapConfig["url"]
	if ok {
		urlString, ok := url.(string)
		if ok {
			http.Redirect(response.(http.ResponseWriter), request, urlString, http.StatusMovedPermanently)
			return
		}
	}

	// 读取错误页面
	viewFile, ok := mapConfig["view"]
	if ok {
		// @TODO
		logs.Println(viewFile)
		/**viewFileString, ok := viewFile.(string)
		if ok {
			data := map[string]interface{}{
				"Request": request,
			}
			response.(http.ResponseWriter).WriteHeader(http.StatusNotFound)

			//templates.Render(response, "default", "views", viewFileString, data, template.FuncMap{})
			return
		}**/
	}

	if !ok {
		http.Error(response.(http.ResponseWriter), message, code)
	}
}
