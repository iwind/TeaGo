package dbs

import "database/sql"

type Result struct {
	originResult sql.Result
}

func (result *Result) LastInsertId() (int64, error) {
	return result.originResult.LastInsertId()
}

func (result *Result) RowsAffected() (int64, error) {
	return result.originResult.RowsAffected()
}

func NewResult(result sql.Result) *Result {
	var newResult = &Result{}
	newResult.originResult = result
	return newResult
}
