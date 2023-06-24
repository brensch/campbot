package providers

import (
	"context"
	"net/http"
	"time"

	"github.com/brensch/campbot/models"
	"go.uber.org/zap"
)

const (
	ProviderRecreationGov = "recreation.gov"
)

type RecreationGovProvider struct {
	BaseURI string
	Client  *http.Client
	Logger  *zap.Logger
}

// Ensure RecreationGovProvider implements Provider interface
var _ Provider = (*RecreationGovProvider)(nil)

// NewRecreationGovProvider returns a new RecreationGov provider
func NewRecreationGovProvider(baseURI string, logger *zap.Logger) *RecreationGovProvider {
	return &RecreationGovProvider{
		BaseURI: baseURI,
		Client:  &http.Client{}, // or pass an http.Client as a parameter if you want more control over its configuration
		Logger:  logger,
	}
}

func (p *RecreationGovProvider) GetAvailability(ctx context.Context, targetTime time.Time, campsiteID string) (models.Availability, error) {
	availabilities, err := GetAvailability(ctx, p.Logger, p.Client, p.BaseURI, campsiteID, targetTime)
	if err != nil {
		return models.Availability{}, err
	}

	return ConvertToModelAvailability(&availabilities)[0], nil
}

func (p *RecreationGovProvider) GetCampsites(ctx context.Context, campgroundID string) ([]models.Campsite, error) {

	campsiteSearchResults, err := GetCampsites(ctx, p.Logger, p.Client, p.BaseURI, campgroundID)
	if err != nil {
		return nil, err
	}

	var modelCampsites []models.Campsite
	for _, apiCampsite := range campsiteSearchResults.Campsites {
		modelCampsite := ConvertToModelCampsite(&apiCampsite)
		modelCampsites = append(modelCampsites, modelCampsite)
	}

	return modelCampsites, nil

}

func (p *RecreationGovProvider) GetCampgrounds(ctx context.Context) ([]models.Campground, error) {

	geoSearchResults, err := GetCampgrounds(ctx, p.Logger, p.Client, p.BaseURI, 38.575764, 37.575764, -119.665388, -120.665388, []string{EntityTypeCampground, EntityTypeRecreationArea})
	if err != nil {
		return nil, err
	}

	var modelCampgrounds []models.Campground
	for _, apiCampground := range geoSearchResults.Results {
		modelCampground := ConvertToModelCampground(&apiCampground)
		modelCampgrounds = append(modelCampgrounds, modelCampground)
	}

	return modelCampgrounds, nil
}
