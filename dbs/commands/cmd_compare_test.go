package commands

import (
	"testing"
	"github.com/iwind/TeaGo/cmd"
)

func TestCompareDBCommand_Run(t *testing.T) {
	cmd.Try([]string{ ":db.compare", "dev", "prodRemote" })
}
