package main

import (
	"courseLanding/internal/service"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"courseLanding/internal/app"
)

type Callback func()

func main() {
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
	application := app.Application{CounterService: counter, PaymentService: payment, RepositoryService: repository}

	//router
	r := mux.NewRouter()
	r.HandleFunc("/buy", application.BuyHandler).Methods("POST")
	r.HandleFunc("/status", application.StatusHandler).Methods("GET")
	r.HandleFunc("/webhook", application.WebhookHandler).Methods("POST")

	//server
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//start
	log.Fatal(srv.ListenAndServe())
}
