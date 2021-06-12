// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package rands

// Bool 生成真假随机值
func Bool() bool {
	return Int(0, 1) == 0
}
