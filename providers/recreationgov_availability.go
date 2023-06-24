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

type State string

var (
	StateAvailable               State = "Available"
	StateReserved                State = "Reserved"
	StateNotReservableManagement State = "Not Reservable Management"
)

type Availability struct {
	Campsites map[string]Campsite `json:"campsites,omitempty"`
	Count     int                 `json:"count,omitempty"`
}

type Campsite struct {
	// keeping this as a string even though it's a time object for less processing
	Availabilities map[string]string `json:"availabilities"`

	CampsiteID          string `json:"campsite_id"`
	CampsiteReserveType string `json:"campsite_reserve_type"`
	CampsiteType        string `json:"campsite_type"`
	CapacityRating      string `json:"capacity_rating"`
	Loop                string `json:"loop"`
	MaxNumPeople        int    `json:"max_num_people"`
	MinNumPeople        int    `json:"min_num_people"`
	Site                string `json:"site"` // not sure what this represents
	TypeOfUse           string `json:"type_of_use"`

	// TODO: find example of this. haven't seen what form it takes yet.
	CampsiteRules interface{} `json:"campsite_rules"`

	// not sure what quantities means
	// TODO: figure out if we need it
	Quantities struct{} `json:"quantities"`
}

// Convert API Availability to Model Availability
func ConvertToModelAvailability(apiAvailability *Availability) []models.Availability {
	var modelAvailabilities []models.Availability

	for _, apiCampsite := range apiAvailability.Campsites {

		// Assume that the apiCampsite.Availabilities map has keys as date strings and values as availability status
		for dateStr, status := range apiCampsite.Availabilities {
			date, err := time.Parse("2006-01-02", dateStr) // adjust date format to whatever the API uses
			if err != nil {
				// log error and continue to next date
				continue
			}

			modelAvailability := models.Availability{
				Date:     date,
				Reserved: status != string(StateAvailable), // assume that anything other than "Available" means reserved
				SiteID:   apiCampsite.CampsiteID,
			}

			modelAvailabilities = append(modelAvailabilities, modelAvailability)
		}
	}

	return modelAvailabilities
}

// GetAvailability ensures that the targettime is snapped to the start of the month, then queries the API for all availabilities at that ground
func GetAvailability(ctx context.Context, olog *zap.Logger, client *http.Client, baseURI string, campgroundID string, targetTime time.Time) (Availability, error) {
	start := time.Now()
	log := olog.With(
		zap.String("base_uri", baseURI),
		zap.String("campground", campgroundID),
		zap.Time("target_time", targetTime),
	)
	log.Debug("getting availability from api")
	endpoint := fmt.Sprintf("%s/api/camps/availability/campground/%s/month", baseURI, campgroundID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Error("couldn't create request", zap.Error(err))
		return Availability{}, err
	}

	// round the time to the start of the target month and put in param "start_date"
	monthStart := GetStartOfMonth(targetTime)

	// params need to be url encoded. ie base64
	v := req.URL.Query()
	v.Add("start_date", monthStart.Format("2006-01-02T15:04:05.000Z"))
	req.URL.RawQuery = v.Encode()

	res, err := client.Do(req)
	if err != nil {
		log.Error("couldn't do request", zap.Error(err))
		return Availability{}, err
	}
	defer res.Body.Close()

	resContents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("couldn't read response", zap.Error(err))
		return Availability{}, err
	}

	if res.StatusCode != http.StatusOK {
		// Leaving this as just a warning so that logs don't count as errors until they fail the retry
		log.Warn("got bad statuscode getting availability", zap.Int("status_code", res.StatusCode))
		log.Debug("body of bad request", zap.String("body", string(resContents)))
		return Availability{}, fmt.Errorf(string(resContents))
	}

	var availability Availability
	err = json.Unmarshal(resContents, &availability)
	if err != nil {
		log.Error("couldn't unmarshal", zap.Error(err))
		return Availability{}, err
	}

	log.Debug("completed getting availability from api", zap.Duration("duration", time.Since(start)))

	return availability, nil

}
