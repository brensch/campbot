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

const (
	EntityTypeCampground     = "campground"
	EntityTypeRecreationArea = "recarea"
)

type GeoSearchResults struct {
	Latitude              string   `json:"latitude"`
	Location              string   `json:"location"`
	Longitude             string   `json:"longitude"`
	Radius                string   `json:"radius"`
	Results               []Entity `json:"results"`
	Size                  int      `json:"size"`
	SpellingAutocorrected bool     `json:"spelling_autocorrected"`
	Start                 string   `json:"start"`
	Total                 int      `json:"total"`
}

type Entity struct {
	AccessibleCampsitesCount int `json:"accessible_campsites_count,omitempty"`
	Activities               []struct {
		ActivityDescription    string `json:"activity_description"`
		ActivityFeeDescription string `json:"activity_fee_description"`
		ActivityID             int    `json:"activity_id"`
		ActivityName           string `json:"activity_name"`
	} `json:"activities"`
	Addresses []struct {
		AddressType    string `json:"address_type"`
		City           string `json:"city"`
		CountryCode    string `json:"country_code"`
		PostalCode     string `json:"postal_code"`
		StateCode      string `json:"state_code"`
		StreetAddress1 string `json:"street_address1"`
		StreetAddress2 string `json:"street_address2"`
		StreetAddress3 string `json:"street_address3"`
	} `json:"addresses"`
	AggregateCellCoverage float64   `json:"aggregate_cell_coverage,omitempty"`
	AverageRating         float64   `json:"average_rating,omitempty"`
	CampsiteAccessible    int       `json:"campsite_accessible,omitempty"`
	CampsiteEquipmentName []string  `json:"campsite_equipment_name,omitempty"`
	CampsiteReserveType   []string  `json:"campsite_reserve_type"`
	CampsiteTypeOfUse     []string  `json:"campsite_type_of_use"`
	CampsitesCount        string    `json:"campsites_count"`
	City                  string    `json:"city"`
	CountryCode           string    `json:"country_code"`
	Description           string    `json:"description"`
	Directions            string    `json:"directions"`
	Distance              string    `json:"distance"`
	EntityID              string    `json:"entity_id"`
	EntityType            string    `json:"entity_type"`
	GoLiveDate            time.Time `json:"go_live_date"`
	HTMLDescription       string    `json:"html_description"`
	ID                    string    `json:"id"`
	Latitude              string    `json:"latitude"`
	Links                 []struct {
		Description string `json:"description"`
		LinkType    string `json:"link_type"`
		Title       string `json:"title"`
		URL         string `json:"url"`
	} `json:"links"`
	Longitude string `json:"longitude"`
	Name      string `json:"name"`
	Notices   []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"notices"`
	NumberOfRatings int    `json:"number_of_ratings,omitempty"`
	OrgID           string `json:"org_id"`
	OrgName         string `json:"org_name"`
	ParentID        string `json:"parent_id"`
	ParentName      string `json:"parent_name"`
	ParentType      string `json:"parent_type"`
	PreviewImageURL string `json:"preview_image_url"`
	PriceRange      struct {
		AmountMax int    `json:"amount_max"`
		AmountMin int    `json:"amount_min"`
		PerUnit   string `json:"per_unit"`
	} `json:"price_range,omitempty"`
	Rate []struct {
		EndDate time.Time `json:"end_date"`
		Prices  []struct {
			Amount    int    `json:"amount"`
			Attribute string `json:"attribute"`
		} `json:"prices"`
		RateMap map[string]struct {
			GroupFees        interface{} `json:"group_fees"`
			SingleAmountFees Fees        `json:"single_amount_fees"`
		} `json:"rate_map"`
		SeasonDescription string    `json:"season_description"`
		SeasonType        string    `json:"season_type"`
		StartDate         time.Time `json:"start_date"`
	} `json:"rate"`
	Reservable bool   `json:"reservable"`
	StateCode  string `json:"state_code"`
	TimeZone   string `json:"time_zone,omitempty"`
	Type       string `json:"type"`
}

type Fees struct {
	Deposit   int `json:"deposit"`
	Holiday   int `json:"holiday"`
	PerNight  int `json:"per_night"`
	PerPerson int `json:"per_person"`
	Weekend   int `json:"weekend"`
}

func ConvertToModelCampground(apiCampground *Entity) models.Campground {
	return models.Campground{
		ID:       apiCampground.EntityID,
		Name:     apiCampground.Name,
		Provider: ProviderRecreationGov,
	}
}

func GetCampgrounds(ctx context.Context, olog *zap.Logger, client *http.Client, baseURI string, latNE, latSW, lngNE, lngSW float64, entityTypes []string) (GeoSearchResults, error) {

	start := time.Now()
	log := olog.With(
		zap.Float64("lat_ne", latNE),
		zap.Float64("lat_sw", latSW),
		zap.Float64("lng_ne", lngNE),
		zap.Float64("lng_sw", lngSW),
		zap.Strings("entity_types", entityTypes),
	)
	log.Debug("doing campground search using api")
	endpoint := fmt.Sprintf("%s/api/search/geo", baseURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Error("couldn't create request", zap.Error(err))
		return GeoSearchResults{}, err
	}

	v := req.URL.Query()

	v.Add("lat_ne", fmt.Sprint(latNE))
	v.Add("lat_sw", fmt.Sprint(latSW))
	v.Add("lng_ne", fmt.Sprint(lngNE))
	v.Add("lng_sw", fmt.Sprint(lngSW))

	// TODO: maybe make these customizable
	v.Add("exact", "false")
	v.Add("size", "1000")
	for _, entityType := range entityTypes {
		v.Add("fq", fmt.Sprintf("entity_type%%3A%s", entityType))
	}

	req.URL.RawQuery = v.Encode()

	res, err := client.Do(req)
	if err != nil {
		log.Error("couldn't do request", zap.Error(err))
		return GeoSearchResults{}, err
	}
	defer res.Body.Close()

	// doing a readall since cloudflare dumps xml on you
	resContents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("couldn't read response", zap.Error(err))
		return GeoSearchResults{}, err
	}

	if res.StatusCode != http.StatusOK {
		log.Warn("got bad statuscode searching geo", zap.Int("status_code", res.StatusCode))
		log.Debug("body of bad request", zap.String("body", string(resContents)))
		return GeoSearchResults{}, fmt.Errorf(string(resContents))
	}

	var results GeoSearchResults
	err = json.Unmarshal(resContents, &results)
	if err != nil {
		log.Error("couldn't unmarshal", zap.Error(err))
		return GeoSearchResults{}, err
	}

	log.Debug("completed campground search using api", zap.Duration("duration", time.Since(start)))

	return results, nil

}
