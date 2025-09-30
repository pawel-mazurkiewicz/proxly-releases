package core

import (
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
)

type CompositeNotifier struct {
	notifiers []Notifier
}

func NewCompositeNotifier(notifiers ...Notifier) *CompositeNotifier {
	return &CompositeNotifier{
		notifiers: notifiers,
	}
}

func (c *CompositeNotifier) NotifyThresholdCrossed(license *types.License) error {
	for _, notifier := range c.notifiers {
		if err := notifier.NotifyThresholdCrossed(license); err != nil {
			// Log error but continue with other notifiers
			// In production, you might want more sophisticated error handling
			continue
		}
	}
	return nil
}

func (c *CompositeNotifier) NotifyActivationExceeded(license *types.License) error {
	for _, notifier := range c.notifiers {
		if err := notifier.NotifyActivationExceeded(license); err != nil {
			continue
		}
	}
	return nil
}

func (c *CompositeNotifier) NotifyLicenseRevoked(license *types.License) error {
	for _, notifier := range c.notifiers {
		if err := notifier.NotifyLicenseRevoked(license); err != nil {
			continue
		}
	}
	return nil
}