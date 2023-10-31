package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	SchniffDir = "schniffs"
)

type Schniff struct {
	SchniffID    string    `json:"schniff_id"`
	Active       bool      `json:"active"`
	CreationTime time.Time `json:"creation_time"`

	CampgroundID           string    `json:"campground_id"`
	CampgroundName         string    `json:"campground_name"`
	CampsiteIDs            []string  `json:"campsite_ids"`
	StartDate              time.Time `json:"start_date"`
	EndDate                time.Time `json:"end_date"`
	UserID                 string    `json:"user_id"`
	UserNick               string    `json:"user_nick"`
	MinimumConsecutiveDays int64     `json:"minimum_consecutive_days"`
}

type SchniffCollection struct {
	schniffs     []*Schniff
	mutex        sync.Mutex
	fileLocation string
}

func NewSchniffCollection(fileLocation string) *SchniffCollection {

	location := filepath.Join(SchniffDir, fileLocation)
	sc := &SchniffCollection{
		schniffs:     make([]*Schniff, 0),
		mutex:        sync.Mutex{},
		fileLocation: location,
	}

	// Check if file exists
	if _, err := os.Stat(location); os.IsNotExist(err) {
		// Create the file if it does not exist
		err := os.Mkdir(SchniffDir, 0755)
		if err != nil {
			panic(err)
		}

		_, err = os.Create(location)
		if err != nil {
			panic(err)
		}
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

func GenerateEmbedMessage(schniffs []*Schniff) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:  "Campground Details",
		Color:  0x00ff00, // Green color
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, schniff := range schniffs {
		campgroundURL := fmt.Sprintf("https://www.recreation.gov/camping/campgrounds/%s", schniff.CampgroundID)

		fieldName := schniff.CampgroundName
		fieldValue := fmt.Sprintf(
			"[Link to Campground](%s)\nStartDate: %s\nEndDate: %s\nUserNick: %s\nCampsiteIDs: %s\nActive: %t",
			campgroundURL,
			schniff.StartDate.Format("2006-01-02"),
			schniff.EndDate.Format("2006-01-02"),
			schniff.UserNick,
			strings.Join(schniff.CampsiteIDs, ","),
			schniff.Active,
		)

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fieldName,
			Value:  fieldValue,
			Inline: false,
		})
	}

	return embed
}
