package dbs

import (
	"strings"
)

type DBFunc struct {
	Clause string
	Values []any
}

type DBFuncParams []any

func (fn *DBFunc) prepareForQuery(query *Query) string {
	pieces := strings.Split(fn.Clause, "?")
	if len(pieces) == 1 {
		return fn.Clause
	}
	count := len(pieces)
	valuesCount := len(fn.Values)
	for i := 1; i < count; i++ {
		valueIndex := i - 1
		if valueIndex >= valuesCount {
			pieces[i] = "NULL" + pieces[i]
		} else {
			pieces[i] = query.wrapValue(fn.Values[valueIndex]) + pieces[i]
		}
	}
	return strings.Join(pieces, "")
}

func FuncAbs(x any) *DBFunc {
	return &DBFunc{
		Clause: "ABS(?)",
		Values: DBFuncParams{x},
	}
}

func FuncAcos(x any) *DBFunc {
	return &DBFunc{
		Clause: "ACOS(?)",
		Values: DBFuncParams{x},
	}
}

func FuncAsin(x any) *DBFunc {
	return &DBFunc{
		Clause: "ASIN(?)",
		Values: DBFuncParams{x},
	}
}

func FuncAtan(x any) *DBFunc {
	return &DBFunc{
		Clause: "ATAN(?)",
		Values: DBFuncParams{x},
	}
}

func FuncAtanXY(x any, y any) *DBFunc {
	return &DBFunc{
		Clause: "ATAN(?,?)",
		Values: DBFuncParams{y, x},
	}
}

func FuncRand() *DBFunc {
	return &DBFunc{
		Clause: "RAND()",
		Values: DBFuncParams{},
	}
}

func FuncRandN(seed any) *DBFunc {
	return &DBFunc{
		Clause: "RAND(?)",
		Values: DBFuncParams{seed},
	}
}

func FuncAscii(str any) *DBFunc {
	return &DBFunc{
		Clause: "ASCII(?)",
		Values: DBFuncParams{str},
	}
}

func FuncBin(str any) *DBFunc {
	return &DBFunc{
		Clause: "Bin(?)",
		Values: DBFuncParams{str},
	}
}

func FuncBitLength(str any) *DBFunc {
	return &DBFunc{
		Clause: "BIT_LENGTH(?)",
		Values: DBFuncParams{str},
	}
}

func FuncConcat(strs ...any) *DBFunc {
	params := DBFuncParams{}

	for _, str := range strs {
		params = append(params, str)
	}
	markers := ""
	count := len(strs)
	if count > 0 {
		markers = strings.Repeat("?, ", count-1) + "?"
	}
	return &DBFunc{
		Clause: "CONCAT(" + markers + ")",
		Values: params,
	}
}

func FuncConcatWs(separator any, strs ...any) *DBFunc {
	params := DBFuncParams{separator}

	for _, str := range strs {
		params = append(params, str)
	}
	markers := ""
	count := len(strs)
	if count > 0 {
		markers = "?, " + strings.Repeat("?, ", count-1) + "?"
	}
	return &DBFunc{
		Clause: "CONCAT_WS(" + markers + ")",
		Values: params,
	}
}

func FuncFindInSet(str any, strList any) *DBFunc {
	return &DBFunc{
		Clause: "FIND_IN_SET(?, ?)",
		Values: DBFuncParams{str, strList},
	}
}

func FuncLeft(str any, len any) *DBFunc {
	return &DBFunc{
		Clause: "LEFT(?, ?)",
		Values: DBFuncParams{str, len},
	}
}

func FuncLength(str any) *DBFunc {
	return &DBFunc{
		Clause: "LENGTH(?)",
		Values: DBFuncParams{str},
	}
}

func FuncLower(str any) *DBFunc {
	return &DBFunc{
		Clause: "LOWER(?)",
		Values: DBFuncParams{str},
	}
}

func FuncUpper(str any) *DBFunc {
	return &DBFunc{
		Clause: "UPPER(?)",
		Values: DBFuncParams{str},
	}
}

func FuncLpad(str any, len any, padStr any) *DBFunc {
	return &DBFunc{
		Clause: "LPAD(?, ?, ?)",
		Values: DBFuncParams{str, len, padStr},
	}
}

func FuncLtrim(str any) *DBFunc {
	return &DBFunc{
		Clause: "LTRIM(?)",
		Values: DBFuncParams{str},
	}
}

func FuncRepeat(str any, count any) *DBFunc {
	return &DBFunc{
		Clause: "REPEAT(?, ?)",
		Values: DBFuncParams{str, count},
	}
}

func FuncReplace(str any, fromStr any, toStr any) *DBFunc {
	return &DBFunc{
		Clause: "REPLACE(?, ?, ?)",
		Values: DBFuncParams{str, fromStr, toStr},
	}
}

func FuncReverse(str any) *DBFunc {
	return &DBFunc{
		Clause: "REVERSE(?)",
		Values: DBFuncParams{str},
	}
}

func FuncRight(str any, len any) *DBFunc {
	return &DBFunc{
		Clause: "RIGHT(?, ?)",
		Values: DBFuncParams{str, len},
	}
}

func FuncRpad(str any, len any, padStr any) *DBFunc {
	return &DBFunc{
		Clause: "RPAD(?, ?, ?)",
		Values: DBFuncParams{str, len, padStr},
	}
}

func FuncRtrim(str any) *DBFunc {
	return &DBFunc{
		Clause: "RTRIM(?)",
		Values: DBFuncParams{str},
	}
}

func FuncSubstring(str any, pos any) *DBFunc {
	return &DBFunc{
		Clause: "SUBSTRING(?, ?)",
		Values: DBFuncParams{str, pos},
	}
}

func FuncSubstringLen(str any, pos any, len any) *DBFunc {
	return &DBFunc{
		Clause: "SUBSTRING(?, ?, ?)",
		Values: DBFuncParams{str, pos, len},
	}
}

func FuncSubstringIndex(str any, delim any, count any) *DBFunc {
	return &DBFunc{
		Clause: "SUBSTRING_INDEX(?, ?, ?)",
		Values: DBFuncParams{str, delim, count},
	}
}

func FuncTrim(str any) *DBFunc {
	return &DBFunc{
		Clause: "TRIM(?)",
		Values: DBFuncParams{str},
	}
}

func FuncFromUnixtime(timestamp any) *DBFunc {
	return &DBFunc{
		Clause: "FROM_UNIXTIME(?)",
		Values: DBFuncParams{timestamp},
	}
}

func FuncFromUnixtimeFormat(timestamp any, format any) *DBFunc {
	return &DBFunc{
		Clause: "FROM_UNIXTIME(?, ?)",
		Values: DBFuncParams{timestamp, format},
	}
}
