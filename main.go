package main

import (
	coin "github.com/sjbodzo/token-parser/pkg/coin"
	"github.com/sjbodzo/token-parser/pkg/db"
	server "github.com/sjbodzo/token-parser/pkg/server"
	"log"
	"runtime"
	"time"
)

func main() {
	// create a channel to receive all inbound requests
	var maxCapacity = 500
	var apiLimit = 50 // reqs/min
	coinChan := make(chan coin.Coin, maxCapacity)

	// use a worker per core
	var numWorkers = runtime.NumCPU()

	// limit our workers (which call the API) to 50 req/min
	coinBuffer := throttle(apiLimit, coinChan)

	// create a verifier that checks the coin again CoinGecko
	verifier := coin.NewCoinGeckoVerifier()

	// create a backend to persist coins to
	var host = "localhost"
	var port = 9000
	var user = "postgres"
	var password = "postgres"
	var dbname = "coins"
	db, err := db.New(host, user, password, dbname, port)
	if err != nil {
		log.Println("Could not connect to db")
		panic(err)
	}

	// init the worker pool to process requests
	initWorkers(numWorkers, coinBuffer, verifier, db)

	// init the server to begin getting requests
	srv := server.New(coinChan)
	if err := srv.Start(); err != nil {
		log.Println("Server exiting")
		panic(err)
	}
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
	log.Println(duration)
	go func() {
		throttle := time.Tick(duration)
		for ; true; <-throttle {
			c := <-source
			bufferChan <- c
		}
	}()
	return bufferChan
}
