package rands

import (
	"math/rand"
	"sync"
	"time"
)

// 生成随机种子
var source = rand.NewSource(time.Now().UnixNano())
var locker = &sync.Mutex{}

// 随机获取一个Int数字
func Int(min int, max int) int {
	if min > max {
		min, max = max, min
	}
	r := max - min + 1
	if r == 0 {
		return min
	}

	locker.Lock()
	result := min + int(source.Int63()%int64(r))
	locker.Unlock()
	return result
}

// 随机获取一个Int64数字
func Int64() int64 {
	locker.Lock()
	result := source.Int63()
	locker.Unlock()
	return result
}
