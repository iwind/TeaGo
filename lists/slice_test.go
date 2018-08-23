package lists

import (
	"testing"
)

func TestContains(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	t.Log(Contains(slice, 10))
	t.Log(Contains(slice, 2))
}

func TestContains2(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	t.Log(Contains(slice, "f"))
	t.Log(Contains(slice, "e"))
}
