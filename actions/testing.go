package actions

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type Testing struct {
	actionPtr  ActionWrapper
	requestURL string
	params     Params
	method     string
	remoteAddr string
	header     http.Header
	cost       bool
}

// 新建测试实例
func NewTesting(actionPtr ActionWrapper) *Testing {
	return &Testing{
		actionPtr: actionPtr,
		method:    "GET",
		header:    http.Header{},
	}
}

// 设置参数
func (this *Testing) Params(params Params) *Testing {
	this.params = params
	return this
}

// 设置请求方法
func (this *Testing) Method(method string) *Testing {
	this.method = method
	return this
}

// 设置URL
func (this *Testing) URL(urlString string) *Testing {
	this.requestURL = urlString
	return this
}

// 设置终端地址
func (this *Testing) RemoteAddr(remoteAddr string) *Testing {
	this.remoteAddr = remoteAddr
	return this
}

// 添加Header
func (this *Testing) AddHeader(key string, value string) *Testing {
	this.header.Add(key, value)
	return this
}

// 设置Header
func (this *Testing) SetHeader(key string, value string) *Testing {
	this.header.Set(key, value)
	return this
}

// 计算耗时
func (this *Testing) Cost() *Testing {
	this.cost = true
	return this
}

// 执行
func (this *Testing) Run(t *testing.T) (*TestingResponseWriter) {
	values := url.Values{}

	for k, v := range this.params {
		values[k] = v
	}

	request, err := http.NewRequest(this.method, this.requestURL, strings.NewReader(values.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	if strings.ToUpper(this.method) == "POST" {
		request.Form = values
	}

	request.RemoteAddr = this.remoteAddr
	request.Header = this.header

	spec := NewActionSpec(this.actionPtr)

	beforeTime := time.Now()

	resp := &TestingResponseWriter{}
	RunAction(this.actionPtr, spec, request, resp, this.params)

	if this.cost {
		t.Logf("cost:%.6f %s", time.Since(beforeTime).Seconds()*1000, "ms")
	}

	return resp
}
