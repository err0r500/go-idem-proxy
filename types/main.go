package types

import (
	"net/http"
	"net/url"
)

type Cacher interface {
	Cache(key string, content string) error
	GetCache(key string) (*string, error)
}

type Handler interface {
	Handle(url *url.URL) http.Handler
}
