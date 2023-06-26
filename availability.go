package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/brensch/campbot/stealthing"

	"go.uber.org/zap"
)

type Availability struct {
	Campsites map[string]Campsite `json:"campsites,omitempty"`
	Count     int                 `json:"count,omitempty"`
}

type AvailabilityWithID struct {
	CampgroundID string
	Availability Availability
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

func GetStartOfMonth(input time.Time) time.Time {
	return time.Date(input.Year(), input.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// GetAvailability ensures that the targettime is snapped to the start of the month, then queries the API for all availabilities at that ground
func GetAvailability(ctx context.Context, olog *zap.Logger, client *http.Client, campgroundID string, targetTime time.Time) (AvailabilityWithID, error) {
	start := time.Now()
	log := olog.With(
		zap.String("campground", campgroundID),
		zap.Time("target_time", targetTime),
	)
	log.Debug("getting availability from api")
	endpoint := fmt.Sprintf("https://recreation.gov/api/camps/availability/campground/%s/month", campgroundID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Error("couldn't create request", zap.Error(err))
		return AvailabilityWithID{}, err
	}

	req.Header.Set("User-Agent", stealthing.RandomUserAgent())

	// round the time to the start of the target month and put in param "start_date"
	monthStart := GetStartOfMonth(targetTime)

	// params need to be url encoded. ie base64
	v := req.URL.Query()
	v.Add("start_date", monthStart.Format("2006-01-02T15:04:05.000Z"))
	req.URL.RawQuery = v.Encode()

	res, err := client.Do(req)
	if err != nil {
		log.Error("couldn't do request", zap.Error(err))
		return AvailabilityWithID{}, err
	}
	defer res.Body.Close()

	resContents, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error("couldn't read response", zap.Error(err))
		return AvailabilityWithID{}, err
	}

	if res.StatusCode != http.StatusOK {
		// Leaving this as just a warning so that logs don't count as errors until they fail the retry
		log.Warn("got bad statuscode getting availability", zap.Int("status_code", res.StatusCode))
		log.Debug("body of bad request", zap.String("body", string(resContents)))
		return AvailabilityWithID{}, fmt.Errorf(string(resContents))
	}

	var availability Availability
	err = json.Unmarshal(resContents, &availability)
	if err != nil {
		log.Error("couldn't unmarshal", zap.Error(err))
		return AvailabilityWithID{}, err
	}

	log.Debug("completed getting availability from api", zap.Duration("duration", time.Since(start)))

	return AvailabilityWithID{Availability: availability, CampgroundID: campgroundID}, nil

}

type AvailabilityRequest struct {
	CampgroundID string    `json:"campground_id"`
	TargetTime   time.Time `json:"target_time"` // this should be the start of the month
}

// ConstructAvailabilityRequests takes a list of schniffs and returns a list of availability requests by
// de-duplicating the campgroundIDs and extracting all the time periods from the schniffs
func ConstructAvailabilityRequests(ctx context.Context, olog *zap.Logger, client *http.Client, sc *SchniffCollection) []AvailabilityRequest {
	campgroundTimes := make(map[string][]time.Time)

	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	// Iterate over schniffs to extract the months the schniff ranges over for each campgroundID
	for _, schniff := range sc.schniffs {
		if !schniff.Active {
			continue
		}
		// Generate all months between StartDate and EndDate
		start, end := schniff.StartDate, schniff.EndDate
		for d := start; d.Before(end) || d.Equal(end); d = d.AddDate(0, 1, 0) {
			monthStart := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC) // Start of the month
			campgroundTimes[schniff.CampgroundID] = append(campgroundTimes[schniff.CampgroundID], monthStart)
		}
	}

	availabilityRequests := make([]AvailabilityRequest, 0)

	// Create availability requests for each campgroundID and targetTime
	for campgroundID, times := range campgroundTimes {
		for _, targetTime := range times {
			availabilityRequests = append(availabilityRequests, AvailabilityRequest{
				CampgroundID: campgroundID,
				TargetTime:   targetTime,
			})
		}
	}

	return availabilityRequests
}

func DeduplicateAvailabilityRequests(requests []AvailabilityRequest) []AvailabilityRequest {
	seen := make(map[string]struct{})
	var deduplicated []AvailabilityRequest

	for _, request := range requests {
		key := request.CampgroundID + request.TargetTime.String()
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			deduplicated = append(deduplicated, request)
		}
	}

	return deduplicated
}

// CheckAvailability does a list of requests and returns a list of availabilities
func DoRequests(ctx context.Context, olog *zap.Logger, client *http.Client, requests []AvailabilityRequest) ([]AvailabilityWithID, error) {
	var availabilities []AvailabilityWithID

	for _, request := range requests {
		// Use the current time as the targetTime
		availability, err := GetAvailability(ctx, olog, client, request.CampgroundID, request.TargetTime)
		if err != nil {
			olog.Error("Unable to get availability", zap.Error(err))
			continue
		}

		availabilities = append(availabilities, availability)
	}

	return availabilities, nil
}

// GenerateNotifications takes a list of schniffs and a list of availabilities and generates notifications
func GenerateNotifications(ctx context.Context, olog *zap.Logger, availabilities []AvailabilityWithID, sc *SchniffCollection) ([]Notification, error) {
	var notifications []Notification
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	for _, schniff := range sc.schniffs {
		if !schniff.Active {
			continue
		}
		notification := Notification{SchniffID: schniff.SchniffID}
		// Find the availability for this schniff
		for _, availability := range availabilities {
			// Check if the schniff campgroundID matches the availability campgroundID
			if schniff.CampgroundID != availability.CampgroundID {
				continue
			}

			for campsiteID, campsite := range availability.Availability.Campsites {
				for date, state := range campsite.Availabilities {
					if state != "Available" {
						continue
					}

					date, err := time.Parse(time.RFC3339, date)
					if err != nil {
						olog.Error("Unable to parse date", zap.Error(err))
						continue
					}
					// Check if the date is in the schniff range, start and end date inclusive
					if !(date.After(schniff.StartDate) || date.Equal(schniff.StartDate)) || !(date.Before(schniff.EndDate) || date.Equal(schniff.EndDate)) {
						continue
					}

					notification.AvailableCampsites = append(notification.AvailableCampsites, CampsiteAvailability{
						CampsiteID: campsiteID,
						Date:       date,
					})

				}
			}
		}

		if len(notification.AvailableCampsites) == 0 {
			continue
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}
