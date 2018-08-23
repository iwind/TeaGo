package sessions

import (
	"github.com/iwind/TeaGo/actions"
	"sync"
	"time"
)

type MemorySessionManager struct {
	sessionMap *sync.Map // sid => { expiredAt, items }
	life       uint

	isInitialized bool
}

func NewMemorySessionManager() *MemorySessionManager {
	return &MemorySessionManager{
		sessionMap: &sync.Map{},
		life:       1200,
	}
}

func (this *MemorySessionManager) Init(config *actions.SessionConfig) () {
	if this.isInitialized {
		return
	}

	this.isInitialized = true

	this.life = config.Life
	this.sessionMap = &sync.Map{}

	// 定期清除过期的SESSION
	tick := time.NewTicker(1800 * time.Second)
	go func() {
		for {
			<-tick.C

			this.sessionMap.Range(func(sid, user interface{}) bool {
				userMap := user.(map[string]interface{})
				expiredAt := userMap["expiredAt"].(int64)
				if expiredAt < time.Now().Unix() {
					this.sessionMap.Delete(sid)
				}
				return true
			})
		}
	}()
}

func (this *MemorySessionManager) Read(sid string) map[string]string {
	user, found := this.sessionMap.Load(sid)
	if !found {
		return map[string]string{}
	}

	userMap := user.(map[string]interface{})
	expiredAt := userMap["expiredAt"].(int64)
	if expiredAt < time.Now().Unix() {
		return map[string]string{}
	}

	return userMap["items"].(map[string]string)
}

func (this *MemorySessionManager) WriteItem(sid string, key string, value string) bool {
	items := this.Read(sid)
	items[key] = value
	this.sessionMap.Store(sid, map[string]interface{}{
		"expiredAt": time.Now().Unix() + int64(this.life),
		"items":     items,
	})
	return true
}

func (this *MemorySessionManager) Delete(sid string) bool {
	this.sessionMap.Delete(sid)
	return true
}
