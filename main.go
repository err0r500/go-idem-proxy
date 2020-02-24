package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/err0r500/go-idem-proxy/cache.redis"
	"github.com/err0r500/go-idem-proxy/types"
	"github.com/gomodule/redigo/redis"
)

func main() {
	address := "redis:6379"
	c, err := redis.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	targetURL := "http://localhost:3000"
	url, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal("couldn't start due to malformed URL", targetURL)
	}
	http.Handle("/", GetHandler(cache.New(c), url))
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func GetHandler(cacher types.Cacher, url *url.URL) http.Handler {
	p := httputil.NewSingleHostReverseProxy(url)

	return http.HandlerFunc(func(origRW http.ResponseWriter, origReq *http.Request) {
		switch origReq.Method {
		case http.MethodPost:
			handlePost(p, cacher, origRW, origReq)
		default:
			p.ServeHTTP(origRW, origReq)
		}
	})
}

func handlePost(p *httputil.ReverseProxy, cacher types.Cacher, origRW http.ResponseWriter, origReq *http.Request) {
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

		// if it fails, bad luck, we don't really care
		go func(idemToken string, rBody []byte) {
			if err := cacher.Cache(idemToken, string(rBody)); err != nil {
				log.Println("failed to put in cache", err.Error())
			}
		}(idemToken, rBody)

		// restore original readCloser
		rf.Body = ioutil.NopCloser(bytes.NewBuffer(rBody))
		return nil
	}
	p.ServeHTTP(origRW, origReq)

}
