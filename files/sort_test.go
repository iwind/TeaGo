package files

import (
	"testing"
	"os"
)

func TestSortSize(t *testing.T) {
	files := NewFile(os.Getenv("GOPATH") + "/src/github.com/iwind/TeaGo").List()

	Sort(files, SortTypeSize)

	for _, file := range files {
		size, _ := file.Size()
		if file.IsDir() {
			t.Log("d:"+file.Name(), size)
		} else {
			t.Log(file.Name(), size)
		}
	}
}

func TestSortKind(t *testing.T) {
	files := NewFile(os.Getenv("GOPATH") + "/src/github.com/iwind/TeaGo").List()

	Sort(files, SortTypeKind)

	for _, file := range files {
		if file.IsDir() {
			t.Log("d:" + file.Name())
		} else {
			t.Log(file.Name())
		}
	}
}

func TestSortKindReverse(t *testing.T) {
	files := NewFile(os.Getenv("GOPATH") + "/src/github.com/iwind/TeaGo").List()

	Sort(files, SortTypeKindReverse)

	for _, file := range files {
		if file.IsDir() {
			t.Log("d:" + file.Name())
		} else {
			t.Log(file.Name())
		}
	}
}
