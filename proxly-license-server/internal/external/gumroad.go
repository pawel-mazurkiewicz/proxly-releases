package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GumroadClient struct {
	baseURL   string
	productID string
	client    *http.Client
}

type GumroadVerifyRequest struct {
	ProductID          string `json:"product_id"`
	LicenseKey         string `json:"license_key"`
	IncrementUsesCount string `json:"increment_uses_count"`
}

type GumroadVerifyResponse struct {
	Success  bool                `json:"success"`
	Uses     int                 `json:"uses"`
	Message  string              `json:"message"`
	Purchase GumroadPurchaseInfo `json:"purchase"`
}

type GumroadBool bool

func (b *GumroadBool) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*b = false
		return nil
	}

	if trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' {
		trimmed = strings.Trim(trimmed, "\"")
	}

	boolVal, err := strconv.ParseBool(strings.ToLower(trimmed))
	if err != nil {
		return fmt.Errorf("gumroad bool: %w", err)
	}

	*b = GumroadBool(boolVal)
	return nil
}

type GumroadPurchaseInfo struct {
	ID                string      `json:"id"`
	Email             string      `json:"email"`
	Chargedback       GumroadBool `json:"chargedback"`
	Disputed          GumroadBool `json:"disputed"`
	Refunded          GumroadBool `json:"refunded"`
	PartiallyRefunded GumroadBool `json:"partially_refunded"`
}

func NewGumroadClient(productID string) *GumroadClient {
	return &GumroadClient{
		baseURL:   "https://api.gumroad.com/v2",
		productID: productID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (g *GumroadClient) VerifyLicense(licenseKey string, incrementUses bool) (*GumroadVerifyResponse, error) {
	verifyURL := fmt.Sprintf("%s/licenses/verify", g.baseURL)

	incrementUsesStr := "false"
	if incrementUses {
		incrementUsesStr = "true"
	}

	// Gumroad expects form data
	data := url.Values{}
	data.Set("product_id", g.productID)
	data.Set("license_key", licenseKey)
	data.Set("increment_uses_count", incrementUsesStr)

	req, err := http.NewRequest("POST", verifyURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Proxly-License-Server/1.0")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var verifyResp GumroadVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &verifyResp, nil
}

func (g *GumroadClient) CheckLicenseValidity(licenseKey string) (bool, string, error) {
	resp, err := g.VerifyLicense(licenseKey, false)
	if err != nil {
		return false, "", err
	}

	if !resp.Success {
		reason := resp.Message
		if reason == "" {
			reason = "Invalid license key"
		}
		return false, reason, nil
	}

	// Additional checks can be added here:
	// - Check if refunded/chargedback
	// - Check usage limits
	// - Check purchase date, etc.

	if bool(resp.Purchase.Refunded) {
		return false, "License has been refunded", nil
	}

	if bool(resp.Purchase.Chargedback) {
		return false, "License has been charged back", nil
	}

	if bool(resp.Purchase.Disputed) {
		return false, "License is currently disputed", nil
	}

	if bool(resp.Purchase.PartiallyRefunded) {
		return false, "License has been partially refunded", nil
	}

	return true, "", nil
}
