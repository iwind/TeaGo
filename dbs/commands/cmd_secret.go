package commands

import (
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/types"
	"github.com/iwind/TeaGo/utils/string"
)

type SecretCommand struct {
	*cmd.Command
}

func (this *SecretCommand) Name() string {
	return "generate secret string"
}

func (this *SecretCommand) Codes() []string {
	return []string{":db.secret"}
}

func (this *SecretCommand) Usage() string {
	return ":db.secret [LENGTH]"
}

func (this *SecretCommand) Run() {
	lengthArg, found := this.Arg(1)
	length := 32
	if found {
		lengthInt := types.Int(lengthArg)
		if lengthInt > 0 {
			length = lengthInt
		}
	}
	this.Output("<code>" + stringutil.Rand(length) + "</code>\n")
}
