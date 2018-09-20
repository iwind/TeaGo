package actions

import (
	"github.com/iwind/TeaGo/types"
	"fmt"
)

// SESSION通用配置
type SessionConfig struct {
	Life   uint
	Secret string
}

// SESSION管理器接口
type SessionWrapper interface {
	Init(config *SessionConfig)
	Read(sid string) map[string]string
	WriteItem(sid string, key string, value string) bool
	Delete(sid string) bool
}

// Session定义
type Session struct {
	Sid     string
	Manager interface{}
}

// 设置sid
func (this *Session) SetSid(sid string) {
	this.Sid = sid
}

// 取得所有存储在session中的值
func (this *Session) Values() map[string]string {
	return this.Manager.(SessionWrapper).Read(this.Sid)
}

// 获取字符串值
func (this *Session) GetString(key string) string {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return ""
	}
	return value
}

// 获取int值
func (this *Session) GetInt(key string) int {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Int(value)
}

// 获取int32值
func (this *Session) GetInt32(key string) int32 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Int32(value)
}

// 获取int64值
func (this *Session) GetInt64(key string) int64 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Int64(value)
}

// 获取uint值
func (this *Session) GetUint(key string) uint {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Uint(value)
}

// 获取uint32值
func (this *Session) GetUint32(key string) uint32 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Uint32(value)
}

// 获取uint64值
func (this *Session) GetUint64(key string) uint64 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Uint64(value)
}

// 获取float32值
func (this *Session) GetFloat32(key string) float32 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Float32(value)
}

// 获取64值
func (this *Session) GetFloat64(key string) float64 {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return 0
	}
	return types.Float64(value)
}

// 获取bool值
func (this *Session) GetBool(key string) bool {
	values := this.Values()
	value, ok := values[key]
	if !ok {
		return false
	}
	return types.Bool(value)
}

// 写入值
func (this *Session) Write(key, value string) bool {
	return this.Manager.(SessionWrapper).WriteItem(this.Sid, key, value)
}

// 写入int值
func (this *Session) WriteInt(key string, value int) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

// 写入int32值
func (this *Session) WriteInt32(key string, value int32) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

// 写入int64值
func (this *Session) WriteInt64(key string, value int64) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

// 写入uint值
func (this *Session) WriteUint(key string, value uint) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

// 写入uint32值
func (this *Session) WriteUint32(key string, value uint32) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

// 写入uint64值
func (this *Session) WriteUint64(key string, value uint64) bool {
	return this.Write(key, fmt.Sprintf("%d", value))
}

// 删除整个session
func (this *Session) Delete() bool {
	return this.Manager.(SessionWrapper).Delete(this.Sid)
}
