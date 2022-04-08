package dbs

import (
	"database/sql"
	"github.com/iwind/TeaGo/maps"
)

// Stmt SQL语句
type Stmt struct {
	accessAt int64
	sqlStmt  *sql.Stmt
}

// NewStmt 构造
func NewStmt(stmt *sql.Stmt) *Stmt {
	return &Stmt{
		sqlStmt:  stmt,
		accessAt: unixTime(),
	}
}

func (this *Stmt) Query(args ...interface{}) (*sql.Rows, error) {
	this.accessAt = unixTime()
	return this.sqlStmt.Query(args...)
}

func (this *Stmt) FindOnes(args ...interface{}) (results []maps.Map, columnNames []string, err error) {
	this.accessAt = unixTime()
	rows, err := this.sqlStmt.Query(args...)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	columnNames, err = rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	results = []maps.Map{}

	var countColumns = len(columnNames)
	var valuePointers = []interface{}{}
	for i := 0; i < countColumns; i++ {
		var v interface{}
		valuePointers = append(valuePointers, &v)
	}

	for rows.Next() {
		err := rows.Scan(valuePointers...)
		if err != nil {
			return nil, nil, err
		}

		var rowMap = maps.Map{}
		for i := 0; i < countColumns; i++ {
			var pointer = valuePointers[i]
			var value = *pointer.(*interface{})

			if value != nil {
				v, isBytes := value.([]byte)
				if isBytes {
					value = string(v)
				}
			}

			rowMap[columnNames[i]] = value
		}

		results = append(results, rowMap)
	}

	return results, columnNames, nil
}

func (this *Stmt) FindOne(args ...interface{}) (result maps.Map, err error) {
	ones, _, err := this.FindOnes(args...)
	if err != nil {
		return nil, err
	}

	if len(ones) == 0 {
		return nil, nil
	}

	return ones[0], nil
}

func (this *Stmt) FindCol(colIndex int, args ...interface{}) (interface{}, error) {
	rows, err := this.Query(args...)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var countColumns = len(columnNames)
	if colIndex < 0 || countColumns <= colIndex {
		return nil, nil
	}

	var valuePointers = []interface{}{}
	for i := 0; i < countColumns; i++ {
		var v interface{}
		valuePointers = append(valuePointers, &v)
	}

	if rows.Next() {
		err := rows.Scan(valuePointers...)
		if err != nil {
			return nil, err
		}

		var pointer = valuePointers[colIndex]
		var value = *pointer.(*interface{})

		if value != nil {
			v, isBytes := value.([]byte)
			if isBytes {
				value = string(v)
			}
		}
		return value, nil
	}

	return nil, nil
}

func (this *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	this.accessAt = unixTime()
	return this.sqlStmt.Exec(args...)
}

// Close 关闭
func (this *Stmt) Close() error {
	return this.sqlStmt.Close()
}

// Raw 获取原始的语句
func (this *Stmt) Raw() *sql.Stmt {
	return this.sqlStmt
}

// AccessAt 获得访问时间
// 用来对比是否可以释放
func (this *Stmt) AccessAt() int64 {
	return this.accessAt
}
