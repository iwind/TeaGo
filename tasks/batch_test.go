package tasks

import (
	"testing"
)

func TestBatch_Run(t *testing.T) {
	b := NewBatch()
	b.Add(func() {
		t.Log("1")
	})
	b.Add(func() {
		t.Log("2")
	})
	b.Add(func() {
		t.Log("3")
	})
	b.Add(func() {
		t.Log("4")
	})
	b.Run()
	t.Log("done")
}
