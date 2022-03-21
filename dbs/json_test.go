// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbs_test

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestJSON_nil(t *testing.T) {
	{
		var a dbs.JSON = nil

		// should not panic
		t.Log(a.String())
	}
	{
		var a []byte = nil

		// should not panic
		t.Log(dbs.JSON(a).String())
	}
}

func TestJSON_Decode(t *testing.T) {
	type A struct {
		Name string `json:"name"`
	}

	var a = new(A)
	var data = dbs.JSON(`{"name": "Lily"}`)
	err := data.UnmarshalTo(a)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", a)
}

func TestJSON_Len(t *testing.T) {
	{
		var data = dbs.JSON(`{"name": "Lily"}`)
		t.Log(data.Len())
	}

	{
		var data = dbs.JSON(``)
		t.Log(data.Len())
	}
}

func TestJSON_String(t *testing.T) {
	var data = dbs.JSON(`{"name": "Lily"}`)
	t.Log(string(data))
	t.Log(data.String())
}

func TestJSON_IsNotNull(t *testing.T) {
	t.Log(dbs.JSON("").IsNotNull() == false)
	t.Log(dbs.JSON("null").IsNotNull() == false)
	t.Log(dbs.JSON("1").IsNotNull())
}

func TestJSON_IsBytes(t *testing.T) {
	var v interface{} = dbs.JSON("123")
	_, ok := v.([]byte)
	// should be false
	t.Log(ok)
}

func TestJSON_ConvertBytes(t *testing.T) {
	var v = []byte(dbs.JSON("123"))
	t.Log(v)
	t.Log(string(v) == "123")
}

func TestJSON_Bytes(t *testing.T) {
	var v = dbs.JSON("123")
	t.Log(v)
	t.Log(string(v) == "123")
	t.Log(v.Bytes())
}
