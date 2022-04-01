package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/sjbodzo/token-parser/pkg/coin"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	port     int
	srv      http.Server
	coins    chan coin.Coin
	taskRuns int
}

// New creates a http.Server for processing REST requests.
// A channel is expected as input to push coins onto as
// they are parsed in inbound requests.
func New(c chan coin.Coin, port int, version string) *Server {
	route := fmt.Sprintf("/api/%s/parse", version)
	mux := http.NewServeMux()

	server := Server{
		port: port,
		srv: http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      mux,
			WriteTimeout: 1 * time.Second,
			ReadTimeout:  2 * time.Second,
		},
		coins: c,
	}
	mux.HandleFunc(route, server.parse())

	return &server
}

// Start handles starting the underlying http.Server
func (s *Server) Start() error {
	return http.ListenAndServe(s.srv.Addr, s.srv.Handler)
}

// parse does all the heavy lifting here.
// note: verification and processing occurs asynchronously, as this collector
// must respond quickly to meet an SLA of 400 ms
func (s *Server) parse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.taskRuns = s.taskRuns + 1 // increment counter to track batch number
		defer r.Body.Close()

		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, err := ioutil.ReadAll(r.Body)
			// log.Println(string(body))
			if err != nil {
				http.Error(w, "Unable to recv body", http.StatusInternalServerError)
			}

			var coins coin.Coins
			err = json.Unmarshal(body, &coins)
			if err != nil {
				log.Println(err)
				http.Error(w,
					"Unable to parse input as json",
					http.StatusInternalServerError)
			}
			for _, c := range coins.List {
				s.coins <- coin.Coin{ID: c, TaskRun: s.taskRuns}
			}

		case "text/csv":
			r := csv.NewReader(r.Body)
			records, err := r.ReadAll()
			if err != nil {
				http.Error(w,
					"Unable to parse input as csv",
					http.StatusInternalServerError)
			}

			if records == nil || len(records) < 2 {
				return
			} else if strings.Join(records[0], "") != "coins" {
				http.Error(w,
					"Unexpected body in request; Expecting 'coins' csv header",
					http.StatusBadRequest)
			} else {
				for _, c := range records[0][1:] {
					s.coins <- coin.Coin{ID: c, TaskRun: s.taskRuns}
				}
			}

		default:
			http.Error(w,
				"Unable to parse request format. Set your 'Content-Type' header to 'application/json' or 'text/csv'",
				http.StatusBadRequest)
		}

	}
}
