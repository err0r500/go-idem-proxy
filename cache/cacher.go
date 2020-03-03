package cache

type Cacher interface {
	Cache(key string, content Response) error
	GetCache(key string) (*Response, error)
}

type Response struct {
	Body       []byte
	StatusCode int
}
