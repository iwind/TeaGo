package lists

import (
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

func TestSliceDelete(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	n := Delete(slice, "b").([]string)
	t.Log(n)

	slice2 := []int{1, 2, 3, 4, 5}
	n2 := Delete(slice2, 3)
	t.Log(n2)
}
