package pages

import "testing"

func TestPageInit(t *testing.T) {
	page := New(100, 30, 2)
	t.Logf("size:%d, length:%d, offset:%d, index:%d", page.Size, page.Length, page.Offset, page.Index)
}
