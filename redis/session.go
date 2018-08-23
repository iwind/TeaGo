package redis

import (
	"github.com/iwind/TeaGo/actions"
	"time"
	"github.com/iwind/TeaGo/logs"
)

type SessionManager struct {
	config *actions.SessionConfig
	client *RedisClient
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		config: &actions.SessionConfig{
			Life: 1200,
		},
	}
}

func (manager *SessionManager) Init(config *actions.SessionConfig) {
	manager.config = config
	client, err := Client()
	if err != nil {
		logs.Errorf("%s", err.Error())
		return
	}
	manager.client = client
}

func (manager *SessionManager) Read(sid string) map[string]string {
	sid = "SESSION_" + sid

	if manager.client == nil {
		return map[string]string{}
	}
	result, err := manager.client.HGetAll(sid)
	if err != nil {
		return map[string]string{}
	}

	// 延长时间
	// @TODO 设定为 10% 的几率延长时间
	manager.client.ExpireAt(sid, time.Now().Add(time.Duration(manager.config.Life)*time.Second))

	return result
}

func (manager *SessionManager) WriteItem(sid string, key string, value string) bool {
	sid = "SESSION_" + sid

	if manager.client == nil {
		return false
	}
	_, err := manager.client.HSet(sid, key, value)
	manager.client.ExpireAt(sid, time.Now().Add(time.Duration(manager.config.Life)*time.Second))
	if err != nil {
		logs.Errorf("%s", err.Error())
		return false
	}
	return true
}

func (manager *SessionManager) Delete(sid string) bool {
	sid = "SESSION_" + sid

	if manager.client == nil {
		return false
	}
	_, err := manager.client.Del(sid)
	if err != nil {
		logs.Errorf("%s", err.Error())
		return false
	}
	return true
}
