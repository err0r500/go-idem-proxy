package cache_inmem

import (
	"errors"
	"sync"

	"github.com/err0r500/go-idem-proxy/cache"
)

type Inmemcache struct {
	store *sync.Map
}

func New() cache.Cacher {
	return &Inmemcache{store: &sync.Map{}}
}

func (c *Inmemcache) Cache(key string, content cache.Response) error {
	c.store.Store(key, content)
	return nil
}

func (c Inmemcache) GetCache(key string) (*cache.Response, error) {
	if value, ok := c.store.Load(key); ok {
		r, ok := value.(cache.Response)
		if !ok {
			return nil, errors.New("cached is not a response")
		}
		return &r, nil
	}
	return nil, nil
}
