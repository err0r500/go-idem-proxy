// +build integration

package cache_redis_test

import (
	"log"
	"testing"

	"github.com/err0r500/go-idem-proxy/cache"
	cache_redis "github.com/err0r500/go-idem-proxy/cache/redis"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestHappy(t *testing.T) {
	address := "redis:6379"
	c, err := redis.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	insertedValue := cache.Response{
		Body:       []byte("salut"),
		StatusCode: 300,
	}
	key := "bla"
	cacher := cache_redis.New(c, 60)

	cacher.Cache(key, insertedValue)
	result, err := cacher.GetCache(key)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	//assert.Equal(t, insertedValue, *result)
}
