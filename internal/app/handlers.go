package app

import (
	"courseLanding/internal/service"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	var url string
	var err error
	var id string

	phone := convertPhoneNumber(params.Phone)

	clicks, err := a.RepositoryService.GetClicks(params.Rate)
	limit, err := a.RepositoryService.GetLimit(params.Rate)

	switch params.Rate {
	case 1:
		err := a.RepositoryService.IncrementClicks(1)
		if err != nil {
			fmt.Println("failed incrementing")
		}

		if clicks > limit && params.Admin == "" {
			http.Error(w, "Sold out", http.StatusMethodNotAllowed)
			return
		}
		url, id, err = a.PaymentService.MakePayment(10000.00, params.Name, params.Email, phone)
		if err != nil {
			http.Error(w, "Problems with ukassa", http.StatusBadRequest)
			return
		}
	case 2:
		err := a.RepositoryService.IncrementClicks(2)
		if err != nil {
			fmt.Println("failed incrementing")
		}

		if clicks > limit && params.Admin == "" {
			http.Error(w, "Sold out", http.StatusMethodNotAllowed)
			return
		}
		url, id, err = a.PaymentService.MakePayment(20000.00, params.Name, params.Email, phone)
		if err != nil {
			http.Error(w, "Problems with ukassa", http.StatusBadRequest)
			return
		}
	case 3:
		err := a.RepositoryService.IncrementClicks(3)
		if err != nil {
			fmt.Println("failed incrementing")
		}

		if clicks > limit && params.Admin == "" {
			http.Error(w, "Sold out", http.StatusMethodNotAllowed)
			return
		}
		url, id, err = a.PaymentService.MakePayment(35000.00, params.Name, params.Email, phone)
		if err != nil {
			http.Error(w, "Problems with ukassa", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Rate is not found", http.StatusBadRequest)
		return
	}
	fmt.Println("New order id:"+id, " email:"+params.Email, " url:"+url)
	err = insertOrder(id, params.Email)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, url) // здесь воозвращаем УРЛ просто текстом
}

func (a *Application) StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	status, err := ReadBoolFromFile("status.txt")
	var statuses map[int]int
	var resp []byte

	if !status {
		statuses = map[int]int{1: 1, 2: 1, 3: 1}
		resp, err = json.Marshal(statuses)
		w.Write(resp)
		return
	}

	statuses = make(map[int]int)

	for i := 1; i < 4; i++ {
		clicks, _ := a.RepositoryService.GetClicks(i)
		limit, _ := a.RepositoryService.GetLimit(i)

		if clicks >= limit {
			statuses[i] = 3
		} else {
			statuses[i] = 2
		}
		resp, err = json.Marshal(statuses)
	}
	w.Write(resp)
	if err != nil {
		fmt.Println("Error while getting statuses")
	}

}

func (a *Application) LimitHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, sec-ch-ua, sec-ch-ua-mobile, sec-ch-ua-platform")

	// Parse the query parameter 'count'
	countStr := r.URL.Query().Get("count")
	if countStr == "" {
		fmt.Fprintf(w, "Count parameter is missing")
		return
	}

	rate := r.URL.Query().Get("rate")
	if countStr == "" {
		fmt.Fprintf(w, "Count parameter is missing")
		return
	}
	rateForDb, err := strconv.Atoi(rate)

	// Convert count to an integer
	count, err := strconv.Atoi(countStr)
	if err != nil {
		fmt.Fprintf(w, "Invalid count value")
		return
	}

	a.RepositoryService.UpdateLimit(rateForDb, count)
	// Print the count value
	fmt.Println("Count:", count)
}

func (a *Application) EnableHandler(w http.ResponseWriter, r *http.Request) {
	value := FlipBoolInFile("status.txt")

	w.Write([]byte(value))
}

// code standarts ignored:

func ReadBoolFromFile(filename string) (bool, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return false, err
	}

	return string(data) == "1", nil
}

// FlipBoolInFile flips the boolean value in the file.
func FlipBoolInFile(filename string) string {
	value, _ := ReadBoolFromFile(filename)

	newValue := "0"
	if !value {
		newValue = "1"
	}

	ioutil.WriteFile(filename, []byte(newValue), 0666)

	return newValue
}

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
