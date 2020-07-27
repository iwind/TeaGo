package actions

import (
	"github.com/iwind/TeaGo/maps"
	"sync"
)

// 上下文变量容器
type ActionContext struct {
	context maps.Map
	locker  sync.RWMutex
}

// 获取新对象
func NewActionContext() *ActionContext {
	return &ActionContext{
		context: maps.Map{},
	}
}

// 设置变量
func (this *ActionContext) Set(key string, value interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()
	this.context[key] = value
}

// 获取变量
func (this *ActionContext) Get(key string) interface{} {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.context.Get(key)
}

// 获取string变量
func (this *ActionContext) GetString(key string) string {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.context.GetString(key)
}

// 获取int变量
func (this *ActionContext) GetInt(key string) int {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.context.GetInt(key)
}

//  获取int64变量
func (this *ActionContext) GetInt64(key string) int64 {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.context.GetInt64(key)
}

// 获取bool变量
func (this *ActionContext) GetBool(key string) bool {
	this.locker.RLock()
	defer this.locker.RUnlock()

	return this.context.GetBool(key)
}
