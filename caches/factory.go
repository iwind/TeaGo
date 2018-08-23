package caches

import (
	"time"
	"sync"
)

type Factory struct {
	items   map[string]*Item
	maxSize int64 // @TODO 实现maxSize
	locker  *sync.Mutex
}

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

func (this *Factory) Set(key string, value interface{}) *Item {
	item := new(Item)
	item.key = key
	item.value = value
	item.expireTime = time.Now().Add(3600 * time.Second)

	this.locker.Lock()
	this.items[key] = item
	this.locker.Unlock()

	return item
}

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

func (this *Factory) Has(key string) bool {
	_, found := this.Get(key)
	return found
}

func (this *Factory) clean() {
	this.locker.Lock()
	defer this.locker.Unlock()

	for _, item := range this.items {
		if item.IsExpired() {
			delete(this.items, item.key)
		}
	}
}
