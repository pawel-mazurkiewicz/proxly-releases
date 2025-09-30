package types

import (
	"github.com/google/uuid"
	"time"
)

type LicenseStatus string

const (
	LicenseStatusActive    LicenseStatus = "active"
	LicenseStatusRevoked   LicenseStatus = "revoked"
	LicenseStatusSuspended LicenseStatus = "suspended"
	LicenseStatusExpired   LicenseStatus = "expired"
)

type ActivationStatus string

const (
	ActivationStatusOK          ActivationStatus = "ok"
	ActivationStatusAtThreshold ActivationStatus = "at_threshold"
	ActivationStatusExceeded    ActivationStatus = "exceeded"
)

type License struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	Key             string        `json:"key" db:"key"`
	MaxActivations  int           `json:"max_activations" db:"max_activations"`
	ActivationCount int           `json:"activation_count" db:"activation_count"`
	Status          LicenseStatus `json:"status" db:"status"`
	Metadata        []byte        `json:"metadata" db:"metadata"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

type Activation struct {
	ID                uuid.UUID `json:"id" db:"id"`
	LicenseID         uuid.UUID `json:"license_id" db:"license_id"`
	ClientID          *string   `json:"client_id,omitempty" db:"client_id"`
	DeviceFingerprint *string   `json:"device_fingerprint,omitempty" db:"device_fingerprint"`
	IP                string    `json:"ip" db:"ip"`
	UserAgent         string    `json:"user_agent" db:"user_agent"`
	OccurredAt        time.Time `json:"occurred_at" db:"occurred_at"`
}

type Webhook struct {
	ID     uuid.UUID `json:"id" db:"id"`
	URL    string    `json:"url" db:"url"`
	Secret *string   `json:"secret,omitempty" db:"secret"`
	Events []string  `json:"events" db:"events"`
}

type AuditLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Actor     string    `json:"actor" db:"actor"`
	Action    string    `json:"action" db:"action"`
	Entity    string    `json:"entity" db:"entity"`
	Payload   []byte    `json:"payload" db:"payload"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ActivationRequest struct {
	LicenseKey        string `json:"license_key"`
	ClientID          string `json:"client_id,omitempty"`
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
	UserAgent         string `json:"user_agent,omitempty"`
}

type ActivationResponse struct {
	LicenseID            uuid.UUID        `json:"license_id"`
	ActivationCount      int              `json:"activation_count"`
	MaxActivations       int              `json:"max_activations"`
	RemainingActivations int              `json:"remaining_activations"`
	Status               ActivationStatus `json:"status"`
	TrialDaysRemaining   *int             `json:"trial_days_remaining,omitempty"`
	TrialExpiresAt       *time.Time       `json:"trial_expires_at,omitempty"`
}

type TrialStatus string

const (
	TrialStatusActive     TrialStatus = "active"
	TrialStatusExpired    TrialStatus = "expired"
	TrialStatusCancelled  TrialStatus = "cancelled"
	TrialStatusConverted  TrialStatus = "converted"
	TrialStatusNotStarted TrialStatus = "not_started"
)

type Trial struct {
	ID                uuid.UUID   `json:"id" db:"id"`
	ClientID          string      `json:"client_id" db:"client_id"`
	DeviceFingerprint string      `json:"device_fingerprint" db:"device_fingerprint"`
	Status            TrialStatus `json:"status" db:"status"`
	StartedAt         time.Time   `json:"started_at" db:"started_at"`
	ExpiresAt         time.Time   `json:"expires_at" db:"expires_at"`
	EndedAt           *time.Time  `json:"ended_at,omitempty" db:"ended_at"`
	Metadata          []byte      `json:"metadata" db:"metadata"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
}

type TrialStartRequest struct {
	ClientID          string `json:"client_id"`
	DeviceFingerprint string `json:"device_fingerprint"`
	UserAgent         string `json:"user_agent,omitempty"`
}

type TrialStartResponse struct {
	ID            uuid.UUID   `json:"id"`
	Status        TrialStatus `json:"status"`
	ExpiresAt     time.Time   `json:"expires_at"`
	DaysRemaining int         `json:"days_remaining"`
}

type TrialStatusResponse struct {
	Status        TrialStatus `json:"status"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
	DaysRemaining *int        `json:"days_remaining,omitempty"`
}
