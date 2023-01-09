package types

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
)

func TestConvert(t *testing.T) {
	t.Log(Int(123))
	t.Log(Int("123.456"))
	t.Log(Bool("abc"), Bool(123), Bool(false), Bool(true))
	t.Log(Float32("123.456"))
	t.Log(Compare("abc", "123"), Compare(123, "12.3"))
	t.Log(Byte(123), Byte(255))
	t.Log(String(1), String(int64(1024)), String(true), String("Hello, World"), String([]string{"Hello"}))

	result, err := Slice([]string{"1", "2", "3"}, reflect.TypeOf([]int64{}))
	if err != nil {
		t.Log("fail to convert slice")
	} else {
		t.Logf("%#v", result)
	}
}

func TestIntN(t *testing.T) {
	assert(t, Int8("1") == 1)
	assert(t, Int8("1024") == math.MaxInt8)
	assert(t, Int8("-1024") == math.MinInt8)

	assert(t, Int16("1") == 1)
	assert(t, Int16(1024) == 1024)
	assert(t, Int16(-1024) == -1024)
	assert(t, Int16(123456789101112) == math.MaxInt16)

	assert(t, Int32("1") == 1)
	assert(t, Int32(1024) == 1024)
	assert(t, Int32(-1024) == -1024)
	assert(t, Int32(123456789101112) == math.MaxInt32)
	t.Log("maxInt32:", math.MaxInt32)

	{

		type A int32
		var a A = 1234
		assert(t, Int32(a) == 1234)
	}
	{
		type A float32
		var a A = 123.456
		assert(t, Int32(a) == 123)
	}

	assert(t, Int64("1") == 1)
	assert(t, Int64(1024) == 1024)
	assert(t, Int64(-1024) == -1024)
	assert(t, Int64(9223372036854775807) == math.MaxInt64)
	t.Log("maxInt64:", math.MaxInt64)

	assert(t, Uint8(123) == 123)
	assert(t, Uint8(1024) == math.MaxUint8)
	t.Log("maxUint8:", math.MaxUint8)

	assert(t, Uint16(123) == 123)
	assert(t, Uint16(65536) == math.MaxUint16)
	t.Log("maxUint16:", math.MaxUint16)

	assert(t, Uint64(123) == 123)
}

func TestString(t *testing.T) {
	t.Log(String(123))
	t.Log(String(123456))
	t.Log(String(123456.123456))
	t.Log(String(123456890123456))
}

func TestIsSlice(t *testing.T) {
	assert(t, !IsSlice(nil))

	{
		var s []string = nil
		assert(t, IsSlice(s))
	}

	{
		var s interface{} = nil
		assert(t, !IsSlice(s))
	}

	{
		var s *[]string = nil
		assert(t, !IsSlice(s))
	}

	{
		assert(t, IsSlice([]string{"a", "b", "c"}))
	}
}

func TestIsMap(t *testing.T) {
	assert(t, !IsMap(nil))

	{
		var s map[string]interface{} = nil
		assert(t, IsMap(s))
	}

	{
		var s interface{} = nil
		assert(t, !IsMap(s))
	}

	{
		assert(t, IsMap(map[string]interface{}{
			"a": "b",
		}))
	}
}

func assert(t *testing.T, b bool) {
	if !b {
		t.Fail()
	}
}

func BenchmarkInt_String(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_ = String(1024)
	}
}

func BenchmarkInt_Sprintf(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%d", 1024)
	}
}
