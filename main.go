package main

import (
	"courseLanding/internal/app"
	"courseLanding/internal/service"
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

const (
	checkDelay = 1 * time.Minute
)

func main() {
	// HTTP-сервер, который будет перенаправлять на HTTPS
	httpServer := &http.Server{
		Addr: ":80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
		}),
	}

	//dbCounter
	dbCounter, err := sql.Open("sqlite3", "./counter.db")
	if err != nil {
		log.Fatal(err)
	}

	dbOrders, err := sql.Open("sqlite3", "./orders.db")
	if err != nil {
		log.Fatal(err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(dbCounter)

	defer func(dbOrders *sql.DB) {
		err := dbOrders.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(dbOrders)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	//services => application
	repository := service.NewRepositoryService(dbCounter, dbOrders)
	counter := service.NewCounterService()
	course := service.NewCourseService()
	payment := service.NewPaymentService(course, repository)

	// Запуск фоновой проверки платежей
	go func() {
		for {
			payment.CheckPayments()
			time.Sleep(checkDelay)
		}
	}()

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
	r.HandleFunc("/limit", application.LimitHandler).Methods("GET")
	r.HandleFunc("/enable", application.EnableHandler).Methods("GET")

	//server

	certFile := "./fullchain.pem"
	keyFile := "./privkey.pem"

	if err := http.ListenAndServeTLS(":443", certFile, keyFile, r); err != nil {
		log.Fatal(err)
	}
}
