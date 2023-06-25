package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/brensch/campbot/models"
	"go.uber.org/zap"
)

func TestConvertToModelAvailability(t *testing.T) {
	apiAvailability := &Availability{
		Campsites: map[string]Campsite{
			"test_campsite": {
				CampsiteID:     "test_campsite_id",
				CampsiteType:   "test_campsite_type",
				MaxNumPeople:   5,
				MinNumPeople:   1,
				TypeOfUse:      "test_type_of_use",
				Loop:           "test_loop",
				Availabilities: map[string]string{"2023-06-22": string(StateAvailable), "2023-06-23": string(StateReserved)},
			},
		},
	}

	expected := []models.Availability{
		{
			Date:     time.Date(2023, 06, 22, 0, 0, 0, 0, time.UTC),
			Reserved: false,
			SiteID:   "test_campsite_id",
		},
		{
			Date:     time.Date(2023, 06, 23, 0, 0, 0, 0, time.UTC),
			Reserved: true,
			SiteID:   "test_campsite_id",
		},
	}

	actual := ConvertToModelAvailability(apiAvailability)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %+v but got %+v", expected, actual)
	}
}

func TestGetAvailability(t *testing.T) {
	payload, err := os.ReadFile("./testdata/availability.json")
	if err != nil {
		t.Error(err)
		return
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/camps/availability/campground/test_campground/month" {
			t.Errorf("Expected path to be '/api/camps/availability/campground/test_campground/month' but got %v", r.URL.Path)
		}

		if r.URL.Query().Get("start_date") != "2023-01-01T00:00:00.000Z" {
			t.Errorf("Expected start_date to be '2023-01-01T00:00:00.000Z' but got %v", r.URL.Query().Get("start_date"))
		}

		w.Write(payload)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := &http.Client{}

	availability, err := GetAvailability(context.Background(), logger, client, server.URL, "test_campground", time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Errorf("Expected no error but got %v", err)
	}

	t.Log(len(availability.Campsites))

	if availability.Count == 0 {
		t.Error("should get non 0 count")
	}

	if availability.Campsites["980"].CampsiteID != "980" {
		t.Error("should get campsite id 980")
	}
}
