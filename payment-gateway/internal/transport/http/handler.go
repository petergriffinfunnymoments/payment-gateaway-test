package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"payment-gateway/internal/middleware"
	"payment-gateway/pkg/models"
)

// PaymentService интерфейс сервиса платежей
type PaymentService interface {
	CreatePayment(merchantID string, req models.CreatePaymentRequest) (*models.PaymentResponse, error)
	GetPaymentStatus(id string) (*models.StatusResponse, error)
}

// Handler обработчик HTTP запросов
type Handler struct {
	service PaymentService
}

// NewHandler создает новый обработчик
func NewHandler(service PaymentService) *Handler {
	return &Handler{service: service}
}

// CreatePaymentHandler обрабатывает запрос на создание платежа
func (h *Handler) CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "METHOD_NOT_ALLOWED", "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем merchant_id из контекста (после middleware аутентификации)
	merchantID, ok := middleware.GetMerchantID(r.Context())
	if !ok {
		h.sendError(w, "UNAUTHORIZED", "Merchant ID not found", http.StatusUnauthorized)
		return
	}

	// Проверяем идемпотентность
	idempotencyKey := r.Header.Get("X-Idempotency-Key")
	if idempotencyKey == "" {
		h.sendError(w, "BAD_REQUEST", "X-Idempotency-Key header is required", http.StatusBadRequest)
		return
	}

	// Парсим запрос
	var req models.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "INVALID_JSON", "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Устанавливаем merchant_id и idempotency_key из запроса/контекста
	req.MerchantID = merchantID
	req.IdempotencyKey = idempotencyKey

	// Валидация базовых полей
	if err := h.validateRequest(req); err != nil {
		h.sendError(w, "VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
		return
	}

	// Вызываем сервис создания платежа
	response, err := h.service.CreatePayment(merchantID, req)
	if err != nil {
		h.sendError(w, "PAYMENT_FAILED", "Failed to create payment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, response, http.StatusCreated)
}

// GetStatusHandler обрабатывает запрос на получение статуса платежа
func (h *Handler) GetStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "METHOD_NOT_ALLOWED", "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID платежа из URL
	paymentID := r.URL.Query().Get("id")
	if paymentID == "" {
		h.sendError(w, "BAD_REQUEST", "Payment ID is required", http.StatusBadRequest)
		return
	}

	response, err := h.service.GetPaymentStatus(paymentID)
	if err != nil {
		h.sendError(w, "NOT_FOUND", "Payment not found", http.StatusNotFound)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

// validateRequest выполняет базовую валидацию запроса
func (h *Handler) validateRequest(req models.CreatePaymentRequest) error {
	if req.PaymentInfo.Amount.Value <= 0 {
		return &json.UnmarshalTypeError{Value: "amount must be positive"}
	}

	if req.PaymentInfo.Amount.Currency == "" {
		return &json.UnmarshalTypeError{Value: "currency is required"}
	}

	if req.PaymentInfo.PaymentMethod == "" {
		return &json.UnmarshalTypeError{Value: "payment_method is required"}
	}

	// Проверка метода оплаты
	switch req.PaymentInfo.PaymentMethod {
	case models.MethodSBP, models.MethodCard, models.MethodWallet:
		// OK
	default:
		return &json.UnmarshalTypeError{Value: "invalid payment_method"}
	}

	// Валидация данных клиента в зависимости от метода оплаты
	switch req.PaymentInfo.PaymentMethod {
	case models.MethodSBP:
		if req.PaymentInfo.CustomerData.Phone == "" {
			return &json.UnmarshalTypeError{Value: "phone is required for SBP"}
		}
	case models.MethodCard:
		if req.PaymentInfo.CustomerData.CardNumber == "" {
			return &json.UnmarshalTypeError{Value: "card_number is required for CARD"}
		}
		if req.PaymentInfo.CustomerData.CVV == "" {
			return &json.UnmarshalTypeError{Value: "cvv_code is required for CARD"}
		}
	case models.MethodWallet:
		if req.PaymentInfo.CustomerData.DigitalWalletID == "" {
			return &json.UnmarshalTypeError{Value: "digital_wallet_id is required for WALLET"}
		}
	}

	return nil
}

// sendJSON отправляет JSON ответ
func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError отправляет ошибку в формате JSON
func (h *Handler) sendError(w http.ResponseWriter, code, message string, status int) {
	errorResp := map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
	h.sendJSON(w, errorResp, status)
}

// MockPaymentService временная реализация сервиса для тестирования
type MockPaymentService struct{}

func (s *MockPaymentService) CreatePayment(merchantID string, req models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	now := time.Now()
	
	// Генерируем ID транзакции
	transactionID := uuid.New().String()

	response := &models.PaymentResponse{
		ID:         transactionID,
		MerchantID: merchantID,
		Status:     models.StatusPending,
		PaymentInfo: req.PaymentInfo,
		TransactionDetails: &models.TransactionDetails{
			TransactionID:  transactionID,
			AntifraudCheck: false,
			ProcessedAt:    now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	return response, nil
}

func (s *MockPaymentService) GetPaymentStatus(id string) (*models.StatusResponse, error) {
	// В реальной реализации здесь будет запрос к БД
	return &models.StatusResponse{
		ID:        id,
		Status:    models.StatusPending,
		UpdatedAt: time.Now(),
	}, nil
}
