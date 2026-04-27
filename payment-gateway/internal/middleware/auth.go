package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const (
	MerchantIDKey contextKey = "merchant_id"
	APIKeyHeader  string     = "X-API-Key"
)

var ErrInvalidAPIKey = errors.New("invalid or missing API key")

// APIKeyMiddleware проверяет API ключ мерчанта
func APIKeyMiddleware(validKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(APIKeyHeader)
			
			if apiKey == "" {
				http.Error(w, `{"error": {"code": "UNAUTHORIZED", "message": "API key is required"}}`, http.StatusUnauthorized)
				return
			}

			merchantID, exists := validKeys[apiKey]
			if !exists {
				http.Error(w, `{"error": {"code": "FORBIDDEN", "message": "Invalid API key"}}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), MerchantIDKey, merchantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetMerchantID извлекает ID мерчанта из контекста
func GetMerchantID(ctx context.Context) (string, bool) {
	merchantID, ok := ctx.Value(MerchantIDKey).(string)
	return merchantID, ok
}

// CORSMiddleware добавляет заголовки CORS
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Idempotency-Key")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// ContentTypeMiddleware устанавливает Content-Type
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "OPTIONS" {
			contentType := r.Header.Get("Content-Type")
			if contentType == "" && (r.Method == "POST" || r.Method == "PUT") {
				http.Error(w, `{"error": {"code": "UNSUPPORTED_MEDIA_TYPE", "message": "Content-Type header is required"}}`, http.StatusUnsupportedMediaType)
				return
			}
			if !strings.Contains(contentType, "application/json") && contentType != "" {
				http.Error(w, `{"error": {"code": "UNSUPPORTED_MEDIA_TYPE", "message": "Only application/json is supported"}}`, http.StatusUnsupportedMediaType)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
