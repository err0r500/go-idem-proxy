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
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8080")
	port := viper.GetString("PORT")

	address := viper.GetString("Redis_Conn")
	if address == "" {
		log.Fatal("need Redis Connection string")
	}

	targetURL := viper.GetString("Target_URL")
	if targetURL == "" {
		log.Fatal("need proxy target string")
	}

	viper.SetDefault("Cache_TTL", 60)
	cacheTTL := viper.GetInt("Cache_TTL")

	c, err := redis.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	url, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal("couldn't start due to malformed URL", targetURL)
	}
	http.Handle("/", GetHandler(cache.New(c, cacheTTL), url))
	if err = http.ListenAndServe(":"+port, nil); err != nil {
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
