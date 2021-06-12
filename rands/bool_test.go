// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package rands

import (
	"testing"
)

func TestRand_Bool_Distribute_1(t *testing.T) {
	m := map[bool]int{} // number => count
	for i := 0; i < 1000000; i++ {
		v := Bool()
		_, ok := m[v]
		if ok {
			m[v]++
		} else {
			m[v] = 1
		}
	}
	t.Log(m)
}
