package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
)

type WebhookEvent struct {
	Event     string      `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type WebhookNotifier struct {
	client     *http.Client
	webhookRepo WebhookRepository
}

type WebhookRepository interface {
	GetAll() ([]*types.Webhook, error)
}

func NewWebhookNotifier(webhookRepo WebhookRepository, timeoutMS int) *WebhookNotifier {
	return &WebhookNotifier{
		client: &http.Client{
			Timeout: time.Duration(timeoutMS) * time.Millisecond,
		},
		webhookRepo: webhookRepo,
	}
}

func (w *WebhookNotifier) NotifyThresholdCrossed(license *types.License) error {
	event := &WebhookEvent{
		Event:     "threshold.crossed",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"license_id":        license.ID,
			"key":              license.Key,
			"activation_count":  license.ActivationCount,
			"max_activations":   license.MaxActivations,
		},
	}
	return w.sendWebhooks(event)
}

func (w *WebhookNotifier) NotifyActivationExceeded(license *types.License) error {
	event := &WebhookEvent{
		Event:     "activation.exceeded",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"license_id":        license.ID,
			"key":              license.Key,
			"activation_count":  license.ActivationCount,
			"max_activations":   license.MaxActivations,
		},
	}
	return w.sendWebhooks(event)
}

func (w *WebhookNotifier) NotifyLicenseRevoked(license *types.License) error {
	event := &WebhookEvent{
		Event:     "license.revoked",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"license_id": license.ID,
			"key":        license.Key,
			"status":     license.Status,
		},
	}
	return w.sendWebhooks(event)
}

func (w *WebhookNotifier) sendWebhooks(event *WebhookEvent) error {
	webhooks, err := w.webhookRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	for _, webhook := range webhooks {
		if w.shouldSendEvent(webhook, event.Event) {
			go w.sendWebhook(webhook, event)
		}
	}

	return nil
}

func (w *WebhookNotifier) shouldSendEvent(webhook *types.Webhook, eventType string) bool {
	for _, event := range webhook.Events {
		if event == eventType || event == "*" {
			return true
		}
	}
	return false
}

func (w *WebhookNotifier) sendWebhook(webhook *types.Webhook, event *WebhookEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Failed to marshal webhook payload: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Failed to create webhook request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Proxly-License-Server/1.0")

	if webhook.Secret != nil && *webhook.Secret != "" {
		signature := w.generateSignature(payload, *webhook.Secret)
		req.Header.Set("X-Proxly-Signature", signature)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send webhook to %s: %v\n", webhook.URL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Webhook to %s failed with status %d: %s\n", webhook.URL, resp.StatusCode, string(body))
		return
	}

	fmt.Printf("Webhook sent successfully to %s\n", webhook.URL)
}

func (w *WebhookNotifier) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}