package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

type Schniff struct {
	CampgroundID   string    `json:"campground_id"`
	CampgroundName string    `json:"campground_name"`
	CampsiteIDs    []string  `json:"campsite_ids"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	UserID         string    `json:"user_id"`
	UserNick       string    `json:"user_nick"`
}

type SchniffCollection struct {
	schniffs     []Schniff
	fileLocation string
	mutex        sync.Mutex
}

func NewSchniffCollection(fileLocation string) *SchniffCollection {
	sc := &SchniffCollection{
		schniffs:     make([]Schniff, 0),
		mutex:        sync.Mutex{},
		fileLocation: fileLocation,
	}

	sc.load()

	return sc
}

func (sc *SchniffCollection) Add(s Schniff) error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.schniffs = append(sc.schniffs, s)

	return sc.save()
}

func (sc *SchniffCollection) GetSchniffsForUser(userID string) []Schniff {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	var schniffsForUser []Schniff
	for _, schniff := range sc.schniffs {
		if schniff.UserID == userID {

			schniffsForUser = append(schniffsForUser, schniff)
		}
	}

	return schniffsForUser
}

func (sc *SchniffCollection) load() error {

	data, err := ioutil.ReadFile(sc.fileLocation)
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

	data, err := json.Marshal(sc.schniffs)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(sc.fileLocation, data, 0644)
}

func GenerateTableMessage(schniffs []Schniff) string {
	var builder strings.Builder

	// Table headers
	builder.WriteString("```\n")
	builder.WriteString(fmt.Sprintf("%-15s %-15s %-15s %-15s %-15s\n", "CampgroundID", "CampsiteIDs", "StartDate", "EndDate", "UserNick"))

	// Table rows
	for _, schniff := range schniffs {
		builder.WriteString(fmt.Sprintf("%-15s %-15s %-15s %-15s %-15s\n", schniff.CampgroundID, strings.Join(schniff.CampsiteIDs, ","), schniff.StartDate.Format("2006-01-02"), schniff.EndDate.Format("2006-01-02"), schniff.UserNick))
	}

	builder.WriteString("```")

	return builder.String()
}
