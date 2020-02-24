package cache

import (
	"fmt"

	"github.com/err0r500/go-idem-proxy/types"
	"github.com/gomodule/redigo/redis"
)

type redisCache struct {
	conn redis.Conn
}

func New(c redis.Conn) types.Cacher {
	return &redisCache{conn: c}
}

func (c redisCache) Cache(key string, content string) error {
	c.conn.Do("SET", key, content)
	return nil
}

func (c redisCache) GetCache(key string) (*string, error) {
	result, err := c.conn.Do("GET", key)
	str := fmt.Sprintf("%s", result)
	return &str, err
}
