package cmd

import (
	"testing"
)

type testCommand struct {
	*Command
}

func (command *testCommand) Codes() []string {
	return []string{"test"}
}

func (command *testCommand) Run() {
	command.Println("Run Command")
}

func TestRegister(t *testing.T) {
	var command = &testCommand{}
	Register(command)
	t.Log(Run("test"))
}

func TestCommand_Param(t *testing.T) {
	commandArgs = []string{"-name", "-age=20"}
	cmd := &Command{}
	t.Log(cmd.Param("name"))
	t.Log(cmd.Param("age"))
	t.Log(cmd.Param("hello"))
}
