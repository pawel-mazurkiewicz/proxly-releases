package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/errorsx"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
	"go.uber.org/zap"
)

type LicenseService struct {
	licenseRepo    LicenseRepository
	activationRepo ActivationRepository
	notifier       Notifier
	auditRepo      AuditLogRepository
	gumroadClient  GumroadVerifier
	trialService   *TrialService
	maxActivations int
	autoCreate     bool
	gumroadEnabled bool
	logger         *zap.Logger
}

type LicenseRepository interface {
	Create(license *types.License) error
	GetByID(id uuid.UUID) (*types.License, error)
	GetByKey(key string) (*types.License, error)
	Update(license *types.License) error
	Delete(id uuid.UUID) error
	List(limit, offset int) ([]*types.License, error)
	IncrementActivationCount(id uuid.UUID) error
	SetActivationCount(id uuid.UUID, count int) error
}

type ActivationRepository interface {
	Create(activation *types.Activation) error
	GetByLicenseID(licenseID uuid.UUID) ([]*types.Activation, error)
	DeleteByLicenseID(licenseID uuid.UUID) error
	GetByLicenseIDAndFingerprint(licenseID uuid.UUID, fingerprint string) (*types.Activation, error)
	HasActivationForFingerprint(fingerprint string) (bool, error)
}

type AuditLogRepository interface {
	Create(log *types.AuditLog) error
}

type Notifier interface {
	NotifyThresholdCrossed(license *types.License) error
	NotifyActivationExceeded(license *types.License) error
	NotifyLicenseRevoked(license *types.License) error
}

type GumroadVerifier interface {
	CheckLicenseValidity(licenseKey string) (bool, string, error)
}

func NewLicenseService(
	licenseRepo LicenseRepository,
	activationRepo ActivationRepository,
	notifier Notifier,
	auditRepo AuditLogRepository,
	gumroadClient GumroadVerifier,
	trialService *TrialService,
	maxActivations int,
	autoCreate bool,
	gumroadEnabled bool,
	logger *zap.Logger,
) *LicenseService {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &LicenseService{
		licenseRepo:    licenseRepo,
		activationRepo: activationRepo,
		notifier:       notifier,
		auditRepo:      auditRepo,
		gumroadClient:  gumroadClient,
		trialService:   trialService,
		maxActivations: maxActivations,
		autoCreate:     autoCreate,
		gumroadEnabled: gumroadEnabled,
		logger:         logger,
	}
}

func (s *LicenseService) ProcessActivation(req *types.ActivationRequest, clientIP string) (*types.ActivationResponse, error) {
	normalizedKey := s.normalizeLicenseKey(req.LicenseKey)

	license, err := s.licenseRepo.GetByKey(normalizedKey)
	if err != nil {
		if s.autoCreate {
			if !s.gumroadEnabled {
				return nil, fmt.Errorf("insecure configuration: auto-creation is enabled but Gumroad verification is disabled")
			}

			// If auto-create is enabled, verify with Gumroad first
			valid, reason, err := s.gumroadClient.CheckLicenseValidity(normalizedKey)
			if err != nil {
				return nil, fmt.Errorf("failed to verify license with Gumroad: %w", err)
			}
			if !valid {
				return nil, fmt.Errorf("license verification failed: %s", reason)
			}

			license, err = s.createLicense(normalizedKey)
			if err != nil {
				return nil, fmt.Errorf("failed to auto-create license: %w", err)
			}
		} else {
			return nil, errorsx.ErrLicenseNotFound
		}
	} else {
		// License exists in our database, but if Gumroad verification is enabled,
		// we should still verify it's valid (e.g., not refunded)
		if s.gumroadEnabled {
			valid, reason, err := s.gumroadClient.CheckLicenseValidity(normalizedKey)
			if err != nil {
				return nil, fmt.Errorf("failed to verify license with Gumroad: %w", err)
			}
			if !valid {
				// Mark license as revoked in our database
				license.Status = types.LicenseStatusRevoked
				s.licenseRepo.Update(license)
				return nil, fmt.Errorf("license verification failed: %s", reason)
			}
		}
	}

	// Validate license state first (revoked/suspended/expired), but do not enforce capacity yet
	if err := s.validateLicenseState(license); err != nil {
		return nil, err
	}

	// If this device already has an activation, allow reuse even when at capacity
	if req.DeviceFingerprint != "" {
		existingActivation, err := s.activationRepo.GetByLicenseIDAndFingerprint(license.ID, req.DeviceFingerprint)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing activation: %w", err)
		}
		if existingActivation != nil {
			resp := &types.ActivationResponse{
				LicenseID:            license.ID,
				ActivationCount:      license.ActivationCount,
				MaxActivations:       license.MaxActivations,
				RemainingActivations: license.MaxActivations - license.ActivationCount,
				Status:               s.getActivationStatus(license),
			}
			s.attachTrialInfo(resp, req)
			return resp, nil
		}
	}

	// Enforce capacity only for new devices (no existing activation found)
	if license.ActivationCount >= license.MaxActivations {
		return nil, errorsx.ErrActivationExceeded
	}

	activation := &types.Activation{
		ID:         uuid.New(),
		LicenseID:  license.ID,
		IP:         clientIP,
		UserAgent:  req.UserAgent,
		OccurredAt: time.Now(),
	}

	if req.ClientID != "" {
		activation.ClientID = &req.ClientID
	}

	if req.DeviceFingerprint != "" {
		activation.DeviceFingerprint = &req.DeviceFingerprint
	}

	if err := s.activationRepo.Create(activation); err != nil {
		return nil, fmt.Errorf("failed to create activation: %w", err)
	}

	if err := s.licenseRepo.IncrementActivationCount(license.ID); err != nil {
		return nil, fmt.Errorf("failed to increment activation count: %w", err)
	}

	updatedLicense, err := s.licenseRepo.GetByID(license.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated license: %w", err)
	}

	if s.trialService != nil && req.DeviceFingerprint != "" {
		_ = s.trialService.UpdateTrialStatusByFingerprint(req.DeviceFingerprint, types.TrialStatusConverted)
	}

	if err := s.auditActivation(activation, updatedLicense); err != nil {
		// Log error but don't fail the activation
		s.logger.Warn("failed to audit activation", zap.Error(err))
	}

	response := &types.ActivationResponse{
		LicenseID:            updatedLicense.ID,
		ActivationCount:      updatedLicense.ActivationCount,
		MaxActivations:       updatedLicense.MaxActivations,
		RemainingActivations: updatedLicense.MaxActivations - updatedLicense.ActivationCount,
		Status:               s.getActivationStatus(updatedLicense),
	}
	s.attachTrialInfo(response, req)

	if err := s.processNotifications(updatedLicense); err != nil {
		// Log error but don't fail the activation
		s.logger.Warn("failed to process notifications", zap.Error(err))
	}

	return response, nil
}

