package types

import (
	"testing"
	"reflect"
)

func TestConvert(t *testing.T) {
	t.Log(Int(123))
	t.Log(Int("123.456"))
	t.Log(Bool("abc"), Bool(123), Bool(false), Bool(true))
	t.Log(Float32("123.456"))
	t.Log(Compare("abc", "123"), Compare(123, "12.3"))
	t.Log(Byte(123), Byte(255))

	result, err := Slice([]string{"1", "2", "3"}, reflect.TypeOf([]int64{}))
	if err != nil {
		t.Log("fail to convert slice")
	} else {
		t.Logf("%#v", result)
	}
}
