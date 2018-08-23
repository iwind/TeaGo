package maps

import (
	"testing"
)

func TestNewOrderedMap(t *testing.T) {
	var om = NewOrderedMap()
	om.Put("Name1", "Zhang San")
	om.Put("Name2", "Li Si")
	om.Put("Name3", "Wang Er")
	om.Put("Name3", "Wang Er 2")
	om.Put("Name4", "Ma Zi")

	for _, key := range om.Keys() {
		value, ok := om.Get(key)
		if !ok {
			continue
		}
		t.Log(key, ":", value)
	}
}

func TestOrderedMap_Range(t *testing.T) {
	var om = NewOrderedMap()
	om.Put("Name1", "Zhang San")
	om.Put("Name2", "Li Si")
	om.Put("Name3", "Wang Er")
	om.Put("Name3", "Wang Er 2")
	om.Put("Name4", "Ma Zi")
	//om.Delete("Name5")
	//om.Delete("Name1")
	om.Range(func(key interface{}, value interface{}) {
		t.Log(key, value)
	})
}

func TestOrderedMap_Sort(t *testing.T) {
	var om = NewOrderedMap()
	om.Put("Name2", "Li Si")
	om.Put("Name1", "Zhang San")
	om.Put("Name5", "Wang Er")
	om.Put("Name4", "Ma Zi")
	om.Put("Name3", "Wang Er 2")
	om.Sort()
	om.Range(func(key interface{}, value interface{}) {
		t.Log(key, value)
	})
}

func TestOrderedMap_SortKeys(t *testing.T) {
	var om = NewOrderedMap()
	om.Put("Name2", "Li Si")
	om.Put("Name1", "Zhang San")
	om.Put("Name5", "Wang Er")
	om.Put("Name4", "Ma Zi")
	om.Put("Name3", "Wang Er 2")
	om.SortKeys()
	om.Reverse()
	om.Range(func(key interface{}, value interface{}) {
		t.Log(key, value)
	})
}

func TestOrderedMap_Reverse(t *testing.T) {
	var om = NewOrderedMap()
	om.Put("Name1", "Zhang San")
	om.Put("Name2", "Li Si")
	om.Put("Name3", "Wang Er")
	om.Put("Name3", "Wang Er 2")
	om.Put("Name4", "Ma Zi")
	om.Reverse()
	om.Range(func(key interface{}, value interface{}) {
		t.Log(key, value)
	})
}