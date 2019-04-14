package caches

import (
	"github.com/iwind/TeaGo/timers"
	"sync"
	"time"
)

// 操作类型
type CacheOperation = string

const (
	CacheOperationSet    = "set"
	CacheOperationDelete = "delete"
)

// 缓存管理器
type Factory struct {
	items       map[string]*Item
	maxSize     int64                               // @TODO 实现maxSize
	onOperation func(op CacheOperation, item *Item) // 操作回调
	locker      *sync.Mutex
	looper      *timers.Looper
}

// 创建一个新的缓存管理器
func NewFactory() *Factory {
	return newFactoryInterval(30 * time.Second)
}

func newFactoryInterval(duration time.Duration) *Factory {
	factory := &Factory{
		items:  map[string]*Item{},
		locker: &sync.Mutex{},
	}

	factory.looper = timers.Loop(duration, func(looper *timers.Looper) {
		factory.clean()
	})

	return factory
}

// 设置缓存
func (this *Factory) Set(key string, value interface{}, duration ...time.Duration) *Item {
	item := new(Item)
	item.Key = key
	item.Value = value

	if len(duration) > 0 {
		item.expireTime = time.Now().Add(duration[0])
	} else {
		item.expireTime = time.Now().Add(3600 * time.Second)
	}

	this.locker.Lock()
	if this.onOperation != nil {
		_, ok := this.items[key]
		if ok {
			this.onOperation(CacheOperationDelete, item)
		}
	}
	this.items[key] = item
	this.locker.Unlock()

	if this.onOperation != nil {
		this.onOperation(CacheOperationSet, item)
	}

	return item
}

// 获取缓存
func (this *Factory) Get(key string) (value interface{}, found bool) {
	this.locker.Lock()
	defer this.locker.Unlock()

	item, found := this.items[key]
	if !found {
		return nil, false
	}

	if item.IsExpired() {
		return nil, false
	}

	return item.Value, true
}

// 判断是否有缓存
func (this *Factory) Has(key string) bool {
	_, found := this.Get(key)
	return found
}

// 删除缓存
func (this *Factory) Delete(key string) {

	this.locker.Lock()
	item, ok := this.items[key]
	if ok {
		delete(this.items, key)
		if this.onOperation != nil {
			this.onOperation(CacheOperationDelete, item)
		}
	}
	this.locker.Unlock()
}

// 设置操作回调
func (this *Factory) OnOperation(f func(op CacheOperation, item *Item)) {
	this.onOperation = f
}

// 关闭
func (this *Factory) Close() {
	this.locker.Lock()
	defer this.locker.Unlock()

	if this.looper != nil {
		this.looper.Stop()
		this.looper = nil
	}

	this.items = map[string]*Item{}
}

// 重置状态
func (this *Factory) Reset() {
	this.locker.Lock()
	defer this.locker.Unlock()
	this.items = map[string]*Item{}
}

// 清理过期的缓存
func (this *Factory) clean() {
	this.locker.Lock()
	defer this.locker.Unlock()

	for _, item := range this.items {
		if item.IsExpired() {
			delete(this.items, item.Key)

			if this.onOperation != nil {
				this.onOperation(CacheOperationDelete, item)
			}
		}
	}
}
