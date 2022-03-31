package coin

import (
	"fmt"
	coingecko "github.com/superoo7/go-gecko/v3"
	"github.com/superoo7/go-gecko/v3/types"
	"net/http"
	"sync"
	"time"
)

type Coin struct {
	ID      string
	Markets []string
	TaskRun int
}

type Coins struct {
	List []string `json:"coins"`
}

// Verify performs verification for a given Coin{}, returning
// relevant data for valid coins.
type Verify interface {
	Verify(c *Coin) (*types.CoinsID, error)
}

// coinSet is a convenience struct to implement a simple HashSet
type coinSet map[string]struct{}

func (set coinSet) add(coin *Coin) {
	set[coin.ID] = struct{}{}
}

func (set coinSet) del(coin *Coin) {
	delete(set, coin.ID)
}

func (set coinSet) has(coin *Coin) bool {
	_, ok := set[coin.ID]
	return ok
}

// GeckoClient satisfies Verify using the CoinGecko API
type GeckoClient struct {
	client *coingecko.Client
	known  coinSet
	mu     sync.Mutex
}

func (gecko *GeckoClient) addCoin(c *Coin) {
	defer gecko.mu.Unlock()
	gecko.mu.Lock()
	gecko.known.add(c)
}

func (gecko *GeckoClient) deleteCoin(c *Coin) {
	defer gecko.mu.Unlock()
	gecko.mu.Lock()
	gecko.known.del(c)
}

func (gecko *GeckoClient) hasCoin(c *Coin) bool {
	return gecko.known.has(c)
}

// Verify the coin is valid by checking if we have already seen it,
// and if it has a valid ID. If it is valid, add it to our list of
// known coins and return metadata about it.
func (gecko *GeckoClient) Verify(c *Coin) (*types.CoinsID, error) {
	if gecko.hasCoin(c) {
		err := fmt.Errorf("coin with id %s already known", c.ID)
		return nil, err
	}

	metadata, err := gecko.client.CoinsID(c.ID, false, true, false, false, false, false)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// NewCoinGeckoVerifier returns a GeckoVerifier for use verifying coins by ID
// against the CoinGecko API.
func NewCoinGeckoVerifier() *GeckoClient {
	c := http.Client{
		Timeout: time.Second * 30,
	}
	cg := coingecko.NewClient(&c)

	return &GeckoClient{
		client: cg,
	}
}
