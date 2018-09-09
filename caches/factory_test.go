package caches

import (
	"testing"
	"time"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	factory.Set("hello", "world").ExpireAt(time.Now().Add(10 * time.Second))

	value, found := factory.Get("hello")
	if !found {
		t.Fatal("[ERROR]", "'hello' not found")
	}

	if value != "world" {
		t.Fatal("[ERROR]", "'hello' not equal 'world'")
	}

	t.Log("ok")
}

func TestNewFactory_Clean(t *testing.T) {
	factory := NewFactory()
	factory.Set("hello", "world").ExpireAt(time.Now().Add(-10 * time.Second))
	t.Log(len(factory.items))
	factory.clean()
	t.Log(len(factory.items))
}

func TestNewFactory_CleanLoop(t *testing.T) {
	factory := newFactoryInterval(1 * time.Second)
	factory.Set("hello", "world").ExpireAt(time.Now().Add(2 * time.Second))

	t.Log(factory.items["hello"].expireTime)

	time.Sleep(3 * time.Second)

	t.Log(time.Now())
	t.Log(len(factory.items))
}
