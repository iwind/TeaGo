package commands

import (
	"github.com/iwind/TeaGo/cmd"
	"testing"
)

func TestCompareDBCommand_Run(t *testing.T) {
	cmd.Try([]string{":db.compare", "dev", "remote"})
}
