package dbs

import (
	"context"
	"database/sql"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"sync/atomic"
)

var globalTxId int64 = 1

type Tx struct {
	db    *DB
	rawTx *sql.Tx
	id    int64

	isDone bool
}

func NewTx(db *DB, raw *sql.Tx) *Tx {
	return &Tx{
		db:    db,
		rawTx: raw,
		id:    atomic.AddInt64(&globalTxId, 1),
	}
}

func (this *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return this.rawTx.Exec(query, args...)
}

func (this *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return this.rawTx.QueryContext(ctx, query, args...)
}

func (this *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return this.rawTx.Query(query, args...)
}

func (this *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return this.rawTx.QueryRowContext(ctx, query, args...)
}

func (this *Tx) QueryRow(query string, args ...any) *sql.Row {
	return this.rawTx.QueryRow(query, args...)
}

func (this *Tx) Prepare(query string) (*Stmt, error) {
	return this.db.stmtManager.Prepare(this.rawTx, query)
}

func (this *Tx) PrepareOnce(query string) (*Stmt, bool, error) {
	return this.db.stmtManager.PrepareOnce(this.rawTx, query, this.id)
}

func (this *Tx) FindOnes(query string, args ...any) (ones []maps.Map, columnNames []string, err error) {
	rawRows, err := this.rawTx.Query(query, args...)
	if err != nil {
		return nil, nil, err
	}

	var rows = NewRows(rawRows)
	defer func() {
		_ = rows.Close()
	}()

	columnNames, err = rows.Columns()
	if err != nil {
		return
	}

	ones, err = rows.FindOnes()
	return
}

func (this *Tx) FindOne(query string, args ...any) (one maps.Map, err error) {
	rawRows, err := this.rawTx.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var rows = NewRows(rawRows)
	defer func() {
		_ = rows.Close()
	}()
	return rows.FindOne()
}

func (this *Tx) FindCol(colIndex int, query string, args ...any) (colValue any, err error) {
	rawRows, err := this.rawTx.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var rows = NewRows(rawRows)
	defer func() {
		_ = rows.Close()
	}()
	return rows.FindCol(colIndex)
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
	return this.rawTx.Commit()
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
	return this.rawTx.Rollback()
}

func (this *Tx) Raw() *sql.Tx {
	return this.rawTx
}

func (this *Tx) close() error {
	return this.db.stmtManager.CloseId(this.id)
}
