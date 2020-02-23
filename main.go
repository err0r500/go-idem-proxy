package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Cacher interface {
	Cache(key string, content string) error
	GetCache(key string) (*string, error)
}

type Inmemcache struct {
	store *sync.Map
}

func NewInMemCache() Cacher {
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

func main() {
	targetURL := "http://localhost:3000"
	url, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal("couldn't start due to malformed URL", targetURL)
	}

	http.Handle("/", GetHandler(NewInMemCache(), url))
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func GetHandler(cacher Cacher, url *url.URL) http.Handler {
	p := httputil.NewSingleHostReverseProxy(url)

	return http.HandlerFunc(func(origRW http.ResponseWriter, origReq *http.Request) {
		switch origReq.Method {
		case http.MethodPost:
			idemToken := origReq.Header.Get("X-idem-token")
			if idemToken == "" {
				origRW.WriteHeader(http.StatusBadRequest)
				return
			}

			cachedResp, err := cacher.GetCache(idemToken)
			if err != nil {
				log.Println("failed to get Cache", err.Error())
				origRW.WriteHeader(http.StatusInternalServerError)
				return
			}
			if cachedResp != nil {
				origRW.Write([]byte(*cachedResp))
				return
			}
			p.ModifyResponse = func(rf *http.Response) error {
				rBody, err := ioutil.ReadAll(rf.Body)
				if err != nil {
					log.Println("failed to read response body", err.Error())
					return nil
				}
				cacher.Cache(idemToken, string(rBody))
				origRW.Write(rBody)
				return nil
			}
			p.ServeHTTP(origRW, origReq)

		default:
			p.ServeHTTP(origRW, origReq)
		}
	})
}
