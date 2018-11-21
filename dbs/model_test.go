package dbs

import (
	"fmt"
	"testing"
)

func TestMakeModel(t *testing.T) {
	type User struct {
		Id        int `field:"id"`
		Gender    int
		Age       int
		Nickname  string
		CreatedAt int `field:"created_at"`
	}

	var model = NewModel(new(User))
	t.Log(fmt.Sprintf("%#v", model))
}
