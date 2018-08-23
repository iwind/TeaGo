package commands

import (
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/types"
	"fmt"
	"github.com/iwind/TeaGo/dbs"
)

type FixCommand struct {
	*cmd.Command
}

func (command *FixCommand) Name() string {
	return "fix database compare issues"
}

func (command *FixCommand) Codes() []string {
	return []string{":db.fix"}
}

func (command *FixCommand) Usage() string {
	return ":db.fix [ISSUE ID]"
}

func (command *FixCommand) Run() {
	issueIdString, found := command.Arg(1)
	if !found {
		command.Output("Usage:\n")
		command.Output("   <code>" + command.Usage() + "</code>\n")
		return
	}

	issueId := types.Int(issueIdString)
	if issueId <= 0 {
		command.ErrorString("issue id should be a valid number")
		return
	}

	issue, found := compareIssues[issueId]
	if !found {
		command.ErrorString(fmt.Sprintf("issue with id '%d' not found", issueId))
		return
	}

	command.Output("\n~~~\n<code>" + issue.sql + "</code>\n~~~\n")
	db, err := dbs.NewInstance(issue.dbId)
	if err != nil {
		command.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec(issue.sql)
	if err != nil {
		command.Error(err)
		return
	}

	command.Output("<ok>fix on '" + issue.dbId + "' successfully</ok>\n")
}
