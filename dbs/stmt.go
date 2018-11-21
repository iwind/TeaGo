package dbs

import (
	"database/sql"
	"errors"
	"github.com/iwind/TeaGo/maps"
	"reflect"
	"sync"
)

type Stmt struct {
	sqlStmt *sql.Stmt
	locker  *sync.Mutex
}

func BuildStmt(stmt *sql.Stmt, err error) (*Stmt, error) {
	if err != nil {
		return nil, err
	}
	var newStmt = &Stmt{}
	newStmt.sqlStmt = stmt
	return newStmt, err
}

func (this *Stmt) Query(args ...interface{}) (*sql.Rows, error) {
	return this.sqlStmt.Query(args...)
}

func (this *Stmt) FindOnes(args ...interface{}) (results []maps.Map, columnNames []string, err error) {
	if this.sqlStmt == nil {
		return nil, nil, errors.New("stmt not be prepared")
	}

	rows, err := this.sqlStmt.Query(args...)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

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
				var valueType = reflect.TypeOf(value).Kind()
				if valueType == reflect.Slice {
					value = string(value.([]byte))
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

	defer rows.Close()

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
			var valueType = reflect.TypeOf(value).Kind()
			if valueType == reflect.Slice {
				value = string(value.([]byte))
			}
		}
		return value, nil
	}

	return nil, nil
}

func (this *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	return this.sqlStmt.Exec(args...)
}

func (this *Stmt) Close() error {
	return this.sqlStmt.Close()
}

func (this *Stmt) OriginStmt() *sql.Stmt {
	return this.sqlStmt
}
