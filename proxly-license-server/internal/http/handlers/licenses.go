package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/core"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/errorsx"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
	"go.uber.org/zap"
)

type LicensesHandler struct {
	licenseService *core.LicenseService
	logger         *zap.Logger
}

func NewLicensesHandler(licenseService *core.LicenseService, logger *zap.Logger) *LicensesHandler {
	return &LicensesHandler{
		licenseService: licenseService,
		logger:         logger,
	}
}

type CreateLicenseRequest struct {
	Key            string `json:"key"`
	MaxActivations int    `json:"max_activations"`
	Metadata       string `json:"metadata,omitempty"`
}

type UpdateLicenseRequest struct {
	MaxActivations *int                 `json:"max_activations,omitempty"`
	Status         *types.LicenseStatus `json:"status,omitempty"`
	Metadata       *string              `json:"metadata,omitempty"`
}

type ResetActivationsResponse struct {
	License            *types.License `json:"license"`
	ActivationsCleared int            `json:"activations_cleared"`
}

func (h *LicensesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, errorsx.ErrInvalidPayload)
		return
	}

	if req.Key == "" {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "key is required"))
		return
	}

	if req.MaxActivations <= 0 {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "max_activations must be positive"))
		return
	}

	var metadata []byte
	if req.Metadata != "" {
		// Validate that metadata is valid JSON
		var temp interface{}
		if err := json.Unmarshal([]byte(req.Metadata), &temp); err != nil {
			h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "metadata must be valid JSON"))
			return
		}
		metadata = []byte(req.Metadata)
	}

	license, err := h.licenseService.CreateLicense(req.Key, req.MaxActivations, metadata)
	if err != nil {
		h.logger.Error("Failed to create license", zap.Error(err))
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, license)
}

func (h *LicensesHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid license ID"))
		return
	}

	license, err := h.licenseService.GetLicense(id)
	if err != nil {
		h.writeErrorResponse(w, errorsx.ErrLicenseNotFound)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, license)
}

func (h *LicensesHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid license ID"))
		return
	}

	license, err := h.licenseService.GetLicense(id)
	if err != nil {
		h.writeErrorResponse(w, errorsx.ErrLicenseNotFound)
		return
	}

	var req UpdateLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, errorsx.ErrInvalidPayload)
		return
	}

	if req.MaxActivations != nil {
		if *req.MaxActivations <= 0 {
			h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "max_activations must be positive"))
			return
		}
		license.MaxActivations = *req.MaxActivations
	}

	if req.Status != nil {
		license.Status = *req.Status
	}

	if req.Metadata != nil {
		license.Metadata = []byte(*req.Metadata)
	}

	if err := h.licenseService.UpdateLicense(license); err != nil {
		h.logger.Error("Failed to update license", zap.Error(err))
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, license)
}

func (h *LicensesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid license ID"))
		return
	}

	if err := h.licenseService.DeleteLicense(id); err != nil {
		if err.Error() == "license not found" {
			h.writeErrorResponse(w, errorsx.ErrLicenseNotFound)
		} else {
			h.logger.Error("Failed to delete license", zap.Error(err))
			h.writeErrorResponse(w, errorsx.ErrInternalServer)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LicensesHandler) ListActivations(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid license ID"))
		return
	}

	activations, err := h.licenseService.ListActivations(id)
	if err != nil {
		if err.Error() == "license not found" {
			h.writeErrorResponse(w, errorsx.ErrLicenseNotFound)
		} else {
			h.logger.Error("Failed to list activations", zap.Error(err))
			h.writeErrorResponse(w, errorsx.ErrInternalServer)
		}
		return
	}

	response := map[string]interface{}{
		"license_id":  id,
		"count":       len(activations),
		"activations": activations,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *LicensesHandler) ResetActivations(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid license ID"))
		return
	}

	activations, err := h.licenseService.ListActivations(id)
	if err != nil && err.Error() != "license not found" {
		h.logger.Error("Failed to fetch activations before reset", zap.Error(err))
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}

	license, err := h.licenseService.ResetActivations(id)
	if err != nil {
		if err.Error() == "license not found" {
			h.writeErrorResponse(w, errorsx.ErrLicenseNotFound)
		} else {
			h.logger.Error("Failed to reset activations", zap.Error(err))
			h.writeErrorResponse(w, errorsx.ErrInternalServer)
		}
		return
	}

	cleared := 0
	if activations != nil {
		cleared = len(activations)
	}

	response := ResetActivationsResponse{
		License:            license,
		ActivationsCleared: cleared,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *LicensesHandler) List(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	licenses, err := h.licenseService.ListLicenses(limit, offset)
	if err != nil {
		h.logger.Error("Failed to list licenses", zap.Error(err))
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}

	response := map[string]interface{}{
		"licenses": licenses,
		"limit":    limit,
		"offset":   offset,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *LicensesHandler) writeErrorResponse(w http.ResponseWriter, err *errorsx.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func (h *LicensesHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
