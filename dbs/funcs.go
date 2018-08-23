package dbs

import (
	"strings"
)

type DBFunc struct {
	Clause string
	Values []interface{}
}

type DBFuncParams []interface{}

func (fn *DBFunc) prepareForQuery(query *Query) string {
	pieces := strings.Split(fn.Clause, "?")
	if len(pieces) == 1 {
		return fn.Clause
	}
	count := len(pieces)
	valuesCount := len(fn.Values)
	for i := 1; i < count; i ++ {
		valueIndex := i - 1
		if valueIndex >= valuesCount {
			pieces[i] = "NULL" + pieces[i]
		} else {
			pieces[i] = query.wrapValue(fn.Values[valueIndex]) + pieces[i]
		}
	}
	return strings.Join(pieces, "")
}

func FuncAbs(x interface{}) *DBFunc {
	return &DBFunc{
		Clause: "ABS(?)",
		Values: DBFuncParams{x},
	}
}

func FuncAcos(x interface{}) *DBFunc {
	return &DBFunc{
		Clause: "ACOS(?)",
		Values: DBFuncParams{x},
	}
}

func FuncAsin(x interface{}) *DBFunc {
	return &DBFunc{
		Clause: "ASIN(?)",
		Values: DBFuncParams{x},
	}
}

func FuncAtan(x interface{}) *DBFunc {
	return &DBFunc{
		Clause: "ATAN(?)",
		Values: DBFuncParams{x},
	}
}

func FuncAtanXY(x interface{}, y interface{}) *DBFunc {
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

func FuncRandN(seed interface{}) *DBFunc {
	return &DBFunc{
		Clause: "RAND(?)",
		Values: DBFuncParams{seed},
	}
}

func FuncAscii(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "ASCII(?)",
		Values: DBFuncParams{str},
	}
}

func FuncBin(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "Bin(?)",
		Values: DBFuncParams{str},
	}
}

func FuncBitLength(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "BIT_LENGTH(?)",
		Values: DBFuncParams{str},
	}
}

func FuncConcat(strs ... interface{}) *DBFunc {
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

func FuncConcatWs(separator interface{}, strs ... interface{}) *DBFunc {
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

func FuncFindInSet(str interface{}, strList interface{}) *DBFunc {
	return &DBFunc{
		Clause: "FIND_IN_SET(?, ?)",
		Values: DBFuncParams{str, strList},
	}
}

func FuncLeft(str interface{}, len interface{}) *DBFunc {
	return &DBFunc{
		Clause: "LEFT(?, ?)",
		Values: DBFuncParams{str, len},
	}
}

func FuncLength(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "LENGTH(?)",
		Values: DBFuncParams{str},
	}
}

func FuncLower(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "LOWER(?)",
		Values: DBFuncParams{str},
	}
}

func FuncUpper(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "UPPER(?)",
		Values: DBFuncParams{str},
	}
}

func FuncLpad(str interface{}, len interface{}, padStr interface{}) *DBFunc {
	return &DBFunc{
		Clause: "LPAD(?, ?, ?)",
		Values: DBFuncParams{str, len, padStr},
	}
}

func FuncLtrim(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "LTRIM(?)",
		Values: DBFuncParams{str},
	}
}

func FuncRepeat(str interface{}, count interface{}) *DBFunc {
	return &DBFunc{
		Clause: "REPEAT(?, ?)",
		Values: DBFuncParams{str, count},
	}
}

func FuncReplace(str interface{}, fromStr interface{}, toStr interface{}) *DBFunc {
	return &DBFunc{
		Clause: "REPLACE(?, ?, ?)",
		Values: DBFuncParams{str, fromStr, toStr},
	}
}

func FuncReverse(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "REVERSE(?)",
		Values: DBFuncParams{str},
	}
}

func FuncRight(str interface{}, len interface{}) *DBFunc {
	return &DBFunc{
		Clause: "RIGHT(?, ?)",
		Values: DBFuncParams{str, len},
	}
}

func FuncRpad(str interface{}, len interface{}, padStr interface{}) *DBFunc {
	return &DBFunc{
		Clause: "RPAD(?, ?, ?)",
		Values: DBFuncParams{str, len, padStr},
	}
}

func FuncRtrim(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "RTRIM(?)",
		Values: DBFuncParams{str},
	}
}

func FuncSubstring(str interface{}, pos interface{}) *DBFunc {
	return &DBFunc{
		Clause: "SUBSTRING(?, ?)",
		Values: DBFuncParams{str, pos},
	}
}

func FuncSubstringLen(str interface{}, pos interface{}, len interface{}) *DBFunc {
	return &DBFunc{
		Clause: "SUBSTRING(?, ?, ?)",
		Values: DBFuncParams{str, pos, len},
	}
}

func FuncSubstringIndex(str interface{}, delim interface{}, count interface{}) *DBFunc {
	return &DBFunc{
		Clause: "SUBSTRING_INDEX(?, ?, ?)",
		Values: DBFuncParams{str, delim, count},
	}
}

func FuncTrim(str interface{}) *DBFunc {
	return &DBFunc{
		Clause: "TRIM(?)",
		Values: DBFuncParams{str},
	}
}

func FuncFromUnixtime(timestamp interface{}) *DBFunc {
	return &DBFunc{
		Clause: "FROM_UNIXTIME(?)",
		Values: DBFuncParams{timestamp},
	}
}

func FuncFromUnixtimeFormat(timestamp interface{}, format interface{}) *DBFunc {
	return &DBFunc{
		Clause: "FROM_UNIXTIME(?, ?)",
		Values: DBFuncParams{timestamp, format},
	}
}
