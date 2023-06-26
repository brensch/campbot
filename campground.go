package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/brensch/campbot/stealthing"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type CampgroundCollection struct {
	mu          sync.Mutex
	Campgrounds []SummarisedCampground
}

type CampgroundSearchResults struct {
	Results               []Campground `json:"results"`
	Size                  int          `json:"size"`
	SpellingAutocorrected bool         `json:"spelling_autocorrected"`
	Start                 string       `json:"start"`
	Total                 int          `json:"total"`
}
type Activity struct {
	ActivityDescription    string `json:"activity_description"`
	ActivityFeeDescription string `json:"activity_fee_description"`
	ActivityID             int    `json:"activity_id"`
	ActivityName           string `json:"activity_name"`
}
type Address struct {
	AddressType    string `json:"address_type"`
	City           string `json:"city"`
	CountryCode    string `json:"country_code"`
	PostalCode     string `json:"postal_code"`
	StateCode      string `json:"state_code"`
	StreetAddress1 string `json:"street_address1"`
	StreetAddress2 string `json:"street_address2"`
	StreetAddress3 string `json:"street_address3"`
}
type Links struct {
	Description string `json:"description"`
	LinkType    string `json:"link_type"`
	Title       string `json:"title"`
	URL         string `json:"url"`
}
type Notices struct {
	Text string `json:"text"`
	Type string `json:"type"`
}
type PriceRange struct {
	AmountMax float64 `json:"amount_max"`
	AmountMin float64 `json:"amount_min"`
	PerUnit   string  `json:"per_unit"`
}

type Campground struct {
	Activities               []Activity `json:"activities,omitempty"`
	Addresses                []Address  `json:"addresses,omitempty"`
	AggregateCellCoverage    float64    `json:"aggregate_cell_coverage,omitempty"`
	AverageRating            float64    `json:"average_rating,omitempty"`
	CampsiteEquipmentName    []string   `json:"campsite_equipment_name,omitempty"`
	CampsiteReserveType      []string   `json:"campsite_reserve_type,omitempty"`
	CampsiteTypeOfUse        []string   `json:"campsite_type_of_use"`
	CampsitesCount           string     `json:"campsites_count,omitempty"`
	City                     string     `json:"city,omitempty"`
	CountryCode              string     `json:"country_code,omitempty"`
	Description              string     `json:"description"`
	Directions               string     `json:"directions,omitempty"`
	EntityID                 string     `json:"entity_id"`
	EntityType               string     `json:"entity_type"`
	GoLiveDate               time.Time  `json:"go_live_date"`
	HTMLDescription          string     `json:"html_description"`
	ID                       string     `json:"id"`
	Latitude                 string     `json:"latitude,omitempty"`
	Links                    []Links    `json:"links,omitempty"`
	Longitude                string     `json:"longitude,omitempty"`
	Name                     string     `json:"name"`
	Notices                  []Notices  `json:"notices"`
	NumberOfRatings          int        `json:"number_of_ratings,omitempty"`
	OrgID                    string     `json:"org_id"`
	OrgName                  string     `json:"org_name"`
	ParentID                 string     `json:"parent_id"`
	ParentName               string     `json:"parent_name"`
	ParentType               string     `json:"parent_type"`
	PreviewImageURL          string     `json:"preview_image_url,omitempty"`
	PriceRange               PriceRange `json:"price_range,omitempty"`
	Reservable               bool       `json:"reservable"`
	StateCode                string     `json:"state_code"`
	Type                     string     `json:"type"`
	TimeZone                 string     `json:"time_zone,omitempty"`
	OfficialSiteURL          string     `json:"official_site_url,omitempty"`
	AccessibleCampsitesCount int        `json:"accessible_campsites_count,omitempty"`
	CampsiteAccessible       int        `json:"campsite_accessible,omitempty"`
}

type SummarisedCampground struct {
	Name       string
	ParentName string
	ID         string
	Source     string
	Rating     float64
}

