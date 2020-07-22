package assert

import (
	"github.com/iwind/TeaGo/maps"
	"reflect"
	"testing"
	"time"
)

func TestAssertion(t *testing.T) {
	a := NewAssertion(t)
	a.
		IsTrue(false, "not true"). // not true not true not true not true not true not true not true not true not true
		IsTrue(false, "false").
		IsTrue(false).
		IsFalse(false).
		IsNil(nil).
		IsNumber("abc").
		IsNumber(123).
		Cost().
		// Fatal("a").
		Log("hello").
		Equals("a", "b").
		Equals("1", "1").
		Match("\\w+", "abc").
		IsMap(maps.Map{}).
		IsSlice([]string{})
}

func TestAssertion_Contains(t *testing.T) {
	a := NewAssertion(t)
	a.Contains([]string{"a", "b", "c"}, "a")
	a.Contains([3]string{"a", "b", "c"}, "a")
	a.Contains(map[string]interface{}{
		"a": "b",
		"b": "c",
		"d": "e",
	}, "b")
}

func TestAssertion_Panic(t *testing.T) {
	a := NewAssertion(t)
	a.
		Panic(func() {
			panic("this is panic")
		})
}

func TestAssertionKind(t *testing.T) {
	a := NewAssertion(t)
	a.
		IsKind("1", reflect.Int).
		IsKind(1, reflect.Int).
		IsKind(1.234, reflect.Float64).
		IsBool(true).
		IsFloat(123.456).
		IsInt(1).
		IsInt32(int32(32)).
		IsString("32").
		IsString(32).
		IsNumber(123).
		IsNaN("123").
		IsFloat(nil)
}

func TestAssertion_Pass(t *testing.T) {
	a := NewAssertion(t)
	a.Fail()
	a.Pass()
}

func TestAssertion_Compare(t *testing.T) {
	a := NewAssertion(t)
	a.
		Gt(123, 10).
		Lt(123, 10).
		Lte(123, 123.0).
		Between(123, 1259, 456)
}

func TestAssertion_NotTimeout(t *testing.T) {
	a := NewAssertion(t)
	a.NotTimeout(3*time.Second, func() {
		time.Sleep(4 * time.Millisecond)
	})
}

func TestAssertion_Map(t *testing.T) {
	a := NewAssertion(t)
	m := maps.Map{
		"name": "lu",
		"age":  20,
	}

	a.IsNil(m["city"])
	a.IsNotNil(m["name"])

	age := m["age"]
	if age != nil {
		a.IsInteger(age)
		a.Gt(age, 1)
	}
}

func TestAssertion_LogJSON(t *testing.T) {
	a := NewAssertion(t)
	a.LogJSON([]string{"a", "b", "c"})
	a.LogJSON("d")
}
