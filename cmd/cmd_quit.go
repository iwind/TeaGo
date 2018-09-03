package cmd

import "os"

// 退出命令
type QuitCommand struct {
	*Command
}

func (this *QuitCommand) Name() string {
	return "Quit command"
}

func (this *QuitCommand) Usage() string {
	return "[quit|exit]"
}

func (this *QuitCommand) Codes() []string {
	return []string{"quit", "exit"}
}

func (this *QuitCommand) Run() {
	os.Exit(0)
}
