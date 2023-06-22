package providers

import (
	"github.com/brensch/campbot/pkg/models"
)

type RecreationGovProvider struct {
	// any data or services the provider needs
}

func (p *RecreationGovProvider) GetAvailability(campsiteID string) (models.Availability, error) {
	// implementation to get availability from the recreation.gov API
	return models.Availability{}, nil
}
