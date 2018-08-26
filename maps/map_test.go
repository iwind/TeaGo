package maps

import "testing"

func TestMap(t *testing.T) {
	m := Map{
		"name":     "Lu",
		"age":      20,
		"price":    123.45,
		"isOnline": true,
	}
	t.Log(m)
	t.Log(m.GetInt("price"))

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
