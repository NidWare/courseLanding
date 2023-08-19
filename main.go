package main

import (
	"courseLanding/internal/app"
	"courseLanding/internal/service"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	apiURL     = "https://api.yookassa.ru/v3/payments/"
	username   = "233943"
	password   = "live_UcVgvEGQhiow2l_nJYE-GYSOjt0mBdiqRBbP7N-_xdE"
	dbPath     = "orders.db"
	checkDelay = 5 * time.Minute
)

func main() {
	course := service.NewCourseService()
	// Создание таблицы, если её нет
	err := createTableIfNotExists()
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Запуск фоновой проверки платежей
	go func(course service.CourseService) {
		for {
			checkPayments(course)
			time.Sleep(checkDelay)
		}
	}(course)

	//db
	db, err := sql.Open("sqlite3", "./counter.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	//services => application
	repository := service.NewRepositoryService(db)
	counter := service.NewCounterService()
	payment := service.NewPaymentService()

	application := app.Application{
		CounterService:    counter,
		PaymentService:    payment,
		RepositoryService: repository,
		CourseService:     course,
	}

	//router
	r := mux.NewRouter()
	r.HandleFunc("/buy", application.BuyHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/status", application.StatusHandler).Methods("GET")
	r.HandleFunc("/webhook", application.WebhookHandler).Methods("POST")

	//server
	cert, err := tls.LoadX509KeyPair("/app/cert.pem", "/app/key.pem")
	if err != nil {
		log.Fatalf("failed to load keys: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// other TLS settings here
	}

	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
		Handler:   r,
		// other server settings here
	}

	log.Fatal(server.ListenAndServeTLS("", ""))
}

type PaymentResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Amount struct {
		Value string `json:"value"`
	} `json:"amount"`
}

func createTableIfNotExists() error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS orders (
		payment_id TEXT PRIMARY KEY,
		email TEXT
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return err
	}

	return nil
}

func checkPayments(c service.CourseService) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Изменен запрос для выбора и payment_id, и email
	rows, err := db.Query("SELECT payment_id, email FROM orders")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var paymentID, email string
		if err := rows.Scan(&paymentID, &email); err != nil {
			log.Fatal(err)
		}

		status, err := checkPaymentStatus(paymentID)
		if err != nil {
			log.Printf("Failed to check payment status for %s: %v", paymentID, err)
			continue
		}

		if status == "succeeded" {
			if
			c.Invite(email, 1)
			if _, err := db.Exec("DELETE FROM orders WHERE payment_id = ?", paymentID); err != nil {
				log.Printf("Failed to delete payment %s from orders: %v", paymentID, err)
			}
		}
	}
}

func checkPaymentStatus(paymentID string) (string, string, error) {
	req, err := http.NewRequest("GET", apiURL+paymentID, nil)
	if err != nil {
		return "","", err
	}
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "","", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "","", err
	}

	var paymentResponse PaymentResponse
	if err := json.Unmarshal(body, &paymentResponse); err != nil {
		return "","", err
	}

	return paymentResponse.Status, paymentResponse.Amount.Value, nil
}
