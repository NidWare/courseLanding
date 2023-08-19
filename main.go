package main

import (
	"courseLanding/internal/app"
	"courseLanding/internal/service"
	"crypto/tls"
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

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
