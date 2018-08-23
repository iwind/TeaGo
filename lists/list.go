package lists

import (
	"reflect"
	"sort"
	"github.com/pquerna/ffjson/ffjson"
	"encoding/json"
	"math/rand"
	"time"
)

type List struct {
	compareFunc func(i, j int) bool
	Slice       interface{}
}

func NewList(slice interface{}) *List {
	return &List{
		Slice: slice,
	}
}

func (list *List) Reverse() {
	list.compareFunc = func(i, j int) bool {
		return i > j
	}
	sort.Sort(list)
}

// 对该List进行排序
func (list *List) Sort(compareFunc func(i, j int) bool) {
	list.compareFunc = compareFunc
	sort.Sort(list)
}

// 遍历List
func (list *List) Each(iterator func(k int, v interface{})) {
	value := reflect.ValueOf(list.Slice)
	count := value.Len()
	for i := 0; i < count; i ++ {
		iterator(i, value.Index(i).Interface())
	}
}

// 对容器中元素应用迭代器,并将每次执行的结果放入新List中
func (list *List) Map(mapFunc func(k int, v interface{}) interface{}) *List {
	value := reflect.ValueOf(list.Slice)

	newValue := reflect.New(value.Type()).Elem()
	result := &List{
		Slice: newValue.Interface(),
	}
	count := value.Len()
	for i := 0; i < count; i ++ {
		v := mapFunc(i, value.Index(i).Interface())
		newValue = reflect.Append(newValue, reflect.ValueOf(v))
	}
	result.Slice = newValue.Interface()
	return result
}

// 同 FindAll()
func (list *List) Filter(filterFunc func(k int, v interface{}) bool) *List {
	return list.FindAll(filterFunc)
}

// 对容器中元素应用迭代器,并判断是否全部返回真
func (list *List) All(iterator func(k int, v interface{}) bool) bool {
	value := reflect.ValueOf(list.Slice)
	count := value.Len()
	if count == 0 {
		return true
	}
	for i := 0; i < count; i ++ {
		v := value.Index(i).Interface()
		if !iterator(i, v) {
			return false
		}
	}
	return true
}

// 对容器中元素应用迭代器,并判断是否至少有一次返回真
func (list *List) Any(iterator func(k int, v interface{}) bool) bool {
	value := reflect.ValueOf(list.Slice)
	count := value.Len()
	if count == 0 {
		return true
	}
	for i := 0; i < count; i ++ {
		v := value.Index(i).Interface()
		if iterator(i, v) {
			return true
		}
	}
	return false
}

func (list *List) Find(iterator func(k int, v interface{}) bool) interface{} {
	value := reflect.ValueOf(list.Slice)
	count := value.Len()
	if count == 0 {
		return nil
	}
	for i := 0; i < count; i ++ {
		v := value.Index(i).Interface()
		if iterator(i, v) {
			return v
		}
	}
	return nil
}

// 对容器中元素应用迭代器,将所有返回真的元素放入一个List中
func (list *List) FindAll(iterator func(k int, v interface{}) bool) *List {
	value := reflect.ValueOf(list.Slice)
	count := value.Len()

	newValue := reflect.New(value.Type()).Elem()
	result := &List{
		Slice: newValue.Interface(),
	}
	if count == 0 {
		return result
	}
	for i := 0; i < count; i ++ {
		v := value.Index(i).Interface()
		if iterator(i, v) {
			newValue = reflect.Append(newValue, reflect.ValueOf(v))
		}
	}
	result.Slice = newValue.Interface()
	return result
}

// 随机截取List片段
func (list *List) Rand(size int) *List {
	newList := list.Copy()
	newList.Shuffle()
	newList.Slice = reflect.ValueOf(newList.Slice).Slice(0, size).Interface()
	return newList
}

// 在尾部加入一个或多个元素
func (list *List) Push(items ... interface{}) {
	value := reflect.ValueOf(list.Slice)

	for _, item := range items {
		value = reflect.Append(value, reflect.ValueOf(item))
	}

	list.Slice = value.Interface()
}

// 在指定位置插入新的元素，index参数支持负值
func (list *List) Insert(index int, v interface{}) {
	value := reflect.ValueOf(list.Slice)
	size := value.Len()

	if index < 0 {
		index = size + index + 1
	}
	if index > size {
		return
	}

	newValue := reflect.MakeSlice(value.Type(), 0, size)
	newValue = reflect.AppendSlice(newValue, value.Slice(0, index))
	newValue = reflect.Append(newValue, reflect.ValueOf(v))
	newValue = reflect.AppendSlice(newValue, value.Slice(index, size))

	list.Slice = newValue.Interface()
}

