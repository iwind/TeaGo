# TeaGo - Go语言快速开发框架

## 定义不带参数的Action
*actions/hello.go*
~~~go
package actions

import "github.com/iwind/TeaGo/actions"

type HelloAction actions.Action

func (action *HelloAction) Run()  {
	action.Write("Hello")
}
~~~

## 定义带参数的Action
~~~go
package actions

import "github.com/iwind/TeaGo/actions"

type HelloAction actions.Action

func (action *HelloAction) Run(params struct {
	Name string
	Age  int
}) {
	action.WriteFormat("Name:%s, Age:%d",
		params.Name,
		params.Age)
}

~~~

## 使用Action
~~~go
package MeloySearch

import (
	"github.com/iwind/TeaGo"
	"github.com/iwind/MeloySearch/actions"
)

func Start() {
	var server = TeaGo.NewServer()
	
	// 注册路由
	server.Get("/hello", &actions.HelloAction{})
	
	// 启动服务
	server.Start("0.0.0.0:8000")
}

~~~
