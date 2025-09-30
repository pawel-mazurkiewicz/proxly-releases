package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/core"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/http/middleware"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/errorsx"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
	"go.uber.org/zap"
)

type ActivationsHandler struct {
	licenseService *core.LicenseService
	logger         *zap.Logger
}

func NewActivationsHandler(licenseService *core.LicenseService, logger *zap.Logger) *ActivationsHandler {
	return &ActivationsHandler{
		licenseService: licenseService,
		logger:         logger,
	}
}

func (h *ActivationsHandler) Activate(w http.ResponseWriter, r *http.Request) {
	var req types.ActivationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, errorsx.ErrInvalidPayload)
		return
	}

	if req.LicenseKey == "" {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "license_key is required"))
		return
	}

	clientIP, ok := r.Context().Value(middleware.ClientIPKey).(string)
	if !ok {
		clientIP = "unknown"
	}

	userAgent, ok := r.Context().Value(middleware.UserAgentKey).(string)
	if !ok {
		userAgent = r.UserAgent()
	}

	if req.UserAgent == "" {
		req.UserAgent = userAgent
	}

	response, err := h.licenseService.ProcessActivation(&req, clientIP)
	if err != nil {
		if customErr, ok := err.(*errorsx.Error); ok {
			h.writeErrorResponse(w, customErr)
		} else {
			h.logger.Error("Failed to process activation", zap.Error(err))
			h.writeErrorResponse(w, errorsx.ErrInternalServer)
		}
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *ActivationsHandler) writeErrorResponse(w http.ResponseWriter, err *errorsx.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func (h *ActivationsHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}