package sessions

import (
	"testing"
	"github.com/iwind/TeaGo/actions"
	"time"
)

func TestMemorySessionManager_Read(t *testing.T) {
	var manager = &MemorySessionManager{}
	manager.Init(&actions.SessionConfig{
		Life: 5,
	})

	t.Log(manager.WriteItem("123456", "name", "liu"))
	t.Log(manager.WriteItem("123456", "age", "20"))
	//t.Log(manager.Delete("123456"))

	time.Sleep(6 * time.Second)
	t.Log(manager.Read("123456"))

	time.Sleep(600 * time.Second)
}
