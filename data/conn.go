package data

import (
	"github.com/go-redis/redis"
	"log"
	"time"
)

var (
	RedisConn         *redis.Client
	SessionExpireTime = 7 * 86400 // session过期时间(秒)
	host              = "127.0.0.1:6379"
	password          = "golyu"
	poolSize          = 100
	poolTimeout       = 30 * time.Second
)

// InitRedis redis初始化
func InitRedis() error {
	RedisConn = redis.NewClient(&redis.Options{
		Addr:        host,
		Password:    password,
		PoolSize:    poolSize,
		PoolTimeout: poolTimeout,
		DB:          0,
	})
	_, err := RedisConn.Ping().Result()
	if err != nil {
		log.Printf("err:%s\n", err.Error())
		return err
	}
	log.Printf("redis连接成功 %s\n", host)
	return nil
}
