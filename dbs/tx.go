package dbs

import (
	"database/sql"
	"github.com/iwind/TeaGo/logs"
	"sync"
)

type Tx struct {
	raw *sql.Tx

	stmtMap map[string]*Stmt // query => stmt
	locker  *sync.Mutex

	isDone bool
}

func NewTx(raw *sql.Tx) *Tx {
	return &Tx{
		raw:     raw,
		stmtMap: map[string]*Stmt{},
		locker:  &sync.Mutex{},
	}
}

func (this *Tx) Exec(query string, params ...interface{}) (sql.Result, error) {
	return this.raw.Exec(query, params...)
}

func (this *Tx) Prepare(query string) (*Stmt, error) {
	return BuildStmt(this.raw.Prepare(query))
}

func (this *Tx) Commit() error {
	this.locker.Lock()
	defer this.locker.Unlock()

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
	return this.raw.Commit()
}

func (this *Tx) Rollback() error {
	this.locker.Lock()
	defer this.locker.Unlock()

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
	return this.raw.Rollback()
}

func (this *Tx) PrepareOnce(query string) (*Stmt, error) {
	this.locker.Lock()
	defer this.locker.Unlock()

	var stmt, ok = this.stmtMap[query]
	if ok {
		return stmt, nil
	}

	sqlStmt, err := this.raw.Prepare(query)
	if err != nil {
		return BuildStmt(sqlStmt, err)
	}

	stmt, _ = BuildStmt(sqlStmt, nil)

	this.stmtMap[query] = stmt
	return stmt, nil
}

func (this *Tx) Raw() *sql.Tx {
	return this.raw
}

func (this *Tx) close() error {
	// 这里不需要locker，因为在调用它的函数里已经使用locker

	for _, stmt := range this.stmtMap {
		err := stmt.Close()
		if err != nil {
			return err
		}
	}
	this.stmtMap = map[string]*Stmt{}
	return nil
}
