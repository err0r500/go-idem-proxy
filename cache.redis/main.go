package cache

import (
	"fmt"
	"strconv"

	"github.com/err0r500/go-idem-proxy/types"
	"github.com/gomodule/redigo/redis"
)

type redisCache struct {
	conn redis.Conn
	ttl  string
}

func New(c redis.Conn, ttlInSec int) types.Cacher {
	return &redisCache{conn: c, ttl: strconv.Itoa(ttlInSec)}
}

func (c redisCache) Cache(key string, content string) error {
	c.conn.Do("SET", key, content)
	c.conn.Do("EXPIRE", key, c.ttl)
	return nil
}

func (c redisCache) GetCache(key string) (*string, error) {
	result, err := c.conn.Do("GET", key)
	str := fmt.Sprintf("%s", result)
	return &str, err
}
