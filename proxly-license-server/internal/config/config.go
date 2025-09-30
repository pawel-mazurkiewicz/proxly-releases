package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                       string
	DBDriver                   string
	DBDSN                      string
	JWTAdminSecret             string
	ClientSigAlgo              string
	HMACSecret                 string
	DefaultMaxActivations      int
	AutoCreateOnActivation     bool
	WebhookTimeoutMS           int
	RateLimitPerLicensePerMin  int
	RateLimitPerIPPerMin       int
	LogLevel                   string
	GumroadProductID           string
	GumroadVerificationEnabled bool
	TrialLengthDays            int
	TrialRetentionDays         int
}

func Load() *Config {
	return &Config{
		Port:                       getEnv("APP_PORT", "8080"),
		DBDriver:                   getEnv("DB_DRIVER", "postgres"),
		DBDSN:                      getEnv("DB_DSN", "postgres://proxly:proxly@localhost:5432/proxly?sslmode=disable"),
		JWTAdminSecret:             getEnv("JWT_ADMIN_SECRET", "change-me"),
		ClientSigAlgo:              getEnv("CLIENT_SIG_ALGO", "hmac"),
		HMACSecret:                 getEnv("HMAC_SECRET", "hmac-secret-key"),
		DefaultMaxActivations:      getEnvAsInt("DEFAULT_MAX_ACTIVATIONS", 5),
		AutoCreateOnActivation:     getEnvAsBool("AUTO_CREATE_ON_ACTIVATION", false),
		WebhookTimeoutMS:           getEnvAsInt("WEBHOOK_TIMEOUT_MS", 3000),
		RateLimitPerLicensePerMin:  getEnvAsInt("RATE_LIMIT_PER_LICENSE_PER_MIN", 10),
		RateLimitPerIPPerMin:       getEnvAsInt("RATE_LIMIT_PER_IP_PER_MIN", 30),
		LogLevel:                   getEnv("LOG_LEVEL", "info"),
		GumroadProductID:           getEnv("GUMROAD_PRODUCT_ID", "byTgBXjYmc63gtN34CtW1A=="),
		GumroadVerificationEnabled: getEnvAsBool("GUMROAD_VERIFICATION_ENABLED", true),
		TrialLengthDays:           getEnvAsInt("TRIAL_LENGTH_DAYS", 7),
		TrialRetentionDays:        getEnvAsInt("TRIAL_RETENTION_DAYS", 365),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
