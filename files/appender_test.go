package files

import (
	"testing"
	"github.com/iwind/TeaGo/Tea"
)

func Test_Appender(t *testing.T) {
	tmpFile := NewFile(Tea.TmpFile("test.txt"))
	appender, err := tmpFile.Appender()
	if err != nil {
		t.Fatal(err)
	}

	//appender.Lock()
	appender.Append([]byte("Hello,a"))
	//appender.Truncate()

	appender.AppendString("[ABC]")

	//appender.Unlock()
	t.Log(appender.Close())
}
