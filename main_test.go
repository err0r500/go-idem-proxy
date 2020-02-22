package main_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	idemProxy "github.com/err0r500/go-idem-proxy"
	"github.com/stretchr/testify/assert"
)

var initial = "initial"

func TestHappy(t *testing.T) {
	target, targetURL := startTarget(initial)
	defer target.Close()

	proxy := httptest.NewServer(idemProxy.GetHandler(targetURL))
	defer proxy.Close()

	resp := hitTwice(proxy.URL)
	assert.Equal(t, initial, resp)
}

func hitTwice(url string) string {
	if _, err := http.Post(url, "", nil); err != nil {
		log.Fatal(err)
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
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
