package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/pawel-mazurkiewicz/proxly-license-server/internal/store"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/errorsx"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
	"go.uber.org/zap"
)

type AdminHandler struct {
	webhookRepo *store.WebhooksRepository
	logger      *zap.Logger
}

func NewAdminHandler(webhookRepo *store.WebhooksRepository, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		webhookRepo: webhookRepo,
		logger:      logger,
	}
}

type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Secret string   `json:"secret,omitempty"`
	Events []string `json:"events"`
}

func (h *AdminHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, errorsx.ErrInvalidPayload)
		return
	}

	if req.URL == "" {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "url is required"))
		return
	}

	if len(req.Events) == 0 {
		h.writeErrorResponse(w, errorsx.NewError(http.StatusBadRequest, "events are required"))
		return
	}

	webhook := &types.Webhook{
		ID:     uuid.New(),
		URL:    req.URL,
		Events: req.Events,
	}

	if req.Secret != "" {
		webhook.Secret = &req.Secret
	}

	if err := h.webhookRepo.Create(webhook); err != nil {
		h.logger.Error("Failed to create webhook", zap.Error(err))
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, webhook)
}

func (h *AdminHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	webhooks, err := h.webhookRepo.GetAll()
	if err != nil {
		h.logger.Error("Failed to list webhooks", zap.Error(err))
		h.writeErrorResponse(w, errorsx.ErrInternalServer)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"webhooks": webhooks,
	})
}

func (h *AdminHandler) writeErrorResponse(w http.ResponseWriter, err *errorsx.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func (h *AdminHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
