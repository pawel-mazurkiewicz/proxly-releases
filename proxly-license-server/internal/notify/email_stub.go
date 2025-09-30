package notify

import (
	"fmt"
	"log"

	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
)

type EmailNotifier struct {
	enabled bool
}

func NewEmailNotifier(enabled bool) *EmailNotifier {
	return &EmailNotifier{
		enabled: enabled,
	}
}

func (e *EmailNotifier) NotifyThresholdCrossed(license *types.License) error {
	if !e.enabled {
		return nil
	}

	message := fmt.Sprintf(
		"License %s (ID: %s) has reached its activation threshold. Current count: %d/%d",
		license.Key,
		license.ID,
		license.ActivationCount,
		license.MaxActivations,
	)

	log.Printf("EMAIL NOTIFICATION: %s", message)
	return nil
}

func (e *EmailNotifier) NotifyActivationExceeded(license *types.License) error {
	if !e.enabled {
		return nil
	}

	message := fmt.Sprintf(
		"License %s (ID: %s) has exceeded its activation limit! Current count: %d/%d",
		license.Key,
		license.ID,
		license.ActivationCount,
		license.MaxActivations,
	)

	log.Printf("EMAIL NOTIFICATION: %s", message)
	return nil
}

func (e *EmailNotifier) NotifyLicenseRevoked(license *types.License) error {
	if !e.enabled {
		return nil
	}

	message := fmt.Sprintf(
		"License %s (ID: %s) has been revoked",
		license.Key,
		license.ID,
	)

	log.Printf("EMAIL NOTIFICATION: %s", message)
	return nil
}