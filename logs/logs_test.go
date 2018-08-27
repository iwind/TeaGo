package logs

import (
	"testing"
	"errors"
)

func TestDump(t *testing.T) {
	var m = map[string]interface{}{
		"name": "Liu",
		"age":  20,
		"book": map[string]interface{}{
			"name":  "Golang",
			"price": 20.00,
		},
	}
	Dump(m)
}

func TestError(t *testing.T) {
	Error(errors.New("This is error!!!"))
}
