package lists

import (
	"sort"
	"reflect"
)

func Sort(slice interface{}, compareFn func(i int, j int) bool) {
	newSlice := &List{
		compareFunc: compareFn,
		Slice:       slice,
	}
	sort.Sort(newSlice)
}

func Reverse(slice interface{}) {
	newSlice := &List{
		compareFunc: nil,
		Slice:       slice,
	}
	newSlice.Reverse()
}

func Contains(slice interface{}, item interface{}) bool {
	value := reflect.ValueOf(slice)
	size := value.Len()
	for i := 0; i < size; i ++ {
		currentItemValue := value.Index(i)
		if currentItemValue.Interface() == item {
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

	for i := 0; i < size; i ++ {
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
