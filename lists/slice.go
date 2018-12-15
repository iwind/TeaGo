package lists

import (
	"github.com/iwind/TeaGo/types"
	"reflect"
	"sort"
	"strings"
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

// 删除所有匹配的项
func DeleteIf(slice interface{}, f func(item interface{}) bool) interface{} {
	value := reflect.ValueOf(slice)
	size := value.Len()

	newSlice := reflect.Indirect(reflect.New(value.Type()))

	for i := 0; i < size; i++ {
		currentItem := value.Index(i)
		if f(currentItem.Interface()) {
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

// 将slice映射为一个新的[]string
func MapString(slice interface{}, mapFunc func(k int, v interface{}) interface{}) []string {
	value := reflect.ValueOf(slice)
	size := value.Len()
	result := []string{}

	for i := 0; i < size; i++ {
		v := value.Index(i)
		result = append(result, types.String(mapFunc(i, v.Interface())))
	}

	return result
}

// 过滤slice中的元素
// 只有filterFunc返回true的才会放到结果中
func Filter(slice interface{}, filterFunc func(k int, v interface{}) bool) []interface{} {
	value := reflect.ValueOf(slice)
	size := value.Len()
	result := []interface{}{}

	for i := 0; i < size; i++ {
		v := value.Index(i).Interface()
		b := filterFunc(i, v)
		if b {
			result = append(result, v)
		}
	}

	return result
}

// 获取item所在位置
func Index(slice interface{}, item interface{}) int {
	value := reflect.ValueOf(slice)
	size := value.Len()

	for i := 0; i < size; i++ {
		v := value.Index(i).Interface()
		if v == item {
			return i
		}
	}

	return -1
}

// 使用函数获取item所在位置
func IndexIf(slice interface{}, f func(item interface{}) bool) int {
	value := reflect.ValueOf(slice)
	size := value.Len()

	for i := 0; i < size; i++ {
		v := value.Index(i).Interface()
		if f(v) {
			return i
		}
	}

	return -1
}

// 获取item所在位置，但从末尾开始查找
func LastIndex(slice interface{}, item interface{}) int {
	value := reflect.ValueOf(slice)
	size := value.Len()

	for i := size - 1; i >= 0; i-- {
		v := value.Index(i).Interface()
		if v == item {
			return i
		}
	}

	return -1
}

// 连接对象
func Join(slice interface{}, sep string, mapFunc func(k int, v interface{}) interface{}) string {
	return strings.Join(MapString(slice, mapFunc), sep)
}
