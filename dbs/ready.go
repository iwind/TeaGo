package dbs

import "sync"

var readyCallbacks = []func(){}
var readyLocker = sync.Mutex{}

// 添加Ready的回调函数
func OnReady(f func()) {
	readyLocker.Lock()
	if f != nil {
		readyCallbacks = append(readyCallbacks, f)
	}
	readyLocker.Unlock()
}

// 调用Ready回调
func NotifyReady() {
	readyLocker.Lock()
	for _, f := range readyCallbacks {
		f()
	}
	readyLocker.Unlock()
}
