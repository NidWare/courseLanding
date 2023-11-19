package service

import (
	"database/sql"
	"fmt"
	"log"
)

type RepositoryService interface {
	LoadOrders() map[string]string
	DeleteOrdersByIds(ids []string)
	IncrementClicks(rateID int) error
	UpdateLimit(rateID int, newLimit int) error
	GetClicks(rateID int) (int, error)
	GetLimit(rateID int) (int, error)
}

type repositoryService struct {
	db       *sql.DB
	dbOrders *sql.DB
}

type RateCounter struct {
	RateID int
	Clicks int
	Limit  int
	Price  string
}

func NewRepositoryService(db *sql.DB, dbOrders *sql.DB) RepositoryService {
	return &repositoryService{db: db, dbOrders: dbOrders}
}

func (r *repositoryService) DeleteOrdersByIds(ids []string) {
	for _, paymentID := range ids {
		if _, err := r.dbOrders.Exec("DELETE FROM orders WHERE payment_id = ?", paymentID); err != nil {
			log.Printf("Failed to delete payment %s from orders: %v", paymentID, err)
		}
	}
}

func (r *repositoryService) LoadOrders() map[string]string {
	rows, err := r.dbOrders.Query("SELECT payment_id, email FROM orders")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var paymentsByIds map[string]string

	paymentsByIds = make(map[string]string)

	for rows.Next() {
		var paymentID, email string
		if err := rows.Scan(&paymentID, &email); err != nil {
			fmt.Println("Got panic while loading orders")
			return map[string]string{}
		}
		paymentsByIds[paymentID] = email
	}
	return paymentsByIds
}

func (r *repositoryService) IncrementClicks(rateID int) error {
	_, err := r.db.Exec("UPDATE rateCounter SET clicks = clicks + 1 WHERE rate_id = ?", rateID)
	return err
}

func (r *repositoryService) UpdateLimit(rateID int, newLimit int) error {
	_, err := r.db.Exec("UPDATE rateCounter SET \"limit\" = ? WHERE rate_id = ?", newLimit, rateID)
	return err
}

func (r *repositoryService) GetClicks(rateID int) (int, error) {
	var clicks int
	err := r.db.QueryRow("SELECT clicks FROM rateCounter WHERE rate_id = ?", rateID).Scan(&clicks)
	return clicks, err
}

func (r *repositoryService) GetLimit(rateID int) (int, error) {
	var limit int
	err := r.db.QueryRow("SELECT \"limit\" FROM rateCounter WHERE rate_id = ?", rateID).Scan(&limit)
	return limit, err
}
