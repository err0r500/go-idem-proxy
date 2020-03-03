package handler

import (
	"net/http"
	"net/url"
)

type Handler interface {
	Handle(url *url.URL) http.Handler
}
