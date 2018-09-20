package caches

import (
	"time"
	"sync"
)

// 缓存管理器
type Factory struct {
	items   map[string]*Item
	maxSize int64 // @TODO 实现maxSize
	locker  *sync.Mutex
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
func (this *Factory) Set(key string, value interface{}, duration ... time.Duration) *Item {
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

// 清理过期的缓存
func (this *Factory) clean() {
	this.locker.Lock()
	defer this.locker.Unlock()

	for _, item := range this.items {
		if item.IsExpired() {
			delete(this.items, item.key)
		}
	}
}
