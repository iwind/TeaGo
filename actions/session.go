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

func (this *Session) GetString(key string) string {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return ""
	}
	return value
}

func (this *Session) GetInt(key string) int {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Int(value)
}

func (this *Session) GetInt32(key string) int32 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Int32(value)
}

func (this *Session) GetInt64(key string) int64 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Int64(value)
}

func (this *Session) GetUint(key string) uint {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Uint(value)
}

func (this *Session) GetUint32(key string) uint32 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Uint32(value)
}

func (this *Session) GetUint64(key string) uint64 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Uint64(value)
}

func (this *Session) GetFloat32(key string) float32 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Float32(value)
}

func (this *Session) GetFloat64(key string) float64 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Float64(value)
}

func (this *Session) GetBool(key string) bool {
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

func (this *Session) WriteInt(key string, value int) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

func (this *Session) WriteInt32(key string, value int32) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

func (this *Session) WriteInt64(key string, value int64) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

func (this *Session) WriteUint(key string, value uint) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

func (this *Session) WriteUint32(key string, value uint32) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

func (this *Session) WriteUint64(key string, value uint64) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

func (this *Session) Delete() bool {
	return this.Manager.(SessionWrapper).Delete(this.Sid)
}
