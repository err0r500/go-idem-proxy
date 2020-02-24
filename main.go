package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/err0r500/go-idem-proxy/cache.redis"
	"github.com/err0r500/go-idem-proxy/proxyHandler"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	viper.SetDefault("Idem_Token", "X-idem-token")
	idemToken := viper.GetString("Idem_Token")

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

	h := proxyHandler.New(cache.New(c, cacheTTL), idemToken)
	url, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal("couldn't start due to malformed URL", targetURL)
	}
	http.Handle("/", h.Handle(url))
	if err = http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