func (s *LicenseService) CreateLicense(key string, maxActivations int, metadata []byte) (*types.License, error) {
	normalizedKey := s.normalizeLicenseKey(key)

	if _, err := s.licenseRepo.GetByKey(normalizedKey); err == nil {
		return nil, fmt.Errorf("license key already exists")
	}

	license := &types.License{
		ID:              uuid.New(),
		Key:             normalizedKey,
		MaxActivations:  maxActivations,
		ActivationCount: 0,
		Status:          types.LicenseStatusActive,
		Metadata:        metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.licenseRepo.Create(license); err != nil {
		return nil, fmt.Errorf("failed to create license: %w", err)
	}

	if err := s.auditLicenseCreation(license); err != nil {
		s.logger.Warn("failed to audit license creation", zap.Error(err))
	}

	return license, nil
}

func (s *LicenseService) GetLicense(id uuid.UUID) (*types.License, error) {
	return s.licenseRepo.GetByID(id)
}

func (s *LicenseService) UpdateLicense(license *types.License) error {
	existing, err := s.licenseRepo.GetByID(license.ID)
	if err != nil {
		return err
	}

	if err := s.licenseRepo.Update(license); err != nil {
		return fmt.Errorf("failed to update license: %w", err)
	}

	if existing.Status != license.Status && license.Status == types.LicenseStatusRevoked {
		if err := s.notifier.NotifyLicenseRevoked(license); err != nil {
			s.logger.Warn("failed to notify license revocation", zap.Error(err))
		}
	}

	if err := s.auditLicenseUpdate(existing, license); err != nil {
		s.logger.Warn("failed to audit license update", zap.Error(err))
	}

	return nil
}

func (s *LicenseService) DeleteLicense(id uuid.UUID) error {
	license, err := s.licenseRepo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.licenseRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete license: %w", err)
	}

	if err := s.auditLicenseDeletion(license); err != nil {
		s.logger.Warn("failed to audit license deletion", zap.Error(err))
	}

	return nil
}

func (s *LicenseService) ListLicenses(limit, offset int) ([]*types.License, error) {
	return s.licenseRepo.List(limit, offset)
}

func (s *LicenseService) ListActivations(licenseID uuid.UUID) ([]*types.Activation, error) {
	if _, err := s.licenseRepo.GetByID(licenseID); err != nil {
		return nil, err
	}

	return s.activationRepo.GetByLicenseID(licenseID)
}

func (s *LicenseService) ResetActivations(licenseID uuid.UUID) (*types.License, error) {
	existing, err := s.licenseRepo.GetByID(licenseID)
	if err != nil {
		return nil, err
	}

	if err := s.activationRepo.DeleteByLicenseID(licenseID); err != nil {
		return nil, fmt.Errorf("failed to delete activations: %w", err)
	}

	if err := s.licenseRepo.SetActivationCount(licenseID, 0); err != nil {
		return nil, fmt.Errorf("failed to reset activation count: %w", err)
	}

	refreshed, err := s.licenseRepo.GetByID(licenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to load updated license: %w", err)
	}

	if err := s.auditLicenseUpdate(existing, refreshed); err != nil {
		s.logger.Warn("failed to audit license reset", zap.Error(err))
	}

	return refreshed, nil
}

