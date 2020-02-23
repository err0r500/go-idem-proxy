package main_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	idemProxy "github.com/err0r500/go-idem-proxy"
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

	proxy := httptest.NewServer(idemProxy.GetHandler(nil, targetURL))
	defer proxy.Close()

	baloo.New(proxy.URL).Post(randomPath()).Expect(t).Status(http.StatusBadRequest).Done()
}

func TestPostRequestsPostUsesCache(t *testing.T) {
	target, targetURL := startTarget(initial)
	defer target.Close()

	proxy := httptest.NewServer(idemProxy.GetHandler(idemProxy.NewInMemCache(), targetURL))
	defer proxy.Close()

	path := randomPath()
	req := baloo.New(proxy.URL).SetHeader("X-idem-token", "bla")
	req.Post(path).Expect(t).Status(200).BodyEquals(initial).Done()
	req.Post(path).Expect(t).Status(200).BodyEquals(initial).Done()
}

func hitTwice(url string) string {
	resp, err := http.Post(url, "", nil)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(b)
	return string(b)
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
