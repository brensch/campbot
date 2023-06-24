package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/brensch/campbot/models"
	"go.uber.org/zap"
)

// Convert API Campsite to Model Campsite
func ConvertToModelCampsite(apiCampsite *CampsiteDetailed) models.Campsite {
	return models.Campsite{
		SiteID:   apiCampsite.CampsiteID,
		SiteType: apiCampsite.Type,
		UseType:  apiCampsite.TypeOfUse,
		Loop:     apiCampsite.Loop,
		// TODO: can we get min max users from the api?
	}
}

type CampsiteSearchResults struct {
	Campsites             []CampsiteDetailed `json:"campsites"`
	Size                  int                `json:"size"`
	SpellingAutocorrected bool               `json:"spelling_autocorrected"`
	Start                 string             `json:"start"`
	Total                 int                `json:"total"`
}
type Attribute struct {
	AttributeCategory string `json:"attribute_category"`
	AttributeID       int    `json:"attribute_id"`
	AttributeName     string `json:"attribute_name"`
	AttributeValue    string `json:"attribute_value"`
}
type FeeTemplates struct {
	Lottery    string `json:"Lottery"`
	OffPeak    string `json:"Off Peak"`
	Peak       string `json:"Peak"`
	Transition string `json:"Transition"`
	WalkIn     string `json:"Walk In"`
}

type PermittedEquipment struct {
	CampsiteEquipmentTypeID int       `json:"campsite_equipment_type_id"`
	CreatedDate             time.Time `json:"created_date"`
	EquipmentName           string    `json:"equipment_name"`
	IsDeactivated           bool      `json:"is_deactivated"`
	MaxLength               int       `json:"max_length"`
	UpdatedDate             time.Time `json:"updated_date"`
}

type CampsiteDetailed struct {
	Accessible            string               `json:"accessible"`
	AggregateCellCoverage float64              `json:"aggregate_cell_coverage"`
	AssetID               string               `json:"asset_id"`
	AssetName             string               `json:"asset_name"`
	AssetType             string               `json:"asset_type"`
	Attributes            []Attribute          `json:"attributes"`
	AverageRating         float64              `json:"average_rating"`
	CampsiteID            string               `json:"campsite_id"`
	CampsiteReserveType   string               `json:"campsite_reserve_type"`
	CampsiteStatus        string               `json:"campsite_status"`
	City                  string               `json:"city"`
	CountryCode           string               `json:"country_code"`
	FeeTemplates          FeeTemplates         `json:"fee_templates"`
	Latitude              string               `json:"latitude"`
	Longitude             string               `json:"longitude"`
	Loop                  string               `json:"loop"`
	Name                  string               `json:"name"`
	NumberOfRatings       int                  `json:"number_of_ratings"`
	OrgID                 string               `json:"org_id"`
	OrgName               string               `json:"org_name"`
	ParentAssetID         string               `json:"parent_asset_id"`
	ParentAssetName       string               `json:"parent_asset_name"`
	ParentAssetType       string               `json:"parent_asset_type"`
	PermittedEquipment    []PermittedEquipment `json:"permitted_equipment"`
	PreviewImageURL       string               `json:"preview_image_url"`
	Reservable            bool                 `json:"reservable"`
	StateCode             string               `json:"state_code"`
	Type                  string               `json:"type"`
	TypeOfUse             string               `json:"type_of_use"`
}

func GetCampsites(ctx context.Context, olog *zap.Logger, client *http.Client, baseURI, campgroundID string) (CampsiteSearchResults, error) {

	start := time.Now()
	log := olog.With(
		zap.String("base_uri", baseURI),
		zap.String("campground", campgroundID),
	)
	log.Debug("doing campsite search using api")
	endpoint := fmt.Sprintf("%s/api/search/campsites?start=0&size=1000&fq=asset_id%%3A%s&include_non_site_specific_campsites=true", baseURI, campgroundID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Error("couldn't create request", zap.Error(err))
		return CampsiteSearchResults{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Error("couldn't do request", zap.Error(err))
		return CampsiteSearchResults{}, err
	}
	defer res.Body.Close()

	// doing a readall since cloudflare dumps xml on you
	resContents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("couldn't read response", zap.Error(err))
		return CampsiteSearchResults{}, err
	}

	if res.StatusCode != http.StatusOK {
		log.Warn("got bad statuscode searching campsites", zap.Int("status_code", res.StatusCode))
		log.Debug("body of bad request", zap.String("body", string(resContents)))
		return CampsiteSearchResults{}, fmt.Errorf(string(resContents))
	}

	var results CampsiteSearchResults
	err = json.Unmarshal(resContents, &results)
	if err != nil {
		log.Error("couldn't unmarshal", zap.Error(err))
		return CampsiteSearchResults{}, err
	}

	log.Debug("completed campsite search using api", zap.Duration("duration", time.Since(start)))

	return results, nil
}
