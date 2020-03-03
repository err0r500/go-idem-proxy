package cache_redis

import (
	"encoding/json"

	"github.com/err0r500/go-idem-proxy/cache"
	"github.com/gomodule/redigo/redis"
)

type redisCache struct {
	conn redis.Conn
	ttl  int
}

func New(c redis.Conn, ttlInSec int) cache.Cacher {
	return &redisCache{conn: c, ttl: ttlInSec}
}

func (c redisCache) Cache(key string, content cache.Response) error {
	toCache, err := json.Marshal(content)
	if err != nil {
		return err
	}

	c.conn.Do("SET", key, toCache)
	c.conn.Do("EXPIRE", key, c.ttl)
	return nil
}

func (c redisCache) GetCache(key string) (*cache.Response, error) {
	result, err := c.conn.Do("GET", key)
	if err != nil {
		return nil, err
	}

	bResult, ok := result.([]byte)
	if !ok {
		return nil, err
	}

	r := &cache.Response{}
	if err := json.Unmarshal(bResult, r); err != nil {
		return nil, err
	}

	return r, nil
}
