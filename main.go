package main

import (
	"courseLanding/internal/app"
	"courseLanding/internal/service"
	"database/sql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
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
	course := service.NewCourseService()

	application := app.Application{
		CounterService:    counter,
		PaymentService:    payment,
		RepositoryService: repository,
		CourseService:     course,
	}

	//router
	r := mux.NewRouter()
	r.HandleFunc("/buy", application.BuyHandler).Methods("POST")
	r.HandleFunc("/status", application.StatusHandler).Methods("GET")
	r.HandleFunc("/webhook", application.WebhookHandler).Methods("POST")

	//server
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8443",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//start
	log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))
}
