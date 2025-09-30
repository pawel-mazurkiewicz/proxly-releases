package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	jwtSecret   string
	hmacSecret  string
	timeWindow  time.Duration
}

func NewAuthService(jwtSecret, hmacSecret string) *AuthService {
	return &AuthService{
		jwtSecret:  jwtSecret,
		hmacSecret: hmacSecret,
		timeWindow: 5 * time.Minute,
	}
}

func (a *AuthService) ValidateHMACSignature(timestamp, clientID, body, signature string) error {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format")
	}

	requestTime := time.Unix(ts, 0)
	if time.Since(requestTime) > a.timeWindow || time.Until(requestTime) > a.timeWindow {
		return fmt.Errorf("request timestamp too skewed")
	}

	expectedSignature := a.generateHMACSignature(timestamp, clientID, body)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (a *AuthService) generateHMACSignature(timestamp, clientID, body string) string {
	data := timestamp + clientID + body
	h := hmac.New(sha256.New, []byte(a.hmacSecret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (a *AuthService) ValidateJWTToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func (a *AuthService) GenerateJWTToken(claims jwt.MapClaims) (string, error) {
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.jwtSecret))
}