package dbs

import (
	"database/sql"
	"github.com/iwind/TeaGo/logs"
	"sync/atomic"
)

var globalTxId int64 = 1

type Tx struct {
	db    *DB
	sqlTx *sql.Tx
	id    int64

	isDone bool
}

func NewTx(db *DB, raw *sql.Tx) *Tx {
	return &Tx{
		db:    db,
		sqlTx: raw,
		id:    atomic.AddInt64(&globalTxId, 1),
	}
}

func (this *Tx) Exec(query string, params ...interface{}) (sql.Result, error) {
	return this.sqlTx.Exec(query, params...)
}

func (this *Tx) Prepare(query string) (*Stmt, error) {
	return this.db.stmtManager.Prepare(this.sqlTx, query)
}

func (this *Tx) PrepareOnce(query string) (*Stmt, bool, error) {
	return this.db.stmtManager.PrepareOnce(this.sqlTx, query, this.id)
}

func (this *Tx) Commit() error {
	if this.isDone {
		return nil
	}
	this.isDone = true

	defer func() {
		err := this.close()
		if err != nil {
			logs.Println("[DB]tx close error: " + err.Error())
		}
	}()
	return this.sqlTx.Commit()
}

func (this *Tx) Rollback() error {
	if this.isDone {
		return nil
	}
	this.isDone = true

	defer func() {
		err := this.close()
		if err != nil {
			logs.Println("[DB]tx close error: " + err.Error())
		}
	}()
	return this.sqlTx.Rollback()
}

func (this *Tx) Raw() *sql.Tx {
	return this.sqlTx
}

func (this *Tx) close() error {
	this.db.stmtManager.CloseId(this.id)
	return nil
}
