package actions

import (
	"github.com/iwind/TeaGo/types"
	"fmt"
)

type SessionConfig struct {
	Life   uint
	Secret string
}

type SessionWrapper interface {
	Init(config *SessionConfig)
	Read(sid string) map[string]string
	WriteItem(sid string, key string, value string) bool
	Delete(sid string) bool
}

type Session struct {
	Sid     string
	Manager interface{}
}

func (this *Session) SetSid(sid string) {
	this.Sid = sid
}

func (this *Session) Values() map[string]string {
	return this.Manager.(SessionWrapper).Read(this.Sid)
}

func (this *Session) StringValue(key string) string {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return ""
	}
	return value
}

func (this *Session) IntValue(key string) int {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return int(types.Int32(value))
}

func (this *Session) Float32Value(key string) float32 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Float32(value)
}

func (this *Session) BoolValue(key string) bool {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return false
	}
	return types.Bool(value)
}

func (this *Session) Write(key, value string) bool {
	return this.Manager.(SessionWrapper).WriteItem(this.Sid, key, value)
}

func (this *Session) WriteInt64(key string, value int64) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

func (this *Session) Delete() bool {
	return this.Manager.(SessionWrapper).Delete(this.Sid)
}
