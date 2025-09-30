package core

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/config"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
	"github.com/stretchr/testify/require"
)

type inMemoryTrialRepo struct {
	mu     sync.Mutex
	trials map[string]*types.Trial
}

func newInMemoryTrialRepo() *inMemoryTrialRepo {
	return &inMemoryTrialRepo{trials: make(map[string]*types.Trial)}
}

func (r *inMemoryTrialRepo) Create(trial *types.Trial) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *trial
	r.trials[trial.DeviceFingerprint] = &copy
	return nil
}

func (r *inMemoryTrialRepo) GetByFingerprint(fingerprint string) (*types.Trial, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	trial, ok := r.trials[fingerprint]
	if !ok {
		return nil, nil
	}
	copy := *trial
	if trial.EndedAt != nil {
		endedCopy := *trial.EndedAt
		copy.EndedAt = &endedCopy
	}
	return &copy, nil
}

func (r *inMemoryTrialRepo) UpdateStatus(id uuid.UUID, status types.TrialStatus, endedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, trial := range r.trials {
		if trial.ID == id {
			trial.Status = status
			if endedAt != nil {
				endedCopy := *endedAt
				trial.EndedAt = &endedCopy
			}
			trial.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("trial not found")
}

func (r *inMemoryTrialRepo) List(limit, offset int) ([]*types.Trial, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*types.Trial
	for _, trial := range r.trials {
		copy := *trial
		if trial.EndedAt != nil {
			endedCopy := *trial.EndedAt
			copy.EndedAt = &endedCopy
		}
		result = append(result, &copy)
	}
	return result, nil
}

func (r *inMemoryTrialRepo) DeleteOlderThan(threshold time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for fingerprint, trial := range r.trials {
		if trial.ExpiresAt.Before(threshold) {
			delete(r.trials, fingerprint)
		}
	}
	return nil
}

func (r *inMemoryTrialRepo) GetByID(id uuid.UUID) (*types.Trial, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, trial := range r.trials {
		if trial.ID == id {
			copy := *trial
			if trial.EndedAt != nil {
				endedCopy := *trial.EndedAt
				copy.EndedAt = &endedCopy
			}
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *inMemoryTrialRepo) Update(trial *types.Trial) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for fingerprint, existing := range r.trials {
		if existing.ID == trial.ID {
			copy := *trial
			if trial.EndedAt != nil {
				endedCopy := *trial.EndedAt
				copy.EndedAt = &endedCopy
			}
			// Keep current map key (fingerprint) to avoid key changes in tests
			r.trials[fingerprint] = &copy
			return nil
		}
	}
	return fmt.Errorf("trial not found")
}

func (r *inMemoryTrialRepo) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for fingerprint, trial := range r.trials {
		if trial.ID == id {
			delete(r.trials, fingerprint)
			return nil
		}
	}
	return fmt.Errorf("trial not found")
}

type fakeActivationRepo struct {
	mu                 sync.Mutex
	activeFingerprints map[string]bool
}

func newFakeActivationRepo() *fakeActivationRepo {
	return &fakeActivationRepo{activeFingerprints: make(map[string]bool)}
}

func (r *fakeActivationRepo) setActive(fingerprint string, active bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.activeFingerprints[fingerprint] = active
}

func (r *fakeActivationRepo) HasActivationForFingerprint(fingerprint string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.activeFingerprints[fingerprint], nil
}

// Remaining ActivationRepository methods to satisfy the interface.
func (r *fakeActivationRepo) Create(*types.Activation) error                        { return nil }
func (r *fakeActivationRepo) Get(uuid.UUID) (*types.Activation, error)              { return nil, nil }
func (r *fakeActivationRepo) Increment(uuid.UUID) error                             { return nil }
func (r *fakeActivationRepo) Delete(uuid.UUID) error                                { return nil }
func (r *fakeActivationRepo) GetByLicenseID(uuid.UUID) ([]*types.Activation, error) { return nil, nil }
func (r *fakeActivationRepo) DeleteByLicenseID(uuid.UUID) error                     { return nil }
func (r *fakeActivationRepo) GetByLicenseIDAndFingerprint(uuid.UUID, string) (*types.Activation, error) {
	return nil, nil
}

func TestTrialService_StartTrial_NewTrial(t *testing.T) {
	repo := newInMemoryTrialRepo()
	activationRepo := newFakeActivationRepo()
	cfg := &config.Config{TrialLengthDays: 7}
	svc := NewTrialService(repo, activationRepo, cfg)

	before := time.Now()
	trial, err := svc.StartTrial("client-1", "fingerprint-1", "ua")
	require.NoError(t, err)
	require.NotNil(t, trial)
	require.Equal(t, types.TrialStatusActive, trial.Status)

	expected := before.Add(7 * 24 * time.Hour)
	require.WithinDuration(t, expected, trial.ExpiresAt, time.Minute)
}

func TestTrialService_StartTrial_ReusesActive(t *testing.T) {
	repo := newInMemoryTrialRepo()
	activationRepo := newFakeActivationRepo()
	cfg := &config.Config{TrialLengthDays: 7}
	svc := NewTrialService(repo, activationRepo, cfg)

	now := time.Now()
	existing := &types.Trial{
		ID:                uuid.New(),
		ClientID:          "client",
		DeviceFingerprint: "fp",
		Status:            types.TrialStatusActive,
		StartedAt:         now.Add(-24 * time.Hour),
		ExpiresAt:         now.Add(5 * 24 * time.Hour),
	}
	repo.trials["fp"] = existing

	trial, err := svc.StartTrial("client", "fp", "ua")
	require.NoError(t, err)
	require.Equal(t, existing.ID, trial.ID)
	require.Equal(t, types.TrialStatusActive, trial.Status)
}

func TestTrialService_StartTrial_DeniedWhenActivationExists(t *testing.T) {
	repo := newInMemoryTrialRepo()
	activationRepo := newFakeActivationRepo()
	activationRepo.setActive("fp", true)
	cfg := &config.Config{TrialLengthDays: 7}
	svc := NewTrialService(repo, activationRepo, cfg)

	_, err := svc.StartTrial("client", "fp", "ua")
	require.Error(t, err)
	require.Contains(t, err.Error(), "license already active")
}

func TestTrialService_StartTrial_AfterExpiryDenied(t *testing.T) {
	repo := newInMemoryTrialRepo()
	activationRepo := newFakeActivationRepo()
	cfg := &config.Config{TrialLengthDays: 7}
	svc := NewTrialService(repo, activationRepo, cfg)

	expired := &types.Trial{
		ID:                uuid.New(),
		ClientID:          "client",
		DeviceFingerprint: "fp",
		Status:            types.TrialStatusExpired,
		StartedAt:         time.Now().Add(-10 * 24 * time.Hour),
		ExpiresAt:         time.Now().Add(-3 * 24 * time.Hour),
	}
	repo.trials["fp"] = expired

	_, err := svc.StartTrial("client", "fp", "ua")
	require.Error(t, err)
	require.Contains(t, err.Error(), "trial exhausted")
}

func TestTrialService_GetStatus_UpdatesExpired(t *testing.T) {
	repo := newInMemoryTrialRepo()
	activationRepo := newFakeActivationRepo()
	cfg := &config.Config{TrialLengthDays: 7}
	svc := NewTrialService(repo, activationRepo, cfg)

	now := time.Now()
	trial := &types.Trial{
		ID:                uuid.New(),
		ClientID:          "client",
		DeviceFingerprint: "fp",
		Status:            types.TrialStatusActive,
		StartedAt:         now.Add(-8 * 24 * time.Hour),
		ExpiresAt:         now.Add(-24 * time.Hour),
	}
	repo.trials["fp"] = trial

	status, err := svc.GetStatus("client", "fp")
	require.NoError(t, err)
	require.Equal(t, types.TrialStatusExpired, status.Status)

	updated := repo.trials["fp"]
	require.Equal(t, types.TrialStatusExpired, updated.Status)
	require.NotNil(t, updated.EndedAt)
}

func TestTrialService_DeleteOldTrials(t *testing.T) {
	repo := newInMemoryTrialRepo()
	activationRepo := newFakeActivationRepo()
	cfg := &config.Config{TrialLengthDays: 7, TrialRetentionDays: 365}
	svc := NewTrialService(repo, activationRepo, cfg)

	oldTrial := &types.Trial{
		ID:                uuid.New(),
		ClientID:          "client",
		DeviceFingerprint: "old",
		Status:            types.TrialStatusExpired,
		StartedAt:         time.Now().Add(-400 * 24 * time.Hour),
		ExpiresAt:         time.Now().Add(-390 * 24 * time.Hour),
	}
	recentTrial := &types.Trial{
		ID:                uuid.New(),
		ClientID:          "client",
		DeviceFingerprint: "recent",
		Status:            types.TrialStatusActive,
		StartedAt:         time.Now().Add(-5 * 24 * time.Hour),
		ExpiresAt:         time.Now().Add(2 * 24 * time.Hour),
	}
	repo.trials["old"] = oldTrial
	repo.trials["recent"] = recentTrial

	err := svc.DeleteOldTrials(365)
	require.NoError(t, err)

	_, okOld := repo.trials["old"]
	_, okRecent := repo.trials["recent"]
	require.False(t, okOld)
	require.True(t, okRecent)
}
