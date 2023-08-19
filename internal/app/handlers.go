package app

import (
	"courseLanding/internal/config"
	"courseLanding/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Application struct {
	CounterService    service.CounterService
	PaymentService    service.PaymentService
	RepositoryService service.RepositoryService
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
	type RequestParams struct {
	}
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
