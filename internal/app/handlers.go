package app

import (
	"courseLanding/internal/config"
	"courseLanding/internal/service"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Application struct {
	CounterService    service.CounterService
	PaymentService    service.PaymentService
	RepositoryService service.RepositoryService
	CourseService     service.CourseService
}

type Webhook struct {
	Type   string  `json:"type"`
	Event  string  `json:"event"`
	Object Payment `json:"object"`
}

// Payment defines the object properties
type Payment struct {
	Amount   Amount   `json:"amount"`
	Metadata Metadata `json:"metadata"`
}

// Amount defines the value and currency
type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// Metadata contains the email
type Metadata struct {
	Email string `json:"email"`
}

func (a *Application) BuyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, sec-ch-ua, sec-ch-ua-mobile, sec-ch-ua-platform")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	//w.Header().Set("Access-Control-Allow-Origin", "https://www.trabun.ai")

	type RequestParams struct {
		Rate  int    `json:"rate"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
		Admin string `json:"admin"`
	}

	var params RequestParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if params.Admin == "" {
		http.Error(w, "Sold out", http.StatusMethodNotAllowed)
		return
	}

	var url string
	var err error
	var id string

	phone := convertPhoneNumber(params.Phone)

	switch params.Rate {
	case 1:
		a.CounterService.Increment(1)
		counter := a.CounterService.GetCounter()
		if counter[0] > config.MaxSell[0] && params.Admin == "" {
			http.Error(w, "Sold out", http.StatusMethodNotAllowed)
			return
		}
		url, id, err = a.PaymentService.MakePayment(10.00, params.Name, params.Email, phone)
		if err != nil {
			http.Error(w, "Problems with ukassa", http.StatusBadRequest)
			return
		}
	case 2:
		a.CounterService.Increment(2)
		counter := a.CounterService.GetCounter()
		if counter[1] > config.MaxSell[1] && params.Admin == "" {
			http.Error(w, "Sold out", http.StatusMethodNotAllowed)
			return
		}
		url, id, err = a.PaymentService.MakePayment(30000.00, params.Name, params.Email, phone)
		if err != nil {
			http.Error(w, "Problems with ukassa", http.StatusBadRequest)
			return
		}
	case 3:
		a.CounterService.Increment(3)
		counter := a.CounterService.GetCounter()
		if counter[2] > config.MaxSell[2] && params.Admin == "" {
			http.Error(w, "Sold out", http.StatusMethodNotAllowed)
			return
		}
		url, id, err = a.PaymentService.MakePayment(60000.00, params.Name, params.Email, phone)
		if err != nil {
			http.Error(w, "Problems with ukassa", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Rate is not found", http.StatusBadRequest)
		return
	}
	fmt.Println("id:"+id, " email:"+params.Email, " url:"+url)
	err = insertOrder(id, params.Email)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, url)
}

func (a *Application) StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	counterService := service.NewCounterService()

	layout := "2006-01-02 15-04-05"
	counter := counterService.GetCounter()
	fmt.Println(counter)
	var statuses []string

	for i := 0; i < len(counter); i++ {
		amount := counter[i]

		if amount >= config.MaxSell[i] {
			statuses = append(statuses, "3")
			continue
		}

		endSell, err := time.Parse(layout, config.EndSell)
		if err != nil {
			fmt.Println(err)
			return
		}

		startSell, err := time.Parse(layout, config.StartSell)
		if err != nil {
			fmt.Println(err)
			return
		}

		now := time.Now()

		if now.Before(startSell) {
			statuses = append(statuses, "1")
			continue
		}

		if endSell.Before(now) {
			statuses = append(statuses, "3")
			continue
		}

		statuses = append(statuses, "2")
	}
	// send here json
	response, err := json.Marshal(statuses)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(response)
	return
}

// code standarts ignored:

func insertOrder(id string, email string) error {
	// Open SQLite database
	db, err := sql.Open("sqlite3", "./orders.db")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement to insert data into "orders" table
	stmt, err := db.Prepare("INSERT INTO orders(payment_id, email) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the statement with provided id and email
	_, err = stmt.Exec(id, email)
	if err != nil {
		return err
	}

	fmt.Println("Inserted order:", id, email)
	return nil
}

func convertPhoneNumber(phoneNumber string) string {
	if strings.HasPrefix(phoneNumber, "+7") {
		return phoneNumber[1:]
	} else if strings.HasPrefix(phoneNumber, "8") {
		return "7" + phoneNumber[1:]
	}
	return phoneNumber
}
