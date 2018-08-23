package dbs

import (
	"testing"
	"github.com/iwind/TeaGo/types"
)

func TestIntValue(t *testing.T) {
	t.Log(types.Int64(123))
	t.Log(types.Int64(int32(123)))
	t.Log(types.Int64(123.45678901))
	t.Log(types.Int64("1234567"))
	t.Log(types.Int64("123.456"))
}
