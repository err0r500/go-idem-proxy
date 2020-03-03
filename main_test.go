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

	cache "github.com/err0r500/go-idem-proxy/cache/inmem"
	proxyHandler "github.com/err0r500/go-idem-proxy/handler/v1"
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

	target, targetURL := startTarget(initial, http.StatusOK)
	defer target.Close()

	pHandler := proxyHandler.New(nil, "bal")
	proxy := httptest.NewServer(pHandler.Handle(targetURL))
	defer proxy.Close()

	baloo.New(proxy.URL).Post(randomPath()).Expect(t).Status(http.StatusBadRequest).Done()
}

func TestPostRequestsUsesCache(t *testing.T) {
	idemToken := "X-idem-token"
	status := http.StatusTeapot

	target, targetURL := startTarget(initial, status)
	defer target.Close()

	pHandler := proxyHandler.New(cache.New(), idemToken)
	proxy := httptest.NewServer(pHandler.Handle(targetURL))
	defer proxy.Close()

	path := randomPath()
	req := baloo.New(proxy.URL).SetHeader(idemToken, "bla")
	req.Post(path).Expect(t).Status(status).BodyEquals(initial).Done()
	req.Post(path).Expect(t).Status(status).BodyEquals(initial).Done()
}

func startTarget(initial string, status int) (*httptest.Server, *url.URL) {
	// hits will be incremented on each call, this allows us to check if the
	// target server actually received the request only once
	hits := 1
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
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
