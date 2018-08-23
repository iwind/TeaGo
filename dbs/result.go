package dbs

import "database/sql"

type Result struct {
	originResult sql.Result
}

func NewResult(result sql.Result) *Result {
	var newResult = &Result{}
	newResult.originResult = result
	return newResult
}

func (this *Result) LastInsertId() (int64, error) {
	return this.originResult.LastInsertId()
}

func (this *Result) RowsAffected() (int64, error) {
	return this.originResult.RowsAffected()
}
