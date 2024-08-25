package service

import (
	"bytes"
	"courseLanding/internal/config"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	apiURL = "https://api.yookassa.ru/v3/payments/"
)

type PaymentService interface {
	MakePayment(value float64, fullName string, email string, phone string) (string, string, error)
	CheckPayments()
}

type paymentService struct {
	c CourseService
	r RepositoryService
}

type PaymentResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Amount struct {
		Value string `json:"value"`
	} `json:"amount"`
}

func NewPaymentService(c CourseService, r RepositoryService) PaymentService {
	return &paymentService{c: c, r: r}
}

func (p *paymentService) MakePayment(value float64, fullName string, email string, phone string) (string, string, error) {
	req, err := createPaymentRequest(value, fullName, email, phone)
	if err != nil {
		return "", "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var paymentResponse struct {
		Confirmation struct {
			ConfirmationURL string `json:"confirmation_url"`
		} `json:"confirmation"`
		ID string `json:"id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&paymentResponse)
	if err != nil {
		return "", "", err
	}

	confirmationURL := paymentResponse.Confirmation.ConfirmationURL
	paymentID := paymentResponse.ID

	return confirmationURL, paymentID, nil
}

func createPaymentRequest(value float64, fullName string, email string, phone string) (*http.Request, error) {
	data := createPaymentDataPayload(value, fullName, email, phone)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.yookassa.ru/v3/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	setHeaders(req)
	return req, nil
}

func createPaymentDataPayload(value float64, fullName, email string, phone string) map[string]interface{} {
	return map[string]interface{}{
		"metadata": map[string]string{
			"email": email,
		},
		"amount": map[string]any{
			"value":    value,
			"currency": "RUB",
		},
		"capture": true,
		"confirmation": map[string]string{
			"type":       "redirect",
			"return_url": "https://www.trabun.ai/",
		},
		"description": "Заказ №1",
		"receipt": map[string]interface{}{
			"customer": map[string]string{
				"full_name": fullName,
				"email":     email,
				"phone":     phone,
			},
			"items": []map[string]interface{}{
				{
					"description":     "Курс",
					"amount":          map[string]any{"value": value, "currency": "RUB"},
					"vat_code":        4,
					"quantity":        "1",
					"payment_subject": "service",
					"payment_mode":    "full_payment",
				},
			},
		},
	}
}

func (p *paymentService) CheckPayments() {
	fmt.Println("Started to check course payments:")
	paymentsByIds := p.r.LoadOrders()
	var paymentsToDelete []string

	for paymentId, email := range paymentsByIds {
		status, amount, err := checkPaymentStatus(paymentId)
		if err != nil {
			log.Printf("Failed to check payment status for %s: %v", paymentId, err)
			continue
		}

		if status == "succeeded" {
			paymentsToDelete = append(paymentsToDelete, paymentId)

			rate, err := p.r.GetRateByPrice(amount)
			if err != nil {
				fmt.Printf("Error during getting rate for amount: %d", amount)
			}

			p.c.Invite(email, rate.GroupID)
		}
	}

	p.r.DeleteOrdersByIds(paymentsToDelete)
}

func checkPaymentStatus(paymentID string) (string, string, error) {
	req, err := http.NewRequest("GET", apiURL+paymentID, nil)
	if err != nil {
		return "", "", err
	}

	req.SetBasicAuth(config.Username, config.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var paymentResponse PaymentResponse
	if err := json.Unmarshal(body, &paymentResponse); err != nil {
		return "", "", err
	}

	return paymentResponse.Status, paymentResponse.Amount.Value, nil
}

func setHeaders(req *http.Request) {
	uid := uuid.New().String()
	req.Header.Set("Idempotence-Key", uid)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(config.Username, config.Password)
}
