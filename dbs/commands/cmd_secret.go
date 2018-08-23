package commands

import (
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/types"
	"github.com/iwind/TeaGo/utils/string"
)

type SecretCommand struct {
	*cmd.Command
}

func (command *SecretCommand) Name() string {
	return "generate secret string"
}

func (command *SecretCommand) Codes() []string {
	return []string{":db.secret"}
}

func (command *SecretCommand) Usage() string {
	return ":db.secret [LENGTH]"
}

func (command *SecretCommand) Run() {
	lengthArg, found := command.Arg(1)
	length := 32
	if found {
		lengthInt := types.Int(lengthArg)
		if lengthInt > 0 {
			length = lengthInt
		}
	}
	command.Output("<code>" + stringutil.Rand(length) + "</code>\n")
}
