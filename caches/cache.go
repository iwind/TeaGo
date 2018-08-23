package caches

import "sync"

type CacheFactory struct {
	factory *Factory
	locker  sync.Mutex
}

func (this *CacheFactory) Cache() *Factory {
	this.locker.Lock()
	defer this.locker.Unlock()
	if this.factory == nil {
		this.factory = NewFactory()
	}
	return this.factory
}
