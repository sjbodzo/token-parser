package cmd

import (
	"github.com/sjbodzo/token-parser/pkg/coin"
	"github.com/sjbodzo/token-parser/pkg/db"
	"github.com/sjbodzo/token-parser/pkg/server"
	"github.com/spf13/cobra"
	"log"
	"runtime"
	"time"
)
import "github.com/spf13/viper"

var (
	apiPort    int
	apiRate    int
	apiVersion string
	dbHost     string
	dbName     string
	dbPort     int
	dbUser     string
	dbPassword string
)

func init() {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
	})

	serverCmd.PersistentFlags().IntVar(&apiPort, "apiPort", 8080, "API port to listen on")
	serverCmd.PersistentFlags().IntVar(&dbPort, "dbPort", 5432, "DB port to listen on")
	serverCmd.PersistentFlags().IntVar(&apiRate, "apiRate", 50, "Rate limit in req / minute")
	serverCmd.PersistentFlags().StringVar(&dbHost, "dbHost", "coinparserpg-postgresql.coinapps.svc.cluster.local", "DB host to connect to")
	serverCmd.PersistentFlags().StringVar(&dbName, "dbName", "coins", "DB table to connect to")
	serverCmd.PersistentFlags().StringVar(&dbPassword, "dbPassword", "", "DB password")
	serverCmd.PersistentFlags().StringVar(&dbUser, "dbUser", "postgres", "DB user to connect with")
	serverCmd.PersistentFlags().StringVar(&apiVersion, "apiVersion", "v1", "API version")
	viper.BindPFlag("apiPort", serverCmd.PersistentFlags().Lookup("apiPort"))
	viper.BindPFlag("apiVersion", serverCmd.PersistentFlags().Lookup("apiVersion"))
	viper.BindPFlag("dbHost", serverCmd.PersistentFlags().Lookup("dbHost"))
	viper.BindPFlag("dbPort", serverCmd.PersistentFlags().Lookup("dbPort"))

	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the coin token-parser server",
	Long:  "Local init of the server token-parser for the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		// create a channel to receive all inbound requests
		var maxCapacity = 500
		coinChan := make(chan coin.Coin, maxCapacity)

		// use a worker per core
		var numWorkers = runtime.NumCPU()

		// limit our workers (which call the API) to "apiRate" req/min
		coinBuffer := throttle(apiRate, coinChan)

		// create a verifier that checks the coin again CoinGecko
		verifier := coin.NewCoinGeckoVerifier()

		// create a token-parser to persist coins to
		db, err := db.New(dbHost, dbUser, viper.GetString("DB_PASSWORD"), dbName, dbPort)
		if err != nil {
			log.Println("Could not connect to db")
			return err
		}

		// init the worker pool to process requests
		initWorkers(numWorkers, coinBuffer, verifier, db)

		// init the server to begin getting requests
		srv := server.New(coinChan, apiPort, apiVersion)
		if err := srv.Start(); err != nil {
			log.Println("Server exiting")
			return err
		}

		return nil
	},
}

// initWorkers initializes a pool of workers to process incoming requests.
// pass in the number of workers you want, the input buffer, your verifier
// to verify tokens, and the db layer (if any)
func initWorkers(num int, buffer <-chan coin.Coin, v coin.Verify, db db.CoinBackend) {
	for i := 1; i <= num; i++ {
		go func() {
			for {
				c := <-buffer
				coinMetadata, err := v.Verify(&c)
				if err != nil {
					log.Println("An error occurred with ", c.ID)
					log.Println(err)
					continue
				}

				// decorate our coin object with the market data
				if coinMetadata.Tickers != nil {
					for _, item := range *coinMetadata.Tickers {
						market := item.Market.Identifier
						c.Markets = append(c.Markets, market)
					}
				}

				if db != nil {
					err := db.AddCoin(&c)
					if err != nil {
						log.Println("An error occurred during db add: ")
						log.Println(err)
					}
				}
			}
		}()
	}
}

// throttle acts as a buffered pipe, throttling requests from the source channel.
// rate is how many requests per second to buffer for, and is useful
// for limiting API traffic due to requests / sec limits.
// source is the channel receiving all requests that we want to throttle.
func throttle(rate int, source chan coin.Coin) <-chan coin.Coin {
	bufferChan := make(chan coin.Coin, 50)
	duration := time.Minute / time.Duration(rate)
	go func() {
		throttle := time.Tick(duration)
		for ; true; <-throttle {
			c := <-source
			bufferChan <- c
		}
	}()
	return bufferChan
}
