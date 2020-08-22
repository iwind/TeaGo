package actions

import "testing"

func TestData(t *testing.T) {
	data := Data{}
	data["a"] = "b"
	t.Log(data)
}
