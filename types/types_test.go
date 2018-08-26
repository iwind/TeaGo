package types

import "testing"

func TestConvert(t *testing.T) {
	t.Log(Int(123))
	t.Log(Int("123.456"))
	t.Log(Bool("abc"), Bool(123), Bool(false), Bool(true))
	t.Log(Float32("123.456"))
	t.Log(Compare("abc", "123"), Compare(123, "12.3"))
	t.Log(Byte(123), Byte(255))
}
