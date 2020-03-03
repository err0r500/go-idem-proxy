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

func (h handlerv1) handlePost(p *httputil.ReverseProxy, origRW http.ResponseWriter, origReq *http.Request) {
	idemToken := origReq.Header.Get(h.idemToken)
	if idemToken == "" {
		origRW.WriteHeader(http.StatusBadRequest)
		return
	}

	cachedResp, err := h.cacher.GetCache(idemToken)
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
			if err := h.cacher.Cache(idemToken, string(rBody)); err != nil {
				log.Println("failed to put in cache", err.Error())
			}
		}(idemToken, rBody)

		// restore original readCloser
		rf.Body = ioutil.NopCloser(bytes.NewBuffer(rBody))
		return nil
	}
	p.ServeHTTP(origRW, origReq)
}
