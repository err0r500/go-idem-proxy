package main_test

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/err0r500/go-idem-proxy/cache.inmem"
	"github.com/err0r500/go-idem-proxy/proxyHandler"
	"gopkg.in/h2non/baloo.v3"
)

var initial = "initial"
var pathChars = "abcdefghijklmnopqrstuvwxyz/-_"

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = pathChars[rand.Intn(len(pathChars))]
	}
	return "/" + string(bytes)
}

func randomPath() string {
	return randomString(30)
}

func TestPostRequestNeedsHeader(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	target, targetURL := startTarget(initial)
	defer target.Close()

	pHandler := proxyHandler.New(nil, "bal")
	proxy := httptest.NewServer(pHandler.Handle(targetURL))
	defer proxy.Close()

	baloo.New(proxy.URL).Post(randomPath()).Expect(t).Status(http.StatusBadRequest).Done()
}

func TestPostRequestsPostUsesCache(t *testing.T) {
	idemToken := "X-idem-token"
	target, targetURL := startTarget(initial)
	defer target.Close()

	pHandler := proxyHandler.New(cache.New(), idemToken)
	proxy := httptest.NewServer(pHandler.Handle(targetURL))
	defer proxy.Close()

	path := randomPath()
	req := baloo.New(proxy.URL).SetHeader(idemToken, "bla")
	req.Post(path).Expect(t).Status(200).BodyEquals(initial).Done()
	req.Post(path).Expect(t).Status(200).BodyEquals(initial).Done()
}

func startTarget(initial string) (*httptest.Server, *url.URL) {
	hits := 1
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, initial)
		hits++
		initial += fmt.Sprintf("%d", hits)
	}))

	targetURL, err := url.Parse(target.URL)
	if err != nil {
		log.Fatal(err)
	}
	return target, targetURL
}
