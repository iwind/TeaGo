package dbs

import (
	"database/sql"
	"github.com/iwind/TeaGo/maps"
)

type SQLExecutor interface {
	// PrepareOnce 可重用的Prepare
	PrepareOnce(query string) (stmt *Stmt, cached bool, err error)

	Exec(query string, args ...any) (result sql.Result, err error)

	FindOnes(query string, args ...any) (ones []maps.Map, columnNames []string, err error)
}
