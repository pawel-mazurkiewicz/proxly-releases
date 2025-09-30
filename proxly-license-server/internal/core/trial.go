package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/config"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
)

type TrialRepository interface {
	Create(trial *types.Trial) error
	GetByFingerprint(fingerprint string) (*types.Trial, error)
	GetByID(id uuid.UUID) (*types.Trial, error)
	UpdateStatus(id uuid.UUID, status types.TrialStatus, endedAt *time.Time) error
	Update(trial *types.Trial) error
	List(limit, offset int) ([]*types.Trial, error)
	DeleteOlderThan(threshold time.Time) error
	Delete(id uuid.UUID) error
}

type TrialService struct {
	repo           TrialRepository
	activationRepo ActivationRepository
	cfg            *config.Config
}

var ErrTrialNotFound = errors.New("trial not found")

type TrialUpdate struct {
	Status     *types.TrialStatus
	ExpiresAt  *time.Time
	EndedAt    *time.Time
	EndedAtSet bool
	Metadata   []byte
}

func NewTrialService(repo TrialRepository, activationRepo ActivationRepository, cfg *config.Config) *TrialService {
	return &TrialService{repo: repo, activationRepo: activationRepo, cfg: cfg}
}

func (s *TrialService) StartTrial(clientID, fingerprint, userAgent string) (*types.Trial, error) {
	if fingerprint == "" {
		return nil, fmt.Errorf("device fingerprint required")
	}

	hasLicense, err := s.activationRepo.HasActivationForFingerprint(fingerprint)
	if err != nil {
		return nil, err
	}
	if hasLicense {
		return nil, fmt.Errorf("trial not permitted: license already active")
	}

	existing, err := s.repo.GetByFingerprint(fingerprint)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		now := time.Now()
		switch existing.Status {
		case types.TrialStatusActive:
			if existing.ExpiresAt.After(now) {
				return existing, nil
			}
			// expired but status active; mark expired and fall through
			_ = s.repo.UpdateStatus(existing.ID, types.TrialStatusExpired, &now)
			return nil, fmt.Errorf("trial exhausted")
		case types.TrialStatusConverted:
			return nil, fmt.Errorf("trial already converted to license")
		case types.TrialStatusCancelled, types.TrialStatusExpired:
			return nil, fmt.Errorf("trial exhausted")
		}
	}

	now := time.Now()
	expires := now.Add(time.Duration(s.cfg.TrialLengthDays) * 24 * time.Hour)

	trial := &types.Trial{
		ID:                uuid.New(),
		ClientID:          clientID,
		DeviceFingerprint: fingerprint,
		Status:            types.TrialStatusActive,
		StartedAt:         now,
		ExpiresAt:         expires,
	}

	if err := s.repo.Create(trial); err != nil {
		return nil, err
	}

	return trial, nil
}

func (s *TrialService) GetStatus(clientID, fingerprint string) (*types.TrialStatusResponse, error) {
	trial, err := s.repo.GetByFingerprint(fingerprint)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if trial == nil {
		return &types.TrialStatusResponse{Status: types.TrialStatusNotStarted}, nil
	}

	if trial.Status == types.TrialStatusActive && trial.ExpiresAt.Before(now) {
		if err := s.repo.UpdateStatus(trial.ID, types.TrialStatusExpired, &now); err != nil {
			return nil, err
		}
		trial.Status = types.TrialStatusExpired
		trial.EndedAt = &now
	}

	resp := &types.TrialStatusResponse{Status: trial.Status}
	if trial.ExpiresAt.After(now) && trial.Status == types.TrialStatusActive {
		remaining := int(trial.ExpiresAt.Sub(now).Hours()/24) + 1
		resp.DaysRemaining = &remaining
		expires := trial.ExpiresAt
		resp.ExpiresAt = &expires
	} else if trial.Status == types.TrialStatusExpired || trial.Status == types.TrialStatusConverted {
		expires := trial.ExpiresAt
		resp.ExpiresAt = &expires
	}
	return resp, nil
}

func (s *TrialService) UpdateTrialStatusByFingerprint(fingerprint string, status types.TrialStatus) error {
	trial, err := s.repo.GetByFingerprint(fingerprint)
	if err != nil || trial == nil {
		return err
	}
	if trial.Status == status {
		return nil
	}
	now := time.Now()
	return s.repo.UpdateStatus(trial.ID, status, &now)
}

func (s *TrialService) ListTrials(limit, offset int) ([]*types.Trial, error) {
	return s.repo.List(limit, offset)
}

func (s *TrialService) UpdateTrialStatus(id uuid.UUID, status types.TrialStatus) error {
	var endedAt *time.Time
	if status != types.TrialStatusActive {
		now := time.Now()
		endedAt = &now
	}
	return s.repo.UpdateStatus(id, status, endedAt)
}

func (s *TrialService) UpdateTrial(id uuid.UUID, update TrialUpdate) (*types.Trial, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrTrialNotFound
	}

	if update.Status != nil {
		existing.Status = *update.Status
	}
	if update.ExpiresAt != nil {
		existing.ExpiresAt = *update.ExpiresAt
	}
	if update.EndedAtSet {
		existing.EndedAt = update.EndedAt
	} else if existing.Status == types.TrialStatusActive {
		existing.EndedAt = nil
	}
	if update.Metadata != nil {
		existing.Metadata = update.Metadata
	}

	if existing.Status != types.TrialStatusActive && existing.EndedAt == nil {
		now := time.Now()
		existing.EndedAt = &now
	}

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *TrialService) DeleteTrial(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *TrialService) DeleteOldTrials(retentionDays int) error {
	threshold := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	return s.repo.DeleteOlderThan(threshold)
}

func (s *TrialService) StartCleanupRoutine() {
	if s.cfg.TrialRetentionDays <= 0 {
		return
	}
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			if err := s.DeleteOldTrials(s.cfg.TrialRetentionDays); err != nil {
				fmt.Printf("failed to cleanup trials: %v\n", err)
			}
		}
	}()
}
