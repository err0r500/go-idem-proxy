package cache

type Cacher interface {
	Cache(key string, content string) error
	GetCache(key string) (*string, error)
}
