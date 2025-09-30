package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/core"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/errorsx"
	"go.uber.org/zap"
)

type contextKey string

const (
	ClientIPKey  contextKey = "client_ip"
	TimestampKey contextKey = "timestamp"
	UserAgentKey contextKey = "user_agent"
)

func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Timestamp, X-Client-ID, X-Signature")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func Security() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimit(ipLimiter, licenseLimiter *core.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if ipLimiter != nil && !ipLimiter.Allow(clientIP) {
				writeErrorResponse(w, errorsx.ErrRateLimitExceeded)
				return
			}

			if licenseLimiter != nil && r.Method == http.MethodPost && r.URL.Path == "/v1/activate" {
				var body []byte
				if r.Body != nil {
					var err error
					body, err = io.ReadAll(r.Body)
					if err != nil {
						writeErrorResponse(w, errorsx.ErrInvalidPayload)
						return
					}
					r.Body = io.NopCloser(bytes.NewBuffer(body))
				}

				if len(body) > 0 {
					var payload struct {
						LicenseKey string `json:"license_key"`
					}
					if err := json.Unmarshal(body, &payload); err == nil {
						licenseKey := strings.TrimSpace(strings.ToLower(payload.LicenseKey))
						if licenseKey != "" && !licenseLimiter.Allow(licenseKey) {
							writeErrorResponse(w, errorsx.ErrRateLimitExceeded)
							return
						}
					}
					r.Body = io.NopCloser(bytes.NewBuffer(body))
				}
			}

			// Store client IP in context for use in handlers
			ctx := context.WithValue(r.Context(), ClientIPKey, clientIP)
			ctx = context.WithValue(ctx, UserAgentKey, r.UserAgent())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func HMACAuth(authService *core.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timestamp := r.Header.Get("X-Timestamp")
			clientID := r.Header.Get("X-Client-ID")
			signature := r.Header.Get("X-Signature")

			if timestamp == "" || signature == "" {
				writeErrorResponse(w, errorsx.ErrInvalidSignature)
				return
			}

			// Read body for signature verification
			body := []byte{}
			if r.Body != nil {
				var err error
				body, err = io.ReadAll(r.Body)
				if err != nil {
					writeErrorResponse(w, errorsx.ErrInvalidPayload)
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}

			if err := authService.ValidateHMACSignature(timestamp, clientID, string(body), signature); err != nil {
				if strings.Contains(err.Error(), "timestamp") {
					writeErrorResponse(w, errorsx.ErrTimestampSkew)
				} else {
					writeErrorResponse(w, errorsx.ErrInvalidSignature)
				}
				return
			}

			ctx := context.WithValue(r.Context(), TimestampKey, timestamp)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JWTAuth(authService *core.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeErrorResponse(w, errorsx.ErrUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				writeErrorResponse(w, errorsx.ErrUnauthorized)
				return
			}

			_, err := authService.ValidateJWTToken(tokenString)
			if err != nil {
				writeErrorResponse(w, errorsx.ErrUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check for X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip, _, _ = strings.Cut(ip, ":")
	}

	return ip
}

func writeErrorResponse(w http.ResponseWriter, err *errorsx.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}
