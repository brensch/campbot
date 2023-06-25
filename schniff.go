package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Schniff struct {
	SchniffID string `json:"schniff_id"`
	Active    bool   `json:"active"`

	CampgroundID   string    `json:"campground_id"`
	CampgroundName string    `json:"campground_name"`
	CampsiteIDs    []string  `json:"campsite_ids"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	UserID         string    `json:"user_id"`
	UserNick       string    `json:"user_nick"`
}

type SchniffCollection struct {
	schniffs     []*Schniff
	fileLocation string
	mutex        sync.Mutex
}

func NewSchniffCollection(fileLocation string) *SchniffCollection {
	sc := &SchniffCollection{
		schniffs:     make([]*Schniff, 0),
		mutex:        sync.Mutex{},
		fileLocation: fileLocation,
	}

	sc.load()

	return sc
}

func (sc *SchniffCollection) Add(s *Schniff) error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.schniffs = append(sc.schniffs, s)

	return sc.save()
}

func (sc *SchniffCollection) SetActive(id string, active bool) error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	for _, schniff := range sc.schniffs {
		if schniff.SchniffID != id {
			continue
		}
		fmt.Println("Found schniff", schniff.SchniffID, "setting active to", active)
		schniff.Active = active
		return sc.save()
	}

	return fmt.Errorf("id not found")
}

func (sc *SchniffCollection) GetSchniff(id string) (*Schniff, error) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	for _, schniff := range sc.schniffs {
		if schniff.SchniffID != id {
			continue
		}
		return schniff, nil
	}

	return nil, fmt.Errorf("id not found")
}

func (sc *SchniffCollection) GetSchniffsForUser(userID string) []*Schniff {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	var schniffsForUser []*Schniff
	for _, schniff := range sc.schniffs {
		if schniff.UserID == userID {

			schniffsForUser = append(schniffsForUser, schniff)
		}
	}

	return schniffsForUser
}

func (sc *SchniffCollection) load() error {

	data, err := os.ReadFile(sc.fileLocation)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &sc.schniffs)
	if err != nil {
		return err
	}

	return nil
}

func (sc *SchniffCollection) save() error {

	data, err := json.MarshalIndent(sc.schniffs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sc.fileLocation, data, 0644)
}

func GenerateTableMessage(schniffs []*Schniff) string {
	var builder strings.Builder

	// Table headers
	builder.WriteString("```\n")
	builder.WriteString(fmt.Sprintf("%-40s %-15s %-15s %-15s %-15s %-15s %-15s\n", "ID", "CampgroundID", "CampgroundName", "StartDate", "EndDate", "UserNick", "CampsiteIDs"))

	// Table rows
	for _, schniff := range schniffs {
		builder.WriteString(fmt.Sprintf("%-40s %-15s %-15s %-15s %-15s %-15s %-15s\n", schniff.SchniffID, schniff.CampgroundID, schniff.CampgroundName, schniff.StartDate.Format("2006-01-02"), schniff.EndDate.Format("2006-01-02"), schniff.UserNick, strings.Join(schniff.CampsiteIDs, ",")))
	}

	builder.WriteString("```")

	return builder.String()
}
