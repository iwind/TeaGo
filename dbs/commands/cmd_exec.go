package commands

import (
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/dbs"
)

type ExecCommand struct {
	*cmd.Command
}

func (this *ExecCommand) Name() string {
	return "exec sql on db"
}

func (this *ExecCommand) Codes() []string {
	return []string{":db.exec"}
}

func (this *ExecCommand) Usage() string {
	return ":db.exec SQL [-db=DB ID]"
}

func (this *ExecCommand) Run() {
	sql, found := this.Arg(1)
	if !found {
		this.ErrorString("need sql to exec")
		this.Output("Usage:\n")
		this.Output("<code>   " + this.Usage() + "</code>\n")
		return
	}

	dbId, found := this.Param("db")
	var db *dbs.DB
	if !found {
		newDB, err := dbs.Default()
		if err != nil {
			this.Error(err)
			return
		}
		db = newDB
	} else {
		newDB, err := dbs.NewInstance(dbId)
		if err != nil {
			this.Error(err)
			return
		}
		db = newDB
		defer newDB.Close()
	}

	result, err := db.Exec(sql)
	if err != nil {
		this.Error(err)
		return
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		this.Error(err)
		return
	}
	this.Output("~~~\n<code>" + sql + "</code>\n~~~\n")
	this.Output("<ok>SQL executed successfully, and affected rows:", affectedRows, "</ok>\n")
}
