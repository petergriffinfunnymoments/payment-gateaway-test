package main

import (
	"log"
	"net/http"
	"os"

	"payment-gateway/internal/middleware"
	httpHandler "payment-gateway/internal/transport/http"
)

func main() {
	// Конфигурация API ключей мерчантов (в реальности будет в БД или конфиге)
	validAPIKeys := map[string]string{
		"test_merchant_key_123": "merchant_001",
		"demo_shop_key_456":     "merchant_002",
	}

	// Создаем сервис платежей (пока мок)
	paymentService := &httpHandler.MockPaymentService{}

	// Создаем HTTP обработчик
	handler := httpHandler.NewHandler(paymentService)

	// Создаем мультиплексор маршрутов
	mux := http.NewServeMux()

	// Регистрируем эндпоинты
	mux.HandleFunc("/api/v1/payments", handler.CreatePaymentHandler)
	mux.HandleFunc("/api/v1/payments/status", handler.GetStatusHandler)

	// Применяем middleware
	var httpHandler http.Handler = mux
	httpHandler = middleware.ContentTypeMiddleware(httpHandler)
	httpHandler = middleware.CORSMiddleware(httpHandler)
	httpHandler = middleware.APIKeyMiddleware(validAPIKeys)(httpHandler)

	// Получаем порт из переменной окружения или используем значение по умолчанию
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Payment Gateway API on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  POST /api/v1/payments - Create payment")
	log.Printf("  GET  /api/v1/payments/status?id={id} - Get payment status")
	log.Printf("\nTest API Keys:")
	log.Printf("  test_merchant_key_123 -> merchant_001")
	log.Printf("  demo_shop_key_456 -> merchant_002")

	if err := http.ListenAndServe(":"+port, httpHandler); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
