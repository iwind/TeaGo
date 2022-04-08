// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbs

import "testing"

func TestTx_id(t *testing.T) {
	for i := 0; i < 10; i++ {
		var tx = NewTx(nil, nil)
		t.Log(tx.id)
	}
}
