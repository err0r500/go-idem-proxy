package cache_inmem

import (
	"fmt"
	"sync"

	"github.com/err0r500/go-idem-proxy/cache"
)

type Inmemcache struct {
	store *sync.Map
}

func New() cache.Cacher {
	return &Inmemcache{store: &sync.Map{}}
}

func (c *Inmemcache) Cache(key string, content string) error {
	c.store.Store(key, content)
	return nil
}

func (c Inmemcache) GetCache(key string) (*string, error) {
	if value, ok := c.store.Load(key); ok {
		valStr := fmt.Sprintf("%v", value)
		return &valStr, nil
	}
	return nil, nil
}