package repository

import (
	"database/sql"
	"log"
)

type CountRepository interface {
	GetCount() int
	IncrementCount()
}

type countRepository struct {
	count int
}

func (c *countRepository) GetCount() int {
	db, err := sql.Open("sqlite3", "./counter.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	count, err := c.getCountNum(db)
	if err != nil {
		log.Fatal(err)
	}
	return count
}

func (c *countRepository) getCountNum(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT count FROM counter").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
