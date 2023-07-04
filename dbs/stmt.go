package dbs

import (
	"database/sql"
	"github.com/iwind/TeaGo/maps"
)

// Stmt SQL语句
type Stmt struct {
	accessAt int64
	rawStmt  *sql.Stmt
}

// NewStmt 构造
func NewStmt(stmt *sql.Stmt) *Stmt {
	return &Stmt{
		rawStmt:  stmt,
		accessAt: unixTime(),
	}
}

func (this *Stmt) Query(args ...any) (*sql.Rows, error) {
	this.accessAt = unixTime()
	return this.rawStmt.Query(args...)
}

func (this *Stmt) FindOnes(args ...any) (ones []maps.Map, columnNames []string, err error) {
	this.accessAt = unixTime()

	rawRows, err := this.rawStmt.Query(args...)
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

func (this *Stmt) FindOne(args ...any) (one maps.Map, err error) {
	this.accessAt = unixTime()

	rawRows, err := this.rawStmt.Query(args...)
	if err != nil {
		return nil, err
	}

	var rows = NewRows(rawRows)
	defer func() {
		_ = rows.Close()
	}()

	return rows.FindOne()
}

func (this *Stmt) FindCol(colIndex int, args ...any) (colValue any, err error) {
	this.accessAt = unixTime()

	rawRows, err := this.rawStmt.Query(args...)
	if err != nil {
		return nil, err
	}

	var rows = NewRows(rawRows)
	defer func() {
		_ = rows.Close()
	}()

	return rows.FindCol(colIndex)
}

func (this *Stmt) Exec(args ...any) (sql.Result, error) {
	this.accessAt = unixTime()
	return this.rawStmt.Exec(args...)
}

// Close 关闭
func (this *Stmt) Close() error {
	return this.rawStmt.Close()
}

// Raw 获取原始的语句
func (this *Stmt) Raw() *sql.Stmt {
	return this.rawStmt
}

// AccessAt 获得访问时间
// 用来对比是否可以释放
func (this *Stmt) AccessAt() int64 {
	return this.accessAt
}
