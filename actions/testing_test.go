package actions

import (
	"testing"
)

type testAction Action

func (this *testAction) Run(params struct {
	Name string
	Age  int
}) {
	//this.WriteString("Hello, World", "\n")
	//this.WriteString(this.Request.URL.String(), "\n")
	//this.WriteString(this.Request.Method, "\n")
	//this.WriteFormat("name:%s, age:%d", params.Name, params.Age)
}

func TestActionTesting(t *testing.T) {
	test := NewTesting(new(testAction)).
		URL("/hello/world").
		Method("POST").
		RemoteAddr("127.0.0.1:1234").
		AddHeader("User-Agent", "Go API").
		Cost().
		Params(Params{
			"name": []string{"Lu"},
			"age":  []string{"20"},
		})
	resp := test.Run(t)
	t.Log(string(resp.Data))
}
