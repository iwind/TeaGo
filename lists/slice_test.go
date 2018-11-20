package lists

import (
	"fmt"
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestSliceContains(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	t.Log(Contains(slice, 10))
	t.Log(Contains(slice, 2))
}

func TestSliceContains2(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	t.Log(Contains(slice, "f"))
	t.Log(Contains(slice, "e"))
}

func TestSliceContainsAll(t *testing.T) {
	a := assert.NewAssertion(t).Quiet()
	a.IsFalse(ContainsAll([]string{"a", "b", "c"}))
	a.IsTrue(ContainsAll([]string{"a", "b", "c"}, "a", "b"))
	a.IsFalse(ContainsAll([]string{"a", "b", "c"}, "a", "b", "c", "d"))
	a.IsFalse(ContainsAll([]string{"a", "b", "c"}, "d"))
	a.IsTrue(ContainsAll([]string{"a", "b", "c"}, "b"))
}

func TestSliceContainsAny(t *testing.T) {
	a := assert.NewAssertion(t).Quiet()
	a.IsFalse(ContainsAny([]string{"a", "b", "c"}))
	a.IsTrue(ContainsAny([]string{"a", "b", "c"}, "a", "b"))
	a.IsTrue(ContainsAny([]string{"a", "b", "c"}, "a", "b", "c", "d"))
	a.IsFalse(ContainsAny([]string{"a", "b", "c"}, "d"))
	a.IsTrue(ContainsAny([]string{"a", "b", "c"}, "b"))
}

func TestSliceMap(t *testing.T) {
	s := []string{"a", "b", "c"}
	t.Log(Map(s, func(k int, v interface{}) interface{} {
		return fmt.Sprintf("%d:%s", k, v)
	}))
}

func TestSliceDelete(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	n := Delete(slice, "b").([]string)
	t.Log(n)

	slice2 := []int{1, 2, 3, 4, 5}
	n2 := Delete(slice2, 3)
	t.Log(n2)
}

func TestSliceReverse(t *testing.T) {
	a := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	t.Log(a)

	Reverse(a)
	t.Log(a)
}
