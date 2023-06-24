package providers

import (
	"github.com/brensch/campbot/models"
)

type ReserveCaliforniaProvider struct {
	// any data or services the provider needs
}

func (p *ReserveCaliforniaProvider) GetAvailability(campsiteID string) (models.Availability, error) {
	// implementation to get availability from the reservecalifornia API
	return models.Availability{}, nil
}

func (p *ReserveCaliforniaProvider) GetCampsites() ([]models.Campsite, error) {
	// implementation to get availability from the recreation.gov API
	return nil, nil
}
