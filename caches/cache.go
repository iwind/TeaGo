package caches

import (
	"sync"
)

type CacheFactory struct {
	factory *Factory
	locker  sync.Once
}

func (this *CacheFactory) Cache() *Factory {
	this.locker.Do(func() {
		this.factory = NewFactory()
	})
	return this.factory
}
