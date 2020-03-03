package handler_v1

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/err0r500/go-idem-proxy/cache"
	"github.com/err0r500/go-idem-proxy/handler"
)

type handlerv1 struct {
	idemToken string
	cacher    cache.Cacher
}

func New(cacher cache.Cacher, idemToken string) handler.Handler {
	return handlerv1{
		idemToken,
		cacher,
	}
}

func (h handlerv1) Handle(url *url.URL) http.Handler {
	p := httputil.NewSingleHostReverseProxy(url)

	return http.HandlerFunc(func(origRW http.ResponseWriter, origReq *http.Request) {
		switch origReq.Method {
		case http.MethodPost:
			h.handlePost(p, origRW, origReq)
		default:
			p.ServeHTTP(origRW, origReq)
		}
	})
}

func (h handlerv1) handlePost(p *httputil.ReverseProxy, proxyRW http.ResponseWriter, proxyReq *http.Request) {
	idemToken := proxyReq.Header.Get(h.idemToken)
	if idemToken == "" {
		proxyRW.WriteHeader(http.StatusBadRequest)
		return
	}

	cachedResp, err := h.cacher.GetCache(idemToken)
	if err != nil {
		log.Println("failed to get Cache", err.Error())
		proxyRW.WriteHeader(http.StatusInternalServerError)
		return
	}
	// the response has been found in cache, we respond this immediatly
	if cachedResp != nil {
		proxyRW.WriteHeader(cachedResp.StatusCode)
		proxyRW.Write(cachedResp.Body)
		return
	}

	// nothing found in cache, we wait for the target response to come back and insert it in cache
	p.ModifyResponse = func(targetResp *http.Response) error {
		rBody, err := ioutil.ReadAll(targetResp.Body)
		if err != nil {
			log.Println("failed to read response body", err.Error())
			return nil
		}

		go func() {
			if err := h.cacher.Cache(
				idemToken,
				cache.Response{
					Body:       rBody,
					StatusCode: targetResp.StatusCode,
				},
			); err != nil {
				// if it fails, bad luck, we just log
				log.Println("failed to put in cache", err.Error())
			}
		}()

		// restore original readCloser
		targetResp.Body = ioutil.NopCloser(bytes.NewBuffer(rBody))
		return nil
	}
	p.ServeHTTP(proxyRW, proxyReq)
}
