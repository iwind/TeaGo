package caches

import (
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
	maxSize     int64                                      // @TODO 实现maxSize
	onOperation func(op CacheOperation, value interface{}) // 操作回调
	locker      *sync.Mutex
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

	go func() {
		for {
			time.Sleep(duration)
			factory.clean()
		}
	}()

	return factory
}

// 设置缓存
func (this *Factory) Set(key string, value interface{}, duration ...time.Duration) *Item {
	item := new(Item)
	item.key = key
	item.value = value

	if len(duration) > 0 {
		item.expireTime = time.Now().Add(duration[0])
	} else {
		item.expireTime = time.Now().Add(3600 * time.Second)
	}

	this.locker.Lock()
	this.items[key] = item
	this.locker.Unlock()

	if this.onOperation != nil {
		this.onOperation(CacheOperationSet, value)
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

	return item.value, true
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
			this.onOperation(CacheOperationDelete, item.value)
		}
	}
	this.locker.Unlock()
}

// 设置操作回调
func (this *Factory) OnOperation(f func(op CacheOperation, value interface{})) {
	this.onOperation = f
}

// 清理过期的缓存
func (this *Factory) clean() {
	this.locker.Lock()
	defer this.locker.Unlock()

	for _, item := range this.items {
		if item.IsExpired() {
			delete(this.items, item.key)

			if this.onOperation != nil {
				this.onOperation(CacheOperationDelete, item.value)
			}
		}
	}
}
