package rands

import (
	"runtime"
	"testing"
)

func TestRand_String(t *testing.T) {
	t.Log(String(32))
	t.Log(String(32))
	t.Log(String(32))
	t.Log(String(0))
	t.Log(String(64))
}

func TestRand_HexString(t *testing.T) {
	t.Log(HexString(32))
	t.Log(HexString(32))
	t.Log(HexString(32))
	t.Log(HexString(0))
	t.Log(HexString(64))
}

func TestRand_UniqueString(t *testing.T) {
	m := map[string]bool{}
	for i := 0; i < 1000_0000; i++ {
		s := String(32)
		_, ok := m[s]
		if ok {
			t.Fatal("duplicated:", s)
		}
		m[s] = true
	}
	t.Log("ok")
}

func BenchmarkRand_String(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_ = String(32)
	}
}
