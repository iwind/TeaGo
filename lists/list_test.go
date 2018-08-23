package lists

import (
	"testing"
)

func TestListMap(t *testing.T) {
	var list = New([]map[string]interface{}{
		{
			"id":   10,
			"name": "Zhang San",
		},
		{
			"id":   5,
			"name": "Wang Er",
		},
		{
			"id":   4,
			"name": "Ma Zi",
		},
		{
			"id":   8,
			"name": "Li Si",
		},
	})
	ids := list.Map(func(k int, v interface{}) interface{} {
		id := v.(map[string]interface{})["id"]
		v.(map[string]interface{})["id"] = id.(int) * id.(int)
		return v
	})
	t.Log(ids.Slice)
}

func TestListFilter(t *testing.T) {
	var list = New([]map[string]interface{}{
		{
			"id":   10,
			"name": "Zhang San",
		},
		{
			"id":   5,
			"name": "Wang Er",
		},
		{
			"id":   4,
			"name": "Ma Zi",
		},
		{
			"id":   8,
			"name": "Li Si",
		},
		{
			"id":   3,
			"name": "Liu Mang",
		},
	})
	result := list.Filter(func(k int, v interface{}) bool {
		var id = v.(map[string]interface{})["id"].(int)
		return id%2 == 1
	})
	t.Log(result)
	t.Log(result.Slice.([]map[string]interface{}))
}

func TestListSize(t *testing.T) {
	var list = New([]map[string]interface{}{
		{
			"id":   10,
			"name": "Zhang San",
		},
		{
			"id":   5,
			"name": "Wang Er",
		},
		{
			"id":   4,
			"name": "Ma Zi",
		},
		{
			"id":   8,
			"name": "Li Si",
		},
		{
			"id":   3,
			"name": "Liu Mang",
		},
	})

	t.Log(list.Size())
}

func TestList_Push(t *testing.T) {
	var list = New([]string{"a", "b", "c"})
	list.Push("d", "f", "g")
	t.Log(list)
}

func TestList_Insert(t *testing.T) {
	var list = New([]string{"a", "b", "c", "d"})
	list.Insert(-1, "E")
	t.Log(list)
}

func TestList_Pop(t *testing.T) {
	var list = New([]string{"a", "b", "c"})
	result := list.Pop()
	t.Log(result.(string))
	t.Log(list)
}

func TestList_First(t *testing.T) {
	var list = New([]string{"a", "b", "c", "d", "e"})
	t.Log(list.First(), list.Last())
	t.Log(list.Get(1), list.Get(3), list.Get(-1))
	t.Log(list.asJSON())

	t.Log("====clear====")
	list.Clear()
	t.Log(list.Slice)
}

func TestList_Remove(t *testing.T) {
	var list = New([]string{"a", "b", "c", "d", "e"})
	//list.Remove(-2)
	//t.Log(list.Slice)

	list.KeepIf(func(k int, v interface{}) bool {
		return k%2 == 0
	})
	t.Log(list)
}

func TestList_Set(t *testing.T) {
	var list = New([]string{"a", "b", "c"})
	list.Set(1, "e")
	t.Log(list)
}

func TestList_Find(t *testing.T) {
	var list = New([]string{"a", "b", "c"})
	t.Log(list.Find(func(k int, v interface{}) bool {
		return k == 2
	}))
}

func TestList_FindAll(t *testing.T) {
	var list = New([]string{"a", "b", "c", "d", "e"})
	var result = list.FindAll(func(k int, v interface{}) bool {
		return k%2 == 0
	})
	t.Log(result)
	t.Log(result.Slice.([]string)[0])
}

func TestList_Rand(t *testing.T) {
	var list = New([]string{"a", "b", "c", "d", "e"})
	t.Log(list.Rand(3))
	t.Log(list)
}

func TestList_Copy(t *testing.T) {
	var list = New([]string{"a", "b", "c", "d", "e"})
	t.Log(list.Copy())
}
