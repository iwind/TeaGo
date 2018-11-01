package lists

import (
	"reflect"
	"sort"
)

// 对slice进行排序
func Sort(slice interface{}, compareFn func(i int, j int) bool) {
	newSlice := &List{
		compareFunc: compareFn,
		Slice:       slice,
	}
	sort.Sort(newSlice)
}

// 对slice进行反转操作
func Reverse(slice interface{}) {
	newSlice := &List{
		compareFunc: nil,
		Slice:       slice,
	}
	newSlice.Reverse()
}

// 判断slice是否包含某个item
func Contains(slice interface{}, item interface{}) bool {
	value := reflect.ValueOf(slice)
	size := value.Len()
	for i := 0; i < size; i++ {
		currentItemValue := value.Index(i)
		if currentItemValue.Interface() == item {
			return true
		}
	}
	return false
}

// 判断slice是否包含所有的item
// 如果没有任何item，则返回false
func ContainsAll(slice interface{}, item ...interface{}) bool {
	if len(item) == 0 {
		return false
	}

	m := map[interface{}]bool{}
	value := reflect.ValueOf(slice)
	size := value.Len()
	for i := 0; i < size; i++ {
		currentItemValue := value.Index(i)
		m[currentItemValue.Interface()] = true
	}

	for _, item1 := range item {
		_, found := m[item1]
		if !found {
			return false
		}
	}
	return true
}

// 判断slice是否包含任意一个item
// 如果没有任何item，则返回false
func ContainsAny(slice interface{}, item ...interface{}) bool {
	if len(item) == 0 {
		return false
	}

	m := map[interface{}]bool{}
	value := reflect.ValueOf(slice)
	size := value.Len()
	for i := 0; i < size; i++ {
		currentItemValue := value.Index(i)
		m[currentItemValue.Interface()] = true
	}

	for _, item1 := range item {
		_, found := m[item1]
		if found {
			return true
		}
	}
	return false
}

// 删除slice中的某个元素值
// 不限制删除的元素个数
func Delete(slice interface{}, item interface{}) interface{} {
	value := reflect.ValueOf(slice)
	size := value.Len()

	newSlice := reflect.Indirect(reflect.New(value.Type()))

	for i := 0; i < size; i++ {
		currentItem := value.Index(i)
		if currentItem.Interface() == item {
			continue
		}

		newSlice = reflect.Append(newSlice, currentItem)
	}

	return newSlice.Interface()
}

// 从slice中移除某个位置上的元素
func Remove(slice interface{}, index int) interface{} {
	value := reflect.ValueOf(slice)
	size := value.Len()
	if index < 0 {
		index = size + index + 1
	}
	if index >= size {
		return slice
	}

	newValue := reflect.AppendSlice(value.Slice(0, index), value.Slice(index+1, size))
	return newValue.Interface()
}

// 将slice映射为一个新的slice
func Map(slice interface{}, mapFunc func(k int, v interface{}) interface{}) []interface{} {
	value := reflect.ValueOf(slice)
	size := value.Len()
	result := []interface{}{}

	for i := 0; i < size; i++ {
		v := value.Index(i)
		result = append(result, mapFunc(i, v.Interface()))
	}

	return result
}
