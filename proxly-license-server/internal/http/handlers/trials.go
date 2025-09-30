package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/core"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/errorsx"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
)

type TrialsHandler struct {
	trialService *core.TrialService
}

func NewTrialsHandler(trialService *core.TrialService) *TrialsHandler {
	return &TrialsHandler{trialService: trialService}
}

func (h *TrialsHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req types.TrialStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, errorsx.ErrInvalidPayload)
		return
	}
	if req.ClientID == "" || req.DeviceFingerprint == "" {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "client_id and device_fingerprint are required"))
		return
	}

	trial, err := h.trialService.StartTrial(req.ClientID, req.DeviceFingerprint, req.UserAgent)
	if err != nil {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusForbidden, err.Error()))
		return
	}

	daysRemaining := int(trial.ExpiresAt.Sub(time.Now()).Hours()/24) + 1
	resp := types.TrialStartResponse{
		ID:            trial.ID,
		Status:        trial.Status,
		ExpiresAt:     trial.ExpiresAt,
		DaysRemaining: daysRemaining,
	}
	h.writeJSONResponse(w, http.StatusOK, resp)
}

func (h *TrialsHandler) Status(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	fingerprint := r.URL.Query().Get("device_fingerprint")
	if fingerprint == "" {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "device_fingerprint is required"))
		return
	}

	status, err := h.trialService.GetStatus(clientID, fingerprint)
	if err != nil {
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, status)
}

func (h *TrialsHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r, 50)
	trials, err := h.trialService.ListTrials(limit, offset)
	if err != nil {
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"trials": trials,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *TrialsHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid trial id"))
		return
	}

	var req struct {
		Status types.TrialStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, errorsx.ErrInvalidPayload)
		return
	}

	if req.Status == "" {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "status is required"))
		return
	}

	if err := h.trialService.UpdateTrialStatus(id, req.Status); err != nil {
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TrialsHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid trial id"))
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		h.writeErrorResponse(w, errorsx.ErrInvalidPayload)
		return
	}

	var update core.TrialUpdate
	var fieldsUpdated bool

	if statusRaw, ok := raw["status"]; ok {
		var status types.TrialStatus
		if err := json.Unmarshal(statusRaw, &status); err != nil {
			h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid status value"))
			return
		}
		update.Status = &status
		fieldsUpdated = true
	}

	if expiresRaw, ok := raw["expires_at"]; ok {
		if string(expiresRaw) == "null" {
			h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "expires_at cannot be null"))
			return
		}
		var expiresStr string
		if err := json.Unmarshal(expiresRaw, &expiresStr); err != nil {
			h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid expires_at format"))
			return
		}
		expiresAt, err := time.Parse(time.RFC3339, expiresStr)
		if err != nil {
			h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "expires_at must be RFC3339"))
			return
		}
		update.ExpiresAt = &expiresAt
		fieldsUpdated = true
	}

	if endedRaw, ok := raw["ended_at"]; ok {
		update.EndedAtSet = true
		if string(endedRaw) == "null" {
			update.EndedAt = nil
		} else {
			var endedStr string
			if err := json.Unmarshal(endedRaw, &endedStr); err != nil {
				h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid ended_at format"))
				return
			}
			endedAt, err := time.Parse(time.RFC3339, endedStr)
			if err != nil {
				h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "ended_at must be RFC3339"))
				return
			}
			update.EndedAt = &endedAt
		}
		fieldsUpdated = true
	}

	if metadataRaw, ok := raw["metadata"]; ok {
		if string(metadataRaw) == "null" {
			update.Metadata = nil
		} else {
			update.Metadata = metadataRaw
		}
		fieldsUpdated = true
	}

	if !fieldsUpdated {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "no updatable fields provided"))
		return
	}

        trial, err := h.trialService.UpdateTrial(id, update)
        if err != nil {
                if errors.Is(err, core.ErrTrialNotFound) {
                        h.writeErrorResponse(w, errorsx.NewError(http.StatusNotFound, "trial not found"))
                        return
                }
                h.writeErrorResponse(w, errorsx.ErrInternalServer)
                return
        }

	h.writeJSONResponse(w, http.StatusOK, trial)
}

func (h *TrialsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "invalid trial id"))
		return
	}

        if err := h.trialService.DeleteTrial(id); err != nil {
                if errors.Is(err, sql.ErrNoRows) || errors.Is(err, core.ErrTrialNotFound) {
                        h.writeErrorResponse(w, errorsx.NewError(http.StatusNotFound, "trial not found"))
                        return
                }
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TrialsHandler) writeErrorResponse(w http.ResponseWriter, err *errorsx.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	_ = json.NewEncoder(w).Encode(err)
}

func (h *TrialsHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func parsePagination(r *http.Request, defaultLimit int) (int, int) {
	limit := defaultLimit
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}
