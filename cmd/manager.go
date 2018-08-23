package cmd

import (
	"errors"
	"log"
	"github.com/iwind/TeaGo/logs"
	"reflect"
)

var commandsPtrMap = map[string]CommandInterface{}

func Register(cmdPtr interface{}) {
	command, ok := cmdPtr.(CommandInterface)
	if !ok {
		logs.Fatalf("command '%#v' must implement 'CommandInterface' interface", cmdPtr)
	}

	value := reflect.ValueOf(command)
	field := value.Elem().FieldByName("Command")
	if field.IsNil() {
		field.Set(reflect.ValueOf(&Command{}))
	}

	codes := command.Codes()
	if len(codes) == 0 {
		log.Fatal("command Codes() should return more than one results")
		return
	}
	for _, code := range codes {
		commandsPtrMap[code] = command
	}
}

func AllCommands() map[string]CommandInterface {
	return commandsPtrMap
}

func Run(code string) error {
	command, ok := commandsPtrMap[code]
	if !ok {
		return errors.New("[cmd]command with code '" + code + "' not found")
	}

	value := reflect.ValueOf(command)
	childCommand := value.Elem().FieldByName("Command").Interface()
	childCommand.(*Command).SubCodeString = code

	command.Run()

	return nil
}

// 尝试执行命令行
func Try(args []string) bool {
	commandArgs = args

	if len(args) == 0 {
		return false
	}

	var code = args[0]
	_, ok := commandsPtrMap[code]

	if ok {
		Run(code)
		return true
	}
	return false
}

// 分析字符串中的参数
func ParseArgs(s string) (args []string) {
	quotesBegin := false
	quotesEscaped := false
	var lastQuote rune
	lastArg := ""
	for index, character := range s {
		if character == '"' || character == '\'' {
			if quotesEscaped {
				lastArg += string(character)
				quotesEscaped = false
				continue
			}
			if quotesBegin {
				if lastQuote == character { // 引号结束
					quotesBegin = false
				} else {
					// 视为参数的一部分
					lastArg += string(character)
				}
			} else {
				quotesBegin = true
				lastQuote = character
			}
		} else if character == '\\' {
			if len(s) > index+1 && (s[index+1:index+2][0] == '"' || s[index+1:index+2][0] == '\'') {
				quotesEscaped = true
			} else {
				lastArg += string(character)
			}
		} else if character == ' ' || character == '\t' || character == '\n' || character == '\r' {
			if quotesBegin { // 如果在引号中，则视为参数的一部分
				lastArg += string(character)
			} else { // 如果不在引号中，则认为参数已结束
				if len(lastArg) > 0 {
					args = append(args, lastArg)
					lastArg = ""
				}
			}
		} else {
			lastArg += string(character)
		}
	}

	if len(lastArg) > 0 {
		args = append(args, lastArg)
	}

	return args
}
