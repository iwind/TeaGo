// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbs

import "encoding/json"

type JSON []byte

func (this JSON) UnmarshalTo(ptr interface{}) error {
	return json.Unmarshal(this, ptr)
}

func (this JSON) IsNotNull() bool {
	if len(this) == 0 {
		return false
	}
	if len(this) == 4 && string(this) == "null" {
		return false
	}
	return true
}

func (this JSON) Len() int {
	return len(this)
}

func (this JSON) String() string {
	return string(this)
}

func (this JSON) Bytes() []byte {
	return this
}
