package db

import (
	"database/sql"
	"fmt"
	pq "github.com/lib/pq"
	"github.com/sjbodzo/token-parser/pkg/coin"
)

type postgres struct {
	db *sql.DB
}

type CoinBackend interface {
	AddCoin(c *coin.Coin) error
	DeleteCoin(c *coin.Coin) error
	GetCoin(id string) (*coin.Coin, error)
}

// New returns a db instance that satisfies CoinBackend.
// This is used to add, delete, or get coins in the db.
func New(host, user, password, dbname string, port int) (CoinBackend, error) {
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgres{db: db}, nil
}

func (p *postgres) Close() {
	p.db.Close()
}

// AddCoin returns the coin by id
func (p *postgres) AddCoin(c *coin.Coin) error {
	statement := `
    	INSERT INTO coins (id, exchanges, taskrun)
		VALUES ($1, $2, $3);`
	if _, err := p.db.Exec(statement, c.ID, pq.Array(c.Markets), c.TaskRun); err != nil {
		return err
	}
	return nil
}

// DeleteCoin deletes the coin by id
func (p *postgres) DeleteCoin(c *coin.Coin) error {
	statement := `DELETE FROM coins WHERE id = $1;`
	res, err := p.db.Exec(statement, c.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	} else if count == 0 {
		return fmt.Errorf("no coin exists with ID: %s", c.ID)
	}

	return nil
}

// GetCoin gets the coin by id
func (p *postgres) GetCoin(id string) (*coin.Coin, error) {
	var taskrun int
	var exchanges []string
	statement := `SELECT exchanges, taskrun FROM coins WHERE id=$1;`
	row := p.db.QueryRow(statement, id)
	if err := row.Scan(&exchanges, &taskrun); err != nil {
		return nil, err
	}
	return &coin.Coin{
		ID:      id,
		Markets: exchanges,
		TaskRun: taskrun,
	}, nil
}
