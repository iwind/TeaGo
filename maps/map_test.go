package maps

import (
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestMap(t *testing.T) {
	m := Map{
		"name":     "Lu",
		"age":      20,
		"price":    123.45,
		"isOnline": true,
		"bytes":    []byte("123"),
		"bytes2":   "456",
	}

	var a = assert.NewAssertion(t)

	t.Log(m)
	a.IsTrue(m.GetInt("price") == 123, "price=123")

	m["price"] = 234.5670
	t.Log(m.GetInt("price"))

	t.Log(m.Get("price"))
	t.Log(m.GetString("price"))
	t.Log(m.GetBool("isOnline"))
	t.Log(m.GetFloat32("price"))
	t.Log(m.GetFloat64("price"))

	t.Log(123 - 234.123)
	t.Log(m.Increase("price", 20))

	t.Log(m.Keys())
	t.Log(m.Len(), len(m))

	t.Log(m.Has("price"))

	t.Log(m.GetBytes("bytes"))
	t.Log(m.GetBytes("bytes2"))
}

func TestNewMap(t *testing.T) {
	m := NewMap()
	m["name"] = "Lu"
	t.Log(m)
}

func TestMapConvert(t *testing.T) {
	m := NewMap(map[string]interface{}{
		"name":     "Lu",
		"age":      20,
		"price":    123.45,
		"isOnline": true,
	})

	t.Log(NewMap(m))
}

func TestMap_JSON(t *testing.T) {
	m := NewMap(map[string]interface{}{
		"name":     "Lu",
		"age":      20,
		"price":    123.45,
		"isOnline": true,
	})

	t.Log(string(m.AsPrettyJSON()))
}

func TestMap_DecodeJSON(t *testing.T) {
	m := NewMap(map[string]interface{}{
		"name":     "Lu",
		"age":      20,
		"price":    123.45,
		"isOnline": true,
	})
	jsonData := m.AsPrettyJSON()

	m, err := DecodeJSON(jsonData)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(m)
}
