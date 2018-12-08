package caches

import (
	"testing"
	"time"
)

func TestOnce(t *testing.T) {
	type OtherObject struct {
		CacheFactory
	}

	o := new(OtherObject)

	b := time.Now()
	for i := 0; i < 10000; i++ {
		go o.Cache()
	}
	t.Log(time.Since(b).Seconds()*1000, "ms")
}

func TestFactory_Integrate(t *testing.T) {
	type OtherObject struct {
		CacheFactory
	}

	var o = new(OtherObject)
	t.Log(o.Cache().Set("hello", "world"))
	t.Log(o.Cache().Get("hello"))
}
