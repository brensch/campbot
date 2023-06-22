package providers

import (
	"github.com/brensch/campbot/pkg/models"
)

type ReserveCaliforniaProvider struct {
	// any data or services the provider needs
}

func (p *ReserveCaliforniaProvider) GetAvailability(campsiteID string) (models.Availability, error) {
	// implementation to get availability from the reservecalifornia API
	return models.Availability{}, nil
}
