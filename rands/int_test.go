package rands

import (
	"github.com/iwind/TeaGo/logs"
	"runtime"
	"testing"
)

func TestRand_Distribute_1(t *testing.T) {
	m := map[int]int{} // number => count
	for i := 0; i < 1000000; i++ {
		v := Int(0, 9)
		_, ok := m[v]
		if ok {
			m[v] ++
		} else {
			m[v] = 1
		}
	}
	logs.PrintAsJSON(m, t)
}

func TestRand_Distribute_2(t *testing.T) {
	m := map[int]int{} // number => count
	for i := 0; i < 1000000; i++ {
		v := Int(15, 5)
		_, ok := m[v]
		if ok {
			m[v] ++
		} else {
			m[v] = 1
		}
	}
	logs.PrintAsJSON(m, t)
}

func BenchmarkRandBetween(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_ = Int(0, 100)
	}
}