func (list *List) Pop() interface{} {
	value := reflect.ValueOf(list.Slice)
	size := value.Len()
	if size == 0 {
		return nil
	}
	lastValue := value.Slice(size-1, size)
	newValue := value.Slice(0, size-1)
	list.Slice = newValue.Interface()
	return lastValue.Index(0).Interface()
}

func (list *List) First() interface{} {
	value := reflect.ValueOf(list.Slice)
	size := value.Len()
	if size == 0 {
		return nil
	}
	return value.Slice(0, 1).Index(0).Interface()
}

func (list *List) Last() interface{} {
	value := reflect.ValueOf(list.Slice)
	size := value.Len()
	if size == 0 {
		return nil
	}
	return value.Slice(size-1, size).Index(0).Interface()
}

func (list *List) Get(index int) interface{} {
	value := reflect.ValueOf(list.Slice)
	size := value.Len()
	if size == 0 || index < 0 || index >= size {
		return nil
	}
	return value.Slice(index, index+1).Index(0).Interface()
}

func (list *List) isEmpty() bool {
	return list.Size() == 0
}

func (list *List) Size() int {
	return list.Len()
}

// 删除某个位置上的值
// 支持负值
func (list *List) Remove(index int) {
	value := reflect.ValueOf(list.Slice)
	size := value.Len()
	if index < 0 {
		index = size + index + 1
	}
	if index >= size {
		return
	}

	newValue := reflect.AppendSlice(value.Slice(0, index), value.Slice(index+1, size))
	list.Slice = newValue.Interface()
}

// 从数组中删除某个值
func (list *List) RemoveIf(iterator func(k int, v interface{}) bool) {
	value := reflect.ValueOf(list.Slice)
	size := value.Len()
	if size == 0 {
		return
	}
	newValue := reflect.MakeSlice(value.Type(), 0, size)
	for i := 0; i < size; i ++ {
		itemValue := value.Index(i)
		if !iterator(i, itemValue.Interface()) {
			newValue = reflect.Append(newValue, itemValue)
		}
	}
	list.Slice = newValue.Interface()
}

func (list *List) KeepIf(iterator func(k int, v interface{}) bool) {
	value := reflect.ValueOf(list.Slice)
	size := value.Len()
	if size == 0 {
		return
	}
	newValue := reflect.MakeSlice(value.Type(), 0, size)
	for i := 0; i < size; i ++ {
		itemValue := value.Index(i)
		if iterator(i, itemValue.Interface()) {
			newValue = reflect.Append(newValue, itemValue)
		}
	}
	list.Slice = newValue.Interface()
}

func (list *List) Clear() {
	value := reflect.ValueOf(list.Slice)
	list.Slice = reflect.MakeSlice(value.Type(), 0, 0)
}

// 设置某个索引位置上的值
func (list *List) Set(index int, v interface{}) {
	value := reflect.ValueOf(list.Slice)
	value.Index(index).Set(reflect.ValueOf(v))
}

func (list *List) Shuffle() {
	list.Sort(func(i, j int) bool {
		var source = rand.NewSource(time.Now().UnixNano())
		return source.Int63()%2 == 0
	})
}

func (list *List) Copy() *List {
	newValue := reflect.New(reflect.TypeOf(list.Slice)).Elem()
	newList := &List{
		Slice: newValue.Interface(),
	}
	list.Each(func(k int, v interface{}) {
		newValue = reflect.Append(newValue, reflect.ValueOf(v))
	})
	newList.Slice = newValue.Interface()
	return newList
}

func (list *List) asJSON() (string, error) {
	jsonBytes, err := ffjson.Marshal(list.Slice)
	return string(jsonBytes), err
}

func (list *List) asPrettyJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(list.Slice, "", "   ")
	return string(jsonBytes), err
}

func (list *List) Len() int {
	value := reflect.ValueOf(list.Slice)
	return value.Len()
}

func (list *List) Swap(i, j int) {
	value := reflect.ValueOf(list.Slice)
	item1 := value.Index(i).Interface()
	item2 := value.Index(j).Interface()

	value.Index(i).Set(reflect.ValueOf(item2))
	value.Index(j).Set(reflect.ValueOf(item1))
}

func (list *List) Less(i, j int) bool {
	return list.compareFunc(i, j)
}
