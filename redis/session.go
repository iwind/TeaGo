package redis

import (
	"github.com/iwind/TeaGo/actions"
	"github.com/iwind/TeaGo/logs"
	"time"
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

func (this *SessionManager) Init(config *actions.SessionConfig) {
	this.config = config
	client, err := Client()
	if err != nil {
		logs.Errorf("%s", err.Error())
		return
	}
	this.client = client
}

func (this *SessionManager) Read(sid string) map[string]string {
	sid = "SESSION_" + sid

	if this.client == nil {
		return map[string]string{}
	}
	result, err := this.client.HGetAll(sid)
	if err != nil {
		return map[string]string{}
	}

	// 延长时间
	// @TODO 设定为 10% 的几率延长时间
	this.client.ExpireAt(sid, time.Now().Add(time.Duration(this.config.Life)*time.Second))

	return result
}

func (this *SessionManager) WriteItem(sid string, key string, value string) bool {
	sid = "SESSION_" + sid

	if this.client == nil {
		return false
	}
	_, err := this.client.HSet(sid, key, value)
	this.client.ExpireAt(sid, time.Now().Add(time.Duration(this.config.Life)*time.Second))
	if err != nil {
		logs.Errorf("%s", err.Error())
		return false
	}
	return true
}

func (this *SessionManager) Delete(sid string) bool {
	sid = "SESSION_" + sid

	if this.client == nil {
		return false
	}
	_, err := this.client.Del(sid)
	if err != nil {
		logs.Errorf("%s", err.Error())
		return false
	}
	return true
}
