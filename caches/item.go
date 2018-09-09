package caches

import "time"

// 缓存条目定义
type Item struct {
	key        string
	value      interface{}
	expireTime time.Time
}

// 设置值
func (this *Item) Set(value interface{}) *Item {
	this.value = value
	return this
}

// 设置过期时间
func (this *Item) ExpireAt(expireTime time.Time) *Item {
	this.expireTime = expireTime
	return this
}

// 设置过期时长
func (this *Item) Expire(duration time.Duration) *Item {
	return this.ExpireAt(time.Now().Add(duration))
}

// 判断是否已过期
func (this *Item) IsExpired() bool {
	return time.Since(this.expireTime) > 0
}
