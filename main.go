package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/err0r500/go-idem-proxy/cache.inmem"
	"github.com/err0r500/go-idem-proxy/types"
)

func main() {
	targetURL := "http://localhost:3000"
	url, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal("couldn't start due to malformed URL", targetURL)
	}

	http.Handle("/", GetHandler(cache.New(), url))
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

		// restore original readCloser
		rf.Body = ioutil.NopCloser(bytes.NewBuffer(rBody))
		cacher.Cache(idemToken, string(rBody))
		return nil
	}
	p.ServeHTTP(origRW, origReq)

}
