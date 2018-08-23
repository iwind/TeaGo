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

func (this *FixCommand) Name() string {
	return "fix database compare issues"
}

func (this *FixCommand) Codes() []string {
	return []string{":db.fix"}
}

func (this *FixCommand) Usage() string {
	return ":db.fix [ISSUE ID]"
}

func (this *FixCommand) Run() {
	issueIdString, found := this.Arg(1)
	if !found {
		this.Output("Usage:\n")
		this.Output("   <code>" + this.Usage() + "</code>\n")
		return
	}

	issueId := types.Int(issueIdString)
	if issueId <= 0 {
		this.ErrorString("issue id should be a valid number")
		return
	}

	issue, found := compareIssues[issueId]
	if !found {
		this.ErrorString(fmt.Sprintf("issue with id '%d' not found", issueId))
		return
	}

	this.Output("\n~~~\n<code>" + issue.sql + "</code>\n~~~\n")
	db, err := dbs.NewInstance(issue.dbId)
	if err != nil {
		this.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec(issue.sql)
	if err != nil {
		this.Error(err)
		return
	}

	this.Output("<ok>fix on '" + issue.dbId + "' successfully</ok>\n")
}
