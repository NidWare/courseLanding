package app

import (
	"courseLanding/internal/config"
	"courseLanding/internal/service"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
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
	type RequestParams struct {
		Rate  int    `json:"rate"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	var params RequestParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var url string
	var err error

	switch params.Rate {
	case 1:
		url, err = a.PaymentService.MakePayment(15000.00, params.Name, params.Email, params.Phone)
		if err != nil {
			http.Error(w, "Problems with ukassa", http.StatusBadRequest)
			return
		}
	case 2:
		url, err = a.PaymentService.MakePayment(30000.00, params.Name, params.Email, params.Phone)
		if err != nil {
			http.Error(w, "Problems with ukassa", http.StatusBadRequest)
			return
		}
	case 3:
		url, err = a.PaymentService.MakePayment(60000.00, params.Name, params.Email, params.Phone)
		if err != nil {
			http.Error(w, "Problems with ukassa", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Rate is not found", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, url)
}

func (a *Application) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var webhook Webhook
	err = json.Unmarshal(body, &webhook)
	if err != nil {
		http.Error(w, "Error unmarshaling JSON", http.StatusBadRequest)
		return
	}

	value, err := strconv.ParseFloat(webhook.Object.Amount.Value, 64)
	if err != nil {
		http.Error(w, "Error converting value to float", http.StatusBadRequest)
		return
	}

	fmt.Printf("Email: %s\n", webhook.Object.Metadata.Email)
	fmt.Printf("Value: %.2f\n", value)

	if value == 15000.00 {
		a.CourseService.Invite(webhook.Object.Metadata.Email, 1)
	} else if value == 30000.00 {
		a.CourseService.Invite(webhook.Object.Metadata.Email, 2)
	} else if value == 60000.00 {
		a.CourseService.Invite(webhook.Object.Metadata.Email, 3)
	}

	// Log the structure to a string
	logString, err := json.MarshalIndent(webhook, "", "  ")
	if err != nil {
		http.Error(w, "Error marshaling JSON for log", http.StatusInternalServerError)
		return
	}

	// Write the log string to a text file
	err = ioutil.WriteFile("../log/webhook_log.txt", logString, 0644)
	if err != nil {
		http.Error(w, "Error writing to log file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (a *Application) StatusHandler(w http.ResponseWriter, r *http.Request) {
	counterService := service.NewCounterService()
	if counterService.GetCounter() >= config.MaxSell {
		fmt.Fprintf(w, "3")
		return
	}

	layout := "2006-01-02 15-04-05"

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
		fmt.Fprintf(w, "1")
		return
	}

	if endSell.Before(now) {
		fmt.Fprintf(w, "3")
		return
	}

	fmt.Fprintf(w, "2")
	return
}
