package redis

import (
	"github.com/go-redis/redis"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/files"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	"strconv"
	"sync"
	"time"
)

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type RedisClient struct {
	client *redis.Client
}

var redisClient *RedisClient
var mu sync.Mutex

// 获取连接的客户端
func Client() (*RedisClient, error) {
	mu.Lock()
	defer mu.Unlock()

	if redisClient != nil {
		return redisClient, nil
	}

	configFile := files.NewFile(Tea.ConfigFile("redis.conf"))
	var config = &RedisConfig{}
	reader, err := configFile.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	err = reader.ReadYAML(config)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Host + ":" + strconv.Itoa(config.Port),
		Password: config.Password,
	})
	redisClient = &RedisClient{
		client: client,
	}
	return redisClient, nil
}

func (this *RedisClient) GetInt(key string, defaultValue int) int {
	result, err := this.client.Get(key).Result()
	if err != nil {
		return defaultValue
	}
	if len(result) == 0 {
		return defaultValue
	}
	return types.Int(result)
}

func (this *RedisClient) GetString(key string) string {
	result, err := this.client.Get(key).Result()
	if err != nil {
		if err.Error() != "redis: nil" {
			logs.Error(err)
		}
		return ""
	}
	return result
}

func (this *RedisClient) HGetAll(key string) (map[string]string, error) {
	return this.client.HGetAll(key).Result()
}

func (this *RedisClient) HSet(key string, field string, value interface{}) (bool, error) {
	return this.client.HSet(key, field, value).Result()
}

func (this *RedisClient) HMSet(key string, fields map[string]interface{}) (string, error) {
	return this.client.HMSet(key, fields).Result()
}

func (this *RedisClient) ExpireAt(key string, expireTime time.Time) (bool, error) {
	return this.client.ExpireAt(key, expireTime).Result()
}

func (this *RedisClient) Del(key ...string) (int64, error) {
	return this.client.Del(key...).Result()
}

func (this *RedisClient) Set(key string, value interface{}, expiration time.Duration) (string, error) {
	return this.client.Set(key, value, expiration).Result()
}
