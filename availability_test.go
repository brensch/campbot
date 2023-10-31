package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"

	"go.uber.org/zap"
)

func TestGenerateNotifications(t *testing.T) {
	// Initialize context and logger
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	// sc := NewSchniffCollection("example_schniffs.json")
	sc := &SchniffCollection{
		schniffs:     make([]*Schniff, 0),
		mutex:        sync.Mutex{},
		fileLocation: "example_schniffs.json",
	}
	sc.load()

	// Read and unmarshal the availabilities.json file
	availabilitiesFile, err := os.ReadFile("availability.json")
	if err != nil {
		t.Fatalf("Error reading availabilities.json file: %v", err)
	}
	var availabilities []AvailabilityWithID
	err = json.Unmarshal(availabilitiesFile, &availabilities)
	if err != nil {
		t.Fatalf("Error unmarshalling availabilities.json: %v", err)
	}

	// Run the GenerateNotifications function
	notifications, records, err := GenerateNotifications(ctx, logger, availabilities, sc, []NotificationRecord{})
	if err != nil {
		t.Fatalf("Error in GenerateNotifications function: %v", err)
	}

	// Add your own assertions based on what you expect `notifications` to be
	// Here's a simple example: check that we have notifications
	if len(notifications) == 0 {
		t.Errorf("Expected notifications, but got none")
	}

	fmt.Println(len(notifications))

	for _, notification := range notifications {
		fmt.Println(len(notification.AvailableCampsites))
		for _, campsite := range notification.AvailableCampsites {
			if campsite.CampsiteID != "71047" {
				continue
			}
			fmt.Println(campsite.CampsiteID, campsite.Date)
		}
	}

	for _, record := range records {
		fmt.Println(record.CampgroundID, record.CampsiteID, record.TargetDate)
	}

	// Add more assertions as needed...
}
