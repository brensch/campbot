package providers

import (
	"github.com/brensch/campbot/pkg/models"
)

// Provider interface that all campsite providers must implement
type Provider interface {
	GetAvailability(campsiteID string) (models.Availability, error)
}

// A function to instantiate a provider based on its type
func NewProvider(providerType string) Provider {
	switch providerType {
	case "recreation.gov":
		return &RecreationGovProvider{}
	case "reservecalifornia":
		return &ReserveCaliforniaProvider{}
	default:
		// handle the case when the provider type doesn't match any existing providers
		return nil
	}
}
