package caches

import "time"

type Item struct {
	key        string
	value      interface{}
	expireTime time.Time
}

func (this *Item) Set(value interface{}) *Item {
	this.value = value
	return this
}

func (this *Item) Expire(expireTime time.Time) *Item {
	this.expireTime = expireTime
	return this
}

func (this *Item) IsExpired() bool {
	return time.Since(this.expireTime) > 0
}
