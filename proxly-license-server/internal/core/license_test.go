package core

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock implementations
type MockLicenseRepository struct {
	mock.Mock
}

func (m *MockLicenseRepository) Create(license *types.License) error {
	args := m.Called(license)
	return args.Error(0)
}

func (m *MockLicenseRepository) GetByID(id uuid.UUID) (*types.License, error) {
	args := m.Called(id)
	return args.Get(0).(*types.License), args.Error(1)
}

func (m *MockLicenseRepository) GetByKey(key string) (*types.License, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.License), args.Error(1)
}

func (m *MockLicenseRepository) Update(license *types.License) error {
	args := m.Called(license)
	return args.Error(0)
}

func (m *MockLicenseRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLicenseRepository) List(limit, offset int) ([]*types.License, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*types.License), args.Error(1)
}

func (m *MockLicenseRepository) IncrementActivationCount(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLicenseRepository) SetActivationCount(id uuid.UUID, count int) error {
	args := m.Called(id, count)
	return args.Error(0)
}

type MockActivationRepository struct {
	mock.Mock
}

func (m *MockActivationRepository) Create(activation *types.Activation) error {
	args := m.Called(activation)
	return args.Error(0)
}

func (m *MockActivationRepository) GetByLicenseID(licenseID uuid.UUID) ([]*types.Activation, error) {
	args := m.Called(licenseID)
	return args.Get(0).([]*types.Activation), args.Error(1)
}

func (m *MockActivationRepository) DeleteByLicenseID(licenseID uuid.UUID) error {
	args := m.Called(licenseID)
	return args.Error(0)
}

func (m *MockActivationRepository) GetByLicenseIDAndFingerprint(licenseID uuid.UUID, fingerprint string) (*types.Activation, error) {
	args := m.Called(licenseID, fingerprint)
	if activation, ok := args.Get(0).(*types.Activation); ok {
		return activation, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockActivationRepository) HasActivationForFingerprint(fingerprint string) (bool, error) {
	args := m.Called(fingerprint)
	return args.Bool(0), args.Error(1)
}

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) NotifyThresholdCrossed(license *types.License) error {
	args := m.Called(license)
	return args.Error(0)
}

func (m *MockNotifier) NotifyActivationExceeded(license *types.License) error {
	args := m.Called(license)
	return args.Error(0)
}

func (m *MockNotifier) NotifyLicenseRevoked(license *types.License) error {
	args := m.Called(license)
	return args.Error(0)
}

type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(log *types.AuditLog) error {
	args := m.Called(log)
	return args.Error(0)
}

type MockGumroadVerifier struct {
	mock.Mock
}

func (m *MockGumroadVerifier) CheckLicenseValidity(licenseKey string) (bool, string, error) {
	args := m.Called(licenseKey)
	return args.Bool(0), args.String(1), args.Error(2)
}

