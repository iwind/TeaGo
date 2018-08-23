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
