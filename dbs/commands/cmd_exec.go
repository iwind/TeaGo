package commands

import (
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/dbs"
)

type ExecCommand struct {
	*cmd.Command
}

func (command *ExecCommand) Name() string {
	return "exec sql on db"
}

func (command *ExecCommand) Codes() []string {
	return []string{":db.exec"}
}

func (command *ExecCommand) Usage() string {
	return ":db.exec SQL [-db=DB ID]"
}

func (command *ExecCommand) Run() {
	sql, found := command.Arg(1)
	if !found {
		command.ErrorString("need sql to exec")
		command.Output("Usage:\n")
		command.Output("<code>   " + command.Usage() + "</code>\n")
		return
	}

	dbId, found := command.Param("db")
	var db *dbs.DB
	if !found {
		newDB, err := dbs.Default()
		if err != nil {
			command.Error(err)
			return
		}
		db = newDB
	} else {
		newDB, err := dbs.NewInstance(dbId)
		if err != nil {
			command.Error(err)
			return
		}
		db = newDB
		defer newDB.Close()
	}

	result, err := db.Exec(sql)
	if err != nil {
		command.Error(err)
		return
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		command.Error(err)
		return
	}
	command.Output("~~~\n<code>" + sql + "</code>\n~~~\n")
	command.Output("<ok>SQL executed successfully, and affected rows:", affectedRows, "</ok>\n")
}
