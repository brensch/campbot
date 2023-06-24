package providers

import (
	"context"
	"time"

	"github.com/brensch/campbot/models"
)

// Provider interface that all campsite providers must implement
type Provider interface {
	GetAvailability(ctx context.Context, targetTime time.Time, campsiteID string) (models.Availability, error)
	GetCampsites(ctx context.Context, campgroundID string) ([]models.Campsite, error)
	GetCampgrounds(ctx context.Context) ([]models.Campground, error)
}

// A function to instantiate a provider based on its type
func NewProvider(providerType string) Provider {
	switch providerType {
	case "recreation.gov":
		return &RecreationGovProvider{}
	// case "reservecalifornia":
	// 	return &ReserveCaliforniaProvider{}
	default:
		// handle the case when the provider type doesn't match any existing providers
		return nil
	}
}
