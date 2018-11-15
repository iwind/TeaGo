package sessions

import (
	"github.com/iwind/TeaGo/actions"
	"sync"
	"time"
)

type MemorySessionManager struct {
	sessionMap map[string]map[string]interface{} // sid => { expiredAt, items }
	life       uint

	isInitialized bool

	locker sync.Mutex
}

func NewMemorySessionManager() *MemorySessionManager {
	return &MemorySessionManager{
		sessionMap: map[string]map[string]interface{}{},
		life:       1200,
	}
}

func (this *MemorySessionManager) Init(config *actions.SessionConfig) {
	if this.isInitialized {
		return
	}

	this.isInitialized = true

	this.life = config.Life
	this.sessionMap = map[string]map[string]interface{}{}

	// 定期清除过期的SESSION
	tick := time.NewTicker(1800 * time.Second)
	go func() {
		for {
			<-tick.C

			this.locker.Lock()
			for sid, user := range this.sessionMap {
				expiredAt := user["expiredAt"].(int64)
				if expiredAt < time.Now().Unix() {
					delete(this.sessionMap, sid)
				}
			}
			this.locker.Unlock()
		}
	}()
}

func (this *MemorySessionManager) Read(sid string) map[string]string {
	this.locker.Lock()
	defer this.locker.Unlock()

	user, found := this.sessionMap[sid]
	if !found {
		return map[string]string{}
	}

	expiredAt := user["expiredAt"].(int64)
	if expiredAt < time.Now().Unix() {
		return map[string]string{}
	}

	return user["items"].(map[string]string)
}

func (this *MemorySessionManager) WriteItem(sid string, key string, value string) bool {
	// 这个要放在lock之前，否则会造成死锁
	items := this.Read(sid)

	this.locker.Lock()
	defer this.locker.Unlock()

	items[key] = value
	this.sessionMap[sid] = map[string]interface{}{
		"expiredAt": time.Now().Unix() + int64(this.life),
		"items":     items,
	}
	return true
}

func (this *MemorySessionManager) Delete(sid string) bool {
	this.locker.Lock()
	defer this.locker.Unlock()

	delete(this.sessionMap, sid)
	return true
}
