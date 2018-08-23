package maps

import (
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
)

type OrderedMap struct {
	keys      []interface{}
	valuesMap map[interface{}]interface{}
}

func NewOrderedMap() *OrderedMap {
	var orderedMap = &OrderedMap{}
	orderedMap.valuesMap = map[interface{}]interface{}{}
	return orderedMap
}

// 取得所有Key
func (this *OrderedMap) Keys() []interface{} {
	return this.keys
}

// 根据元素值进行排序
func (this *OrderedMap) Sort() {
	lists.Sort(this.keys, func(i int, j int) bool {
		value1 := this.valuesMap[this.keys[i]]
		value2 := this.valuesMap[this.keys[j]]

		return !types.Compare(value1, value2)
	})
}

// 根据Key进行排序
func (this *OrderedMap) SortKeys() {
	lists.Sort(this.keys, func(i int, j int) bool {
		key1 := this.keys[i]
		key2 := this.keys[j]

		return !types.Compare(key1, key2)
	})
}

// 翻转键
func (this *OrderedMap) Reverse() {
	lists.Sort(this.keys, func(i int, j int) bool {
		return i > j
	})
}

// 添加元素
func (this *OrderedMap) Put(key interface{}, value interface{}) {
	_, ok := this.valuesMap[key]
	if !ok {
		this.keys = append(this.keys, key)
	}
	this.valuesMap[key] = value
}

// 取得元素值
func (this *OrderedMap) Get(key interface{}) (value interface{}, ok bool) {
	value, ok = this.valuesMap[key]
	return
}

// 删除元素
func (this *OrderedMap) Delete(key interface{}) {
	var index = -1
	for itemIndex, itemKey := range this.keys {
		if itemKey == key {
			index = itemIndex
			break
		}
	}
	if index > -1 {
		this.keys = append(this.keys[0:index], this.keys[index+1:]...)
		delete(this.valuesMap, key)
	}
}

// 对每个元素执行迭代器
func (this *OrderedMap) Range(iterator func(key interface{}, value interface{})) {
	for _, key := range this.keys {
		iterator(key, this.valuesMap[key])
	}
}

// 取得Map的长度
func (this *OrderedMap) Len() int {
	return len(this.keys)
}
