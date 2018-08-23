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