func TestLicenseService_ProcessActivation_Success(t *testing.T) {
	// Setup mocks
	licenseRepo := new(MockLicenseRepository)
	activationRepo := new(MockActivationRepository)
	notifier := new(MockNotifier)
	auditRepo := new(MockAuditLogRepository)
	gumroadVerifier := new(MockGumroadVerifier)

	service := NewLicenseService(licenseRepo, activationRepo, notifier, auditRepo, gumroadVerifier, nil, 5, false, false, zap.NewNop())

	// Test data
	licenseID := uuid.New()
	license := &types.License{
		ID:              licenseID,
		Key:             "TEST-KEY",
		MaxActivations:  5,
		ActivationCount: 2,
		Status:          types.LicenseStatusActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	updatedLicense := &types.License{
		ID:              licenseID,
		Key:             "TEST-KEY",
		MaxActivations:  5,
		ActivationCount: 3,
		Status:          types.LicenseStatusActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	req := &types.ActivationRequest{
		LicenseKey: "test-key",
		ClientID:   "test-client",
		UserAgent:  "test-agent",
	}

	// Setup expectations
	licenseRepo.On("GetByKey", "TEST-KEY").Return(license, nil)
	activationRepo.On("Create", mock.AnythingOfType("*types.Activation")).Return(nil)
	licenseRepo.On("IncrementActivationCount", licenseID).Return(nil)
	licenseRepo.On("GetByID", licenseID).Return(updatedLicense, nil)
	auditRepo.On("Create", mock.AnythingOfType("*types.AuditLog")).Return(nil)

	// Execute
	response, err := service.ProcessActivation(req, "127.0.0.1")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, licenseID, response.LicenseID)
	assert.Equal(t, 3, response.ActivationCount)
	assert.Equal(t, 5, response.MaxActivations)
	assert.Equal(t, 2, response.RemainingActivations)
	assert.Equal(t, types.ActivationStatusOK, response.Status)

	// Verify all expectations were met
	licenseRepo.AssertExpectations(t)
	activationRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLicenseService_ProcessActivation_SameFingerprint(t *testing.T) {
	licenseRepo := new(MockLicenseRepository)
	activationRepo := new(MockActivationRepository)
	notifier := new(MockNotifier)
	auditRepo := new(MockAuditLogRepository)
	gumroadVerifier := new(MockGumroadVerifier)

	service := NewLicenseService(licenseRepo, activationRepo, notifier, auditRepo, gumroadVerifier, nil, 5, false, false, zap.NewNop())

	licenseID := uuid.New()
	license := &types.License{
		ID:              licenseID,
		Key:             "TEST-FP",
		MaxActivations:  5,
		ActivationCount: 1,
		Status:          types.LicenseStatusActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	req := &types.ActivationRequest{
		LicenseKey:        "test-fp",
		DeviceFingerprint: "fp-123",
		UserAgent:         "test-agent",
	}

	licenseRepo.On("GetByKey", "TEST-FP").Return(license, nil)
	activationRepo.On("GetByLicenseIDAndFingerprint", licenseID, "fp-123").Return(&types.Activation{ID: uuid.New(), LicenseID: licenseID}, nil)

	response, err := service.ProcessActivation(req, "127.0.0.1")

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, licenseID, response.LicenseID)
	assert.Equal(t, 1, response.ActivationCount)
	assert.Equal(t, 4, response.RemainingActivations)

	activationRepo.AssertNotCalled(t, "Create", mock.Anything)
	licenseRepo.AssertNotCalled(t, "IncrementActivationCount", mock.Anything)
}

func TestLicenseService_ProcessActivation_LicenseNotFound(t *testing.T) {
	// Setup mocks
	licenseRepo := new(MockLicenseRepository)
	activationRepo := new(MockActivationRepository)
	notifier := new(MockNotifier)
	auditRepo := new(MockAuditLogRepository)
	gumroadVerifier := new(MockGumroadVerifier)

	service := NewLicenseService(licenseRepo, activationRepo, notifier, auditRepo, gumroadVerifier, nil, 5, false, false, zap.NewNop())

	req := &types.ActivationRequest{
		LicenseKey: "nonexistent-key",
	}

	// Setup expectations
	licenseRepo.On("GetByKey", "NONEXISTENT-KEY").Return((*types.License)(nil), assert.AnError)

	// Execute
	response, err := service.ProcessActivation(req, "127.0.0.1")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)

	licenseRepo.AssertExpectations(t)
}

func TestLicenseService_CreateLicense_Success(t *testing.T) {
	// Setup mocks
	licenseRepo := new(MockLicenseRepository)
	activationRepo := new(MockActivationRepository)
	notifier := new(MockNotifier)
	auditRepo := new(MockAuditLogRepository)
	gumroadVerifier := new(MockGumroadVerifier)

	service := NewLicenseService(licenseRepo, activationRepo, notifier, auditRepo, gumroadVerifier, nil, 5, false, false, zap.NewNop())

	// Setup expectations
	licenseRepo.On("GetByKey", "TEST-KEY").Return((*types.License)(nil), assert.AnError)
	licenseRepo.On("Create", mock.AnythingOfType("*types.License")).Return(nil)
	auditRepo.On("Create", mock.AnythingOfType("*types.AuditLog")).Return(nil)

	// Execute
	license, err := service.CreateLicense("test-key", 10, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, license)
	assert.Equal(t, "TEST-KEY", license.Key)
	assert.Equal(t, 10, license.MaxActivations)
	assert.Equal(t, 0, license.ActivationCount)
	assert.Equal(t, types.LicenseStatusActive, license.Status)

	licenseRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLicenseService_NormalizeLicenseKey(t *testing.T) {
	service := &LicenseService{}

	tests := []struct {
		input    string
		expected string
	}{
		{"test-key", "TEST-KEY"},
		{"  Test-Key  ", "TEST-KEY"},
		{"ALREADY-UPPER", "ALREADY-UPPER"},
		{"", ""},
	}

	for _, test := range tests {
		result := service.normalizeLicenseKey(test.input)
		assert.Equal(t, test.expected, result)
	}
}
