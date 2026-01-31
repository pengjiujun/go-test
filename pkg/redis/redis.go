package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"test/pkg/config"
)

var RedisClient *redis.Client

func InitRedis() {

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Conf.Redis.Addr, strconv.Itoa(config.Conf.Redis.Port)),
		Password: config.Conf.Redis.Password, // no password set
		DB:       config.Conf.Redis.DB,       // use default DB
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	RedisClient = rdb
}
