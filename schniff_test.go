package main

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestSchniffCollection(t *testing.T) {
	// Arrange
	tempFile, err := os.CreateTemp("", "schniffs")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	expectedSchniff := &Schniff{
		CampgroundID: "camp1",
		CampsiteIDs:  []string{"site1", "site2"},
		StartDate:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		UserID:       "user1",
	}

	// Act
	sc := NewSchniffCollection(tempFile.Name())
	err = sc.Add(expectedSchniff)
	if err != nil {
		t.Fatalf("Failed to add schniff: %v", err)
	}

	// Assert
	// Verify the schniff was added correctly
	if len(sc.schniffs) != 1 {
		t.Errorf("Expected 1 schniff, got %d", len(sc.schniffs))
	}

	if diff := cmp.Diff(expectedSchniff, sc.schniffs[0]); diff != "" {
		t.Errorf("Schniff mismatch (-want +got):\n%s", diff)
	}

	// Act
	// Save and load the schniffs
	err = sc.save()
	if err != nil {
		t.Fatalf("Failed to save schniffs: %v", err)
	}

	err = sc.load()
	if err != nil {
		t.Fatalf("Failed to load schniffs: %v", err)
	}

	// Assert
	// Verify the schniff was loaded correctly
	if len(sc.schniffs) != 1 {
		t.Errorf("Expected 1 schniff, got %d", len(sc.schniffs))
	}

	if diff := cmp.Diff(expectedSchniff, sc.schniffs[0]); diff != "" {
		t.Errorf("Schniff mismatch (-want +got):\n%s", diff)
	}
}