func (s *LicenseService) normalizeLicenseKey(key string) string {
	return strings.TrimSpace(strings.ToUpper(key))
}

func (s *LicenseService) attachTrialInfo(resp *types.ActivationResponse, req *types.ActivationRequest) {
	if s.trialService == nil || req.DeviceFingerprint == "" {
		return
	}
	status, err := s.trialService.GetStatus(req.ClientID, req.DeviceFingerprint)
	if err != nil || status == nil {
		return
	}
	if status.DaysRemaining != nil {
		resp.TrialDaysRemaining = status.DaysRemaining
	}
	if status.ExpiresAt != nil {
		expires := *status.ExpiresAt
		resp.TrialExpiresAt = &expires
	}
}

func (s *LicenseService) createLicense(key string) (*types.License, error) {
	license := &types.License{
		ID:              uuid.New(),
		Key:             key,
		MaxActivations:  s.maxActivations,
		ActivationCount: 0,
		Status:          types.LicenseStatusActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.licenseRepo.Create(license); err != nil {
		return nil, err
	}

	return license, nil
}

// (Removed validateLicenseStatus; capacity is enforced contextually in ProcessActivation.)

// validateLicenseState ensures the license is in a usable state (not revoked, suspended, or expired),
// but intentionally does not enforce the max-activations capacity. Capacity should be checked
// contextually in flows that may allow existing devices to reuse their activation without consuming
// an additional slot.
func (s *LicenseService) validateLicenseState(license *types.License) error {
	switch license.Status {
	case types.LicenseStatusRevoked:
		return errorsx.ErrLicenseRevoked
	case types.LicenseStatusSuspended:
		return errorsx.ErrLicenseSuspended
	case types.LicenseStatusExpired:
		return errorsx.ErrLicenseExpired
	case types.LicenseStatusActive:
		return nil
	default:
		return fmt.Errorf("unknown license status: %s", license.Status)
	}
}

func (s *LicenseService) getActivationStatus(license *types.License) types.ActivationStatus {
	if license.ActivationCount > license.MaxActivations {
		return types.ActivationStatusExceeded
	} else if license.ActivationCount == license.MaxActivations {
		return types.ActivationStatusAtThreshold
	}
	return types.ActivationStatusOK
}

func (s *LicenseService) processNotifications(license *types.License) error {
	status := s.getActivationStatus(license)

	switch status {
	case types.ActivationStatusAtThreshold:
		return s.notifier.NotifyThresholdCrossed(license)
	case types.ActivationStatusExceeded:
		return s.notifier.NotifyActivationExceeded(license)
	}

	return nil
}

func (s *LicenseService) auditActivation(activation *types.Activation, license *types.License) error {
	payload := map[string]interface{}{
		"activation_id": activation.ID,
		"license_id":    activation.LicenseID,
		"ip":            activation.IP,
		"user_agent":    activation.UserAgent,
		"new_count":     license.ActivationCount,
	}

	payloadBytes, _ := json.Marshal(payload)

	auditLog := &types.AuditLog{
		ID:        uuid.New(),
		Actor:     "system",
		Action:    "activation_created",
		Entity:    "license",
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}

	return s.auditRepo.Create(auditLog)
}

func (s *LicenseService) auditLicenseCreation(license *types.License) error {
	payload := map[string]interface{}{
		"license_id":      license.ID,
		"key":             license.Key,
		"max_activations": license.MaxActivations,
	}

	payloadBytes, _ := json.Marshal(payload)

	auditLog := &types.AuditLog{
		ID:        uuid.New(),
		Actor:     "admin",
		Action:    "license_created",
		Entity:    "license",
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}

	return s.auditRepo.Create(auditLog)
}

func (s *LicenseService) auditLicenseUpdate(old, new *types.License) error {
	payload := map[string]interface{}{
		"license_id": new.ID,
		"old_status": old.Status,
		"new_status": new.Status,
	}

	payloadBytes, _ := json.Marshal(payload)

	auditLog := &types.AuditLog{
		ID:        uuid.New(),
		Actor:     "admin",
		Action:    "license_updated",
		Entity:    "license",
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}

	return s.auditRepo.Create(auditLog)
}

func (s *LicenseService) auditLicenseDeletion(license *types.License) error {
	payload := map[string]interface{}{
		"license_id": license.ID,
		"key":        license.Key,
	}

	payloadBytes, _ := json.Marshal(payload)

	auditLog := &types.AuditLog{
		ID:        uuid.New(),
		Actor:     "admin",
		Action:    "license_deleted",
		Entity:    "license",
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}

	return s.auditRepo.Create(auditLog)
}
