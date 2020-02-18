package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/iwind/TeaGo/cmd"
	_ "github.com/iwind/TeaGo/dbs/commands"
	"github.com/iwind/TeaGo/lists"
	"os"
	"time"
)

// TeaTool工具
func main() {
	r := bufio.NewReader(os.Stdin)
	lastCommand := ""

	for {
		time.Sleep(400 * time.Millisecond)
		fmt.Print("> ")

		line, _, err := r.ReadLine()
		if err != nil {
			continue
		}

		command := string(bytes.TrimSpace(line))

		// 命令帮助
		if len(command) == 0 || command == "help" || command == "h" || command == "?" || command == "/?" {
			lastCommand = command
			fmt.Println("TeaTool commands:")
			commands := cmd.AllCommands()

			// 对命令代码进行排序
			codes := []string{}
			for code := range commands {
				codes = append(codes, code)
			}

			lists.Sort(codes, func(i int, j int) bool {
				code1 := codes[i]
				code2 := codes[j]
				return code1 < code2
			})

			//输出
			for _, code := range codes {
				ptr := commands[code]
				fmt.Println("  ", code+"\n\t\t"+ptr.Name())
			}
			continue
		}

		if command == "retry" || command == "!!" /** csh like **/ || command == "!-1" /** csh like **/ {
			command = lastCommand
			fmt.Println("retry '" + command + "'")
		}
		lastCommand = command

		found := cmd.Try(cmd.ParseArgs(command))
		if !found {
			fmt.Println("command '" + command + "' not found")
		}
	}

	time.Sleep(1 * time.Hour)
}
