package service

import (
	"database/sql"
	"log"
)

type CounterService interface {
	Increment()
	GetCounter() int
}

type counterService struct{}

type Callback func(db *sql.DB) (any, error)

func NewCounterService() CounterService {
	return &counterService{}
}

func (c *counterService) Increment() {
	db, err := sql.Open("sqlite3", "./counter.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = c.incrementCounter(db)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *counterService) GetCounter() (count int) {
	db, err := sql.Open("sqlite3", "./counter.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	count, err = counter(db)
	if err != nil {
		log.Fatal(err)
	}
	return count
}

func counter(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT count FROM counter").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *counterService) incrementCounter(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	var count int
	err = tx.QueryRow("SELECT count FROM counter").Scan(&count)
	if err != nil {
		tx.Rollback()
		return err
	}

	count++
	_, err = tx.Exec("UPDATE counter SET count = ?", count)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *counterService) dbExecute(db *sql.DB, callback Callback) any {
	db, err := sql.Open("sqlite3", "./counter.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	result, err := callback(db)
	if err != nil {
		log.Fatal(err)
	}
	return result
}
