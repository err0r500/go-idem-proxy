// +build integration

package cache_inmem_test

import (
	"log"
	"testing"

	"github.com/err0r500/go-idem-proxy/cache.inmem"
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

	insertedValue := "salut"
	key := "bla"
	cacher := cache.New(c)

	cacher.Cache(key, insertedValue)
	result, err := cacher.GetCache(key)
	assert.NotNil(t, result)
	assert.Equal(t, insertedValue, *result)
}
