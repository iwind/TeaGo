package files

import (
	"testing"
	"github.com/iwind/TeaGo/Tea"
)

func TestWriter_Write(t *testing.T) {
	tmpFile := NewFile(Tea.TmpFile("test.txt"))
	writer, err := tmpFile.Writer()
	if err != nil {
		t.Fatal(err)
	}

	//writer.Write([]byte("Hello,a"))
	//writer.Truncate()

	//writer.Seek(10)
	//writer.WriteString("ba")

	writer.Close()
}