// GetCampgrounds retrieves all campgrounds. It looks like they forgot to actually apply the limit that you
// specify, meaning we can get the entire database in one call. Should only do this once every week or so to
// be kind.
func CampgroundRequest(ctx context.Context, log *zap.Logger, client *http.Client, start int) (CampgroundSearchResults, error) {
	log.Debug("getting campground")
	endpoint := fmt.Sprintf("https://recreation.gov/api/search?fq=entity_type%%3Acampground&size=100&start=%d", start)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Error("couldn't create request", zap.Error(err))
		return CampgroundSearchResults{}, err
	}
	req.Header.Set("User-Agent", stealthing.RandomUserAgent())

	res, err := client.Do(req)
	if err != nil {
		log.Error("couldn't do request", zap.Error(err))
		return CampgroundSearchResults{}, err
	}
	defer res.Body.Close()

	var target CampgroundSearchResults
	err = json.NewDecoder(res.Body).Decode(&target)
	if err != nil {
		log.Error("couldn't decode campground response", zap.Error(err))
		return CampgroundSearchResults{}, err
	}

	return target, nil
}

func SummariseCampground(apiCampground Campground) SummarisedCampground {
	campground := SummarisedCampground{
		Source: "recreation.gov",

		ID:         apiCampground.EntityID,
		Name:       cases.Title(language.Und).String(apiCampground.Name),
		ParentName: apiCampground.ParentName,
		Rating:     apiCampground.AverageRating,
	}

	return campground
}

// GetCampgrounds iterates through the search api until all campgrounds are retrieved
func GetCampgrounds(ctx context.Context, log *zap.Logger, client *http.Client) ([]SummarisedCampground, error) {

	page := 0
	startingSize := 100
	size := startingSize
	var campgrounds []SummarisedCampground

	// 100 indicates that we received a partial page, and are therefore at the end
	for size == startingSize {
		log.Debug("getting campground page", zap.Int("page", page))
		apiCampgrounds, err := CampgroundRequest(ctx, log.With(), client, page)
		if err != nil {
			log.Error("got error getting campgrounds", zap.Error(err))
			return nil, err
		}
		for _, apiCampground := range apiCampgrounds.Results {
			campgrounds = append(campgrounds, SummariseCampground(apiCampground))
		}

		size = apiCampgrounds.Size
		page = page + size
	}

	return campgrounds, nil
}

func NewCampgroundCollection(ctx context.Context, log *zap.Logger, s *discordgo.Session, client *http.Client) (*CampgroundCollection, error) {
	// get campgrounds if campgrounds.json doesn't exist
	cc := &CampgroundCollection{
		mu: sync.Mutex{},
	}

	_, err := os.Stat("campgrounds.json")
	if os.IsNotExist(err) {
		// update campgrounds
		err = cc.UpdateCampgrounds(ctx, log, s, client)
		if err != nil {
			log.Error("couldn't update campgrounds", zap.Error(err))
			return nil, err
		}
		return cc, nil
	}

	// if we do have the file, read it and populate the campgrounds and discord options
	campgroundsJSON, err := os.ReadFile("campgrounds.json")
	if err != nil {
		log.Error("couldn't read campgrounds.json", zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(campgroundsJSON, &cc.Campgrounds)
	if err != nil {
		log.Error("couldn't unmarshal campgrounds", zap.Error(err))
		return nil, err
	}

	return cc, nil
}

// UpdateCampgrounds updates the campground colleciton with the latest campgrounds
func (cc *CampgroundCollection) UpdateCampgrounds(ctx context.Context, log *zap.Logger, s *discordgo.Session, client *http.Client) error {
	// get campgrounds
	campgrounds, err := GetCampgrounds(ctx, log, s.Client)
	if err != nil {
		log.Error("Cannot open the session", zap.Error(err))
		return err
	}

	campgroundsJSON, err := json.Marshal(campgrounds)
	if err != nil {
		log.Error("cannot marshal campgrounds", zap.Error(err))
		return err
	}
	err = os.WriteFile("campgrounds.json", campgroundsJSON, 0644)
	if err != nil {
		log.Error("Cannot write campgrounds to disk", zap.Error(err))
		return err
	}

	cc.mu.Lock()
	cc.Campgrounds = campgrounds
	cc.mu.Unlock()

	return nil
}

func (cc *CampgroundCollection) GetCampgrounds() []SummarisedCampground {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Make a copy of the DiscordOptions slice
	optionsCopy := make([]SummarisedCampground, len(cc.Campgrounds))
	copy(optionsCopy, cc.Campgrounds)

	return optionsCopy
}

func (cc *CampgroundCollection) GetCampground(id string) (SummarisedCampground, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	for _, campground := range cc.Campgrounds {
		if campground.ID == id {
			return campground, nil
		}
	}

	return SummarisedCampground{}, fmt.Errorf("campground not found")
}
