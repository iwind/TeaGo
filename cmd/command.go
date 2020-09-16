package cmd

import (
	"fmt"
	"github.com/iwind/TeaGo/logs"
	"os"
	"strings"
)

var commandArgs = os.Args[1:]

type Command struct {
	SubCodeString string
}

func (this *Command) SubCode() string {
	return this.SubCodeString
}

func (this *Command) Arg(index int) (value string, found bool) {
	if index >= len(commandArgs) {
		return "", false
	}
	return commandArgs[index], true
}

func (this *Command) Output(message ...interface{}) {
	for index, arg := range message {

		_, ok := arg.(string)
		if ok {
			fmt.Print(logs.Sprintf(arg.(string)))
		} else {
			fmt.Print(logs.Sprintf(fmt.Sprintf("%#v", arg)))
		}

		if index < len(message)-1 {
			fmt.Print(" ")
		}
	}
}

func (this *Command) Println(message ...interface{}) {
	logs.Println(message...)
}

func (this *Command) Printf(format string, args ...interface{}) {
	logs.Printf(format, args...)
}

func (this *Command) Error(err error) {
	this.Output("<error>"+err.Error()+"</error>", "\n")
}

func (this *Command) ErrorString(err string) {
	this.Output("<error>"+err+"</error>", "\n")
}

// 获取参数值
func (this *Command) Param(key string) (value string, found bool) {
	if len(key) == 0 {
		return "", false
	}
	for _, arg := range commandArgs {
		for _, prefix := range []string{"--" + key, "-" + key + "=", key + "="} {
			if strings.HasPrefix(arg, prefix) {
				return arg[len(prefix):], true
			}
		}
		if arg == "--"+key || arg == "-"+key || arg == key {
			return "", true
		}
	}
	return "", false
}

// 判断是否含有某参数
func (this *Command) HasParam(key string) bool {
	_, found := this.Param(key)
	return found
}
