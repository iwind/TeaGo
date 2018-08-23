package cmd

import (
	"os"
	"github.com/iwind/TeaGo/logs"
	"strings"
	"fmt"
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

func (this *Command) Output(message ... interface{}) {
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

func (this *Command) Println(message ... interface{}) {
	logs.Println(message ...)
}

func (this *Command) Printf(format string, args ... interface{}) {
	logs.Printf(format, args ...)
}

func (this *Command) Error(err error) {
	this.Output("<error>"+err.Error()+"</error>", "\n")
}

func (this *Command) ErrorString(err string) {
	this.Output("<error>"+err+"</error>", "\n")
}

func (this *Command) Param(key string) (value string, found bool) {
	if len(key) == 0 {
		return "", false
	}
	for _, arg := range commandArgs {
		for _, prefix := range []string{"-" + key + "=", key + "="} {
			if strings.HasPrefix(arg, prefix) {
				return string(arg[len(prefix):]), true
			}
		}
	}
	return "", false
}
