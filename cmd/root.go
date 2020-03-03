package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	cache "github.com/err0r500/go-idem-proxy/cache/redis"
	proxyHandler "github.com/err0r500/go-idem-proxy/handler/v1"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	port      int
	idemToken string
	targetURL string
	redisURL  string
	cacheTTL  int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-idem-proxy",
	Short: "proxy bringing idempotency to your POST requests",
	Long: `This proxy will look for an HTTP header with a token in order
to cache POST requests in a Redis.
`,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-idem-proxy.yaml)")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "the port the proxy listens on")
	rootCmd.PersistentFlags().StringVarP(&idemToken, "idem-token", "i", "X-idem-token", "the header that will have to be provided on POST requests")
	rootCmd.PersistentFlags().StringVarP(&targetURL, "target-url", "t", "localhost:3000", "where the proxy will forward the traffic")
	rootCmd.PersistentFlags().StringVarP(&redisURL, "redis-url", "r", "localhost:6379", "the URL of the redis database for caching")
	rootCmd.PersistentFlags().IntVarP(&cacheTTL, "cache-ttl", "c", 60, "how long the cached requests are persisted")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".go-idem-proxy")
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func run() {
	c, err := redis.Dial("tcp", redisURL)
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
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatal(err)
	}

	log.Println("server started")
}
