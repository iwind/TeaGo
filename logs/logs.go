package logs

import (
	"fmt"
	"github.com/iwind/TeaGo/utils/string"
	"log"
	"runtime"
	"os"
	"bytes"
	"strings"
	"path/filepath"
)

type Writer interface {
	Init()
	Write(log string)
	Close()
}

const escape = "\x1b"

var colors = map[string]string{
	"warn":    "1;33", // 0细体，1粗体，
	"warning": "1;33",
	"error":   "1;31", //"0;31",
	"success": "1;32",
	"ok":      "1;32",
	"code":    "1;33",

	"black":  "1;30",
	"red":    "1;31",
	"green":  "1;32",
	"yellow": "1;33",
	"blue":   "1;34",
	"pink":   "1;35",
	"cyan":   "1;36",
	"white":  "1;37",
}
var supportsColor = runtime.GOOS != "windows"
var writer Writer
var isOn = true

func On() {
	isOn = true
}

func Off() {
	isOn = false
}

func SetWriter(newWriter Writer) {
	if writer != nil {
		writer.Close()
	}
	writer = newWriter
}

func Println(args ... interface{}) {
	if !isOn {
		return
	}

	if writer != nil {
		// 给args中间加入空格
		newArgs := []interface{}{}
		countArgs := len(args)
		for index, arg := range args {
			if index < countArgs-1 {
				newArgs = append(newArgs, arg, " ")
			} else {
				newArgs = append(newArgs, arg)
			}
		}
		writer.Write(fmt.Sprint(newArgs ...))
	} else {
		log.Println(args ...)
	}
}

func Printf(format string, args ... interface{}) {
	if !isOn {
		return
	}

	if writer != nil {
		writer.Write(Sprintf(format, args ...))
	} else {
		log.Println(Sprintf(format, args ...))
	}
}

func Sprintf(format string, args ... interface{}) string {
	{
		reg, _ := stringutil.RegexpCompile("<(\\w+)>")
		format = reg.ReplaceAllStringFunc(format, func(value string) string {
			var tag = value[1 : len(value)-1]
			color, ok := colors[tag]
			if !ok {
				return value
			}
			if !supportsColor {
				return ""
			}
			return escape + "[" + color + "m"
		})
	}
	{
		reg, _ := stringutil.RegexpCompile("</(\\w+)>")
		format = reg.ReplaceAllStringFunc(format, func(value string) string {
			var tag = value[2 : len(value)-1]
			_, ok := colors[tag]
			if !ok {
				return value
			}
			if !supportsColor {
				return ""
			}
			return escape + "[0m"
		})
	}

	return fmt.Sprintf(format, args ...)
}

func Infof(format string, args ... interface{}) {
	Printf("[INFO]"+format, args ...)
}

func Codef(format string, args ...interface{}) {
	Printf("[CODE]<code>"+format+"</code>", args ...)
}

func Debugf(format string, args ...interface{}) {
	Printf("[DEBUG]<code>"+format+"</code>", args ...)
}

func Successf(format string, args ...interface{}) {
	Printf("[SUCCESS]<success>"+format+"</success>", args ...)
}

func Warnf(format string, args ...interface{}) {
	Printf("[WARN]<warn>"+format+"</warn>", args ...)
}

func Errorf(format string, args ... interface{}) {
	errorString := fmt.Sprintf(format, args ...)

	// 调用stack
	_, currentFilename, _, currentOk := runtime.Caller(0)
	if currentOk {
		for i := 1; i < 32; i ++ {
			_, filename, lineNo, ok := runtime.Caller(i)
			if !ok {
				break
			}

			if filename == currentFilename {
				continue
			}

			goPath := os.Getenv("GOPATH")
			if len(goPath) > 0 {
				absGoPath, err := filepath.Abs(goPath)
				if err == nil {
					filename = strings.TrimPrefix(filename, absGoPath)[1:]
				}
			}

			errorString += "\n\t\t" + string(filename) + ":" + fmt.Sprintf("%d", lineNo)

			break
		}
	}

	Printf("%s", errorString)
}

func Error(err error) {
	if err == nil {
		Errorf("nil")
		return
	}
	errorString := err.Error()

	Errorf("%s", errorString)
}

func Fatalf(format string, args ... interface{}) {
	Printf("[FATAL]<error>"+format+"</error>", args ...)
	os.Exit(0)
}

func Fatal(err error) {
	Println("[FATAL]<error>" + err.Error() + "</error>")
	os.Exit(0)
}

func Dump(variable interface{}) {
	var s = fmt.Sprintf("%#v", variable)
	var buffer = bytes.Buffer{}
	var indent = 0
	var last rune
	var inQuote = false
	for _, c := range s {
		if inQuote {
			if c == '"' && last != '\\' {
				inQuote = false
			} else {
				buffer.WriteRune(c)
				last = c
				continue
			}
		} else {
			if c == '"' && last != '\\' {
				inQuote = true
			}
		}

		if c != '}' && last == '{' {
			indent ++
			buffer.WriteRune('\n')
			buffer.WriteRune(' ')
			buffer.WriteString(strings.Repeat("  ", indent))
		} else if c == '}' && last != '{' {
			buffer.WriteRune('\n')
			if indent < 0 {
				break
			}
			buffer.WriteString(strings.Repeat("  ", indent))
			indent --
		}

		buffer.WriteRune(c)

		if c == ',' {
			buffer.WriteRune('\n')
			buffer.WriteString(strings.Repeat("  ", indent))
		}

		last = c
	}

	Printf("<code>%s</code>", buffer.String())
}
