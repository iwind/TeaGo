# TeaGo - Go语言快速开发框架

## 定义不带参数的Action
*actions/hello.go*
~~~go
package actions

import "github.com/iwind/TeaGo/actions"

type HelloAction actions.Action

func (this *HelloAction) Run()  {
	this.Write("Hello")
}
~~~

## 定义带参数的Action
~~~go
package actions

import "github.com/iwind/TeaGo/actions"

type HelloAction actions.Action

func (this *HelloAction) Run(params struct {
	Name string
	Age  int
}) {
	this.WriteFormat("Name:%s, Age:%d",
		params.Name,
		params.Age)
}

~~~

## 使用Action
~~~go
package MeloySearch

import (
	"github.com/iwind/TeaGo"
	"github.com/iwind/MyProject/actions"
)

func Start() {
	var server = TeaGo.NewServer()
	
	// 注册路由
	server.Get("/hello", new(actions.HelloAction))
	
	// 启动服务
	server.Start("0.0.0.0:8000")
}

~~~
