package service

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

type RepositoryService interface {
	LoadOrders() map[string]string
	DeleteOrdersByIds(ids []string)
	IncrementClicks(rateID int) error
	UpdateLimit(rateID int, newLimit int) error
	GetClicks(rateID int) (int, error)
	GetLimit(rateID int) (int, error)

	GetRateByID(rateID int) (Rate, error)
	GetRateByPrice(price string) (Rate, error)
	GetAllRates() ([]Rate, error)
}

type repositoryService struct {
	db       *sql.DB
	dbOrders *sql.DB
}

type Rate struct {
	RateID  int
	Clicks  int
	Limit   int
	Price   float64
	GroupID int
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
	_, err := r.db.Exec("UPDATE rate SET clicks = clicks + 1 WHERE rate_id = ?", rateID)
	return err
}

func (r *repositoryService) UpdateLimit(rateID int, newLimit int) error {
	_, err := r.db.Exec("UPDATE rate SET \"limit\" = ? WHERE rate_id = ?", newLimit, rateID)
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

func (r *repositoryService) GetRateByID(rateID int) (Rate, error) {
	var rate Rate

	query := "SELECT rate_id, clicks, \"limit\", price, group_id FROM rate WHERE rate_id = ?"
	row := r.db.QueryRow(query, rateID)

	err := row.Scan(&rate.RateID, &rate.Clicks, &rate.Limit, &rate.Price, &rate.GroupID)
	if err != nil {
		return rate, err
	}

	return rate, nil
}

func (r *repositoryService) GetRateByPrice(price string) (Rate, error) {
	var rate Rate

	// Parse the price into a float
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return rate, err
	}

	// Calculate the lower bound (10% lower than the provided price)
	lowerBound := priceFloat * 0.9

	// Adjust the query to use BETWEEN for the price range
	query := "SELECT rate_id, clicks, \"limit\", price, group_id FROM rate WHERE price BETWEEN ? AND ?"

	// Execute the query with the lower and upper bounds
	row := r.db.QueryRow(query, lowerBound, priceFloat)

	// Scan the result into the rate struct
	err = row.Scan(&rate.RateID, &rate.Clicks, &rate.Limit, &rate.Price, &rate.GroupID)
	if err != nil {
		return rate, err
	}

	return rate, nil
}

func (r *repositoryService) GetAllRates() ([]Rate, error) {
	var rates []Rate

	query := "SELECT rate_id, clicks, \"limit\", price, group_id FROM rate"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rate Rate
		err := rows.Scan(&rate.RateID, &rate.Clicks, &rate.Limit, &rate.Price, &rate.GroupID)
		if err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rates, nil
}
