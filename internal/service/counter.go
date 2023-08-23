package service

import (
	"database/sql"
	"fmt"
	"log"
)

type CounterService interface {
	Increment(id int)
	GetCounter() [3]int
}

type counterService struct{}

type Callback func(db *sql.DB) (any, error)

func NewCounterService() CounterService {
	return &counterService{}
}

func (c *counterService) Increment(id int) {
	err := incrementRate(id)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *counterService) GetCounter() [3]int {
	db, err := sql.Open("sqlite3", "./counter.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	count, err := getAllRates()
	if err != nil {
		log.Fatal(err)
	}
	return count
}

func getAllRates() ([3]int, error) {
	db, err := sql.Open("sqlite3", "your_database_file_path.db")
	if err != nil {
		return [3]int{}, err
	}
	defer db.Close()

	var rates [3]int
	query := "SELECT one, two, three FROM rates LIMIT 1"
	err = db.QueryRow(query).Scan(&rates[0], &rates[1], &rates[2])
	if err != nil {
		return [3]int{}, err
	}

	return rates, nil
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

//coding standards ignore start

func incrementRate(rateNumber int) error {
	db, err := sql.Open("sqlite3", "./counter.db")
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE rates SET %s = %s + 1", getRateColumnName(rateNumber), getRateColumnName(rateNumber))
	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func getRateColumnName(rateNumber int) string {
	switch rateNumber {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	default:
		return ""
	}
}
