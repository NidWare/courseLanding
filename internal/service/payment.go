package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

type PaymentService interface {
	MakePayment(value float64, fullName string, email string, phone string) (string, string, error)
}

type paymentService struct {
}

func NewPaymentService() PaymentService {
	return &paymentService{}
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

	var x map[string]string
	err = json.NewDecoder(resp.Body).Decode(&x)
	if err != nil {
		return "", "", err
	}

	fmt.Println(x)
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

func setHeaders(req *http.Request) {
	uid := uuid.New().String()
	req.Header.Set("Idempotence-Key", uid)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic 233943:live_ntj41YKTpjuM97qEj56ykuvAGfqJ13SPRJldPpuqnDg")
}
