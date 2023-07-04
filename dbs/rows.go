// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dbs

import (
	"database/sql"
	"errors"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

type Rows struct {
	rawRows *sql.Rows
}

func NewRows(rawRows *sql.Rows) *Rows {
	return &Rows{
		rawRows: rawRows,
	}
}

func (this *Rows) Columns() ([]string, error) {
	return this.rawRows.Columns()
}

func (this *Rows) Close() error {
	return this.rawRows.Close()
}

func (this *Rows) FindOnes() (ones []maps.Map, err error) {
	var columnNames []string
	columnNames, err = this.Columns()
	if err != nil {
		return
	}

	var countColumns = len(columnNames)
	var valuePointers = []any{}
	for i := 0; i < countColumns; i++ {
		var v any
		valuePointers = append(valuePointers, &v)
	}

	for this.rawRows.Next() {
		err = this.rawRows.Scan(valuePointers...)
		if err != nil {
			return
		}

		var rowMap = maps.Map{}
		for i := 0; i < countColumns; i++ {
			var pointer = valuePointers[i]
			var value = *(pointer.(*any))

			if value != nil {
				v, isBytes := value.([]byte)
				if isBytes {
					value = string(v)
				}
			}

			rowMap[columnNames[i]] = value
		}

		ones = append(ones, rowMap)
	}

	// retrieve error in iteration
	err = this.rawRows.Err()

	return
}

func (this *Rows) FindOne() (one maps.Map, err error) {
	var columnNames []string
	columnNames, err = this.Columns()
	if err != nil {
		return
	}

	var countColumns = len(columnNames)
	var valuePointers = []any{}
	for i := 0; i < countColumns; i++ {
		var v any
		valuePointers = append(valuePointers, &v)
	}

	if this.rawRows.Next() { // once only loop
		err = this.rawRows.Scan(valuePointers...)
		if err != nil {
			return
		}

		var rowMap = maps.Map{}
		for i := 0; i < countColumns; i++ {
			var pointer = valuePointers[i]
			var value = *(pointer.(*any))

			if value != nil {
				v, isBytes := value.([]byte)
				if isBytes {
					value = string(v)
				}
			}

			rowMap[columnNames[i]] = value
		}
		one = rowMap
	}

	// retrieve error in iteration
	err = this.rawRows.Err()

	return
}

func (this *Rows) FindCol(colIndex int) (colValue any, err error) {
	var columnNames []string
	columnNames, err = this.Columns()
	if err != nil {
		return
	}

	var countColumns = len(columnNames)
	if colIndex < 0 || colIndex >= countColumns {
		return nil, errors.New("invalid column index '" + types.String(colIndex) + "'")
	}

	var valuePointers = []any{}
	for i := 0; i < countColumns; i++ {
		var v any
		valuePointers = append(valuePointers, &v)
	}

	if this.rawRows.Next() { // once only loop
		err = this.rawRows.Scan(valuePointers...)
		if err != nil {
			return
		}

		var pointer = valuePointers[colIndex]
		colValue = *(pointer.(*any))

		if colValue != nil {
			v, isBytes := colValue.([]byte)
			if isBytes {
				colValue = string(v)
			}
		}
	}

	// retrieve error in iteration
	err = this.rawRows.Err()

	return
}
