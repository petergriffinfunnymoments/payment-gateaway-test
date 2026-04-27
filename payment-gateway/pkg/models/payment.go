package models

import "time"

// PaymentMethod определяет метод оплаты
type PaymentMethod string

const (
	MethodSBP       PaymentMethod = "SBP"
	MethodCard      PaymentMethod = "CARD"
	MethodWallet    PaymentMethod = "WALLET"
)

// PaymentStatus статус платежа
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "PENDING"
	StatusProcessing PaymentStatus = "PROCESSING"
	StatusSuccess   PaymentStatus = "SUCCESS"
	StatusFailed    PaymentStatus = "FAILED"
	StatusRefunded  PaymentStatus = "REFUNDED"
)

// Amount структура суммы платежа
type Amount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"` // RUB, USD, EUR
}

// CustomerData данные клиента
type CustomerData struct {
	Email           string `json:"email,omitempty"`
	Phone           string `json:"phone,omitempty"`
	CardNumber      string `json:"card_number,omitempty"`
	CardDate        string `json:"card_date,omitempty"`
	CVV             string `json:"cvv_code,omitempty"`
	DigitalWalletID string `json:"digital_wallet_id,omitempty"`
}

// PaymentInfo основная информация о платеже
type PaymentInfo struct {
	Amount            Amount        `json:"amount"`
	PaymentMethod     PaymentMethod `json:"payment_method"`
	CustomerData      CustomerData  `json:"customer_data,omitempty"`
	Description       string        `json:"description"`
}

// CreatePaymentRequest запрос на создание платежа
type CreatePaymentRequest struct {
	MerchantID      string      `json:"merchant_id"`
	IdempotencyKey  string      `json:"idempotency_key"`
	PaymentInfo     PaymentInfo `json:"payment_info"`
}

// TransactionDetails детали транзакции
type TransactionDetails struct {
	TransactionID   string    `json:"transaction_id"`
	ExternalID      string    `json:"external_id,omitempty"`
	PaymentSystem   string    `json:"payment_system,omitempty"`
	Token           string    `json:"token,omitempty"`
	AntifraudCheck  bool      `json:"antifraud_check"`
	ProcessedAt     time.Time `json:"processed_at,omitempty"`
}

// ErrorDetail деталь ошибки
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PaymentResponse ответ платежного шлюза
type PaymentResponse struct {
	ID                string             `json:"id"`
	MerchantID        string             `json:"merchant_id"`
	Status            PaymentStatus      `json:"status"`
	PaymentInfo       PaymentInfo        `json:"payment_info"`
	TransactionDetails *TransactionDetails `json:"transaction_details,omitempty"`
	Error             *ErrorDetail       `json:"error,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// StatusResponse ответ для проверки статуса
type StatusResponse struct {
	ID         string        `json:"id"`
	Status     PaymentStatus `json:"status"`
	UpdatedAt  time.Time     `json:"updated_at"`
}
