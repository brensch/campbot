package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Notification struct {
	AvailableCampsites []CampsiteAvailability
	SchniffID          string
}

type CampsiteAvailability struct {
	CampsiteID string
	Date       time.Time
}

type campsiteWithDays struct {
	campsiteID string
	daysCount  int
}

func GenerateDiscordMessage(sc *SchniffCollection, notification Notification) (string, error) {
	var message string
	baseURL := "https://www.recreation.gov/camping/campsites/"

	schniff, err := sc.GetSchniff(notification.SchniffID)
	if err != nil {
		return "", err
	}

	// Create an opening sentence with more details about the campground and the date range
	message = fmt.Sprintf("Hello %s, here is the availability of campsites for campground %s (ID: %s) from %s to %s:\n",
		schniff.UserNick,
		schniff.CampgroundName,
		schniff.CampgroundID,
		schniff.StartDate.Format("2006-01-02"),
		schniff.EndDate.Format("2006-01-02"),
	)

	message += "```\nCampsite  Coverage  Link\n"

	// Calculate total number of days in the date range
	totalDays := int(schniff.EndDate.Sub(schniff.StartDate).Hours() / 24)

	// Create a map to hold campsite IDs and counts
	campsiteDayCount := make(map[string]int)

	// Populate the map with the count of days for each campsite
	for _, campsite := range notification.AvailableCampsites {
		campsiteDayCount[campsite.CampsiteID]++
	}

	// Convert map to a slice
	campsites := make([]campsiteWithDays, 0, len(campsiteDayCount))
	for id, count := range campsiteDayCount {
		campsites = append(campsites, campsiteWithDays{campsiteID: id, daysCount: count})
	}

	// Sort the slice in descending order by daysCount
	sort.Slice(campsites, func(i, j int) bool {
		return campsites[i].daysCount > campsites[j].daysCount
	})

	// Take top 10 campsites
	remainingSites := 0
	if len(campsites) > 10 {
		remainingSites = len(campsites) - 10
		campsites = campsites[:10]
	}

	// Add sorted campsites to the message, with available days as percentage
	for _, campsite := range campsites {
		campsiteLink := baseURL + campsite.campsiteID
		availabilityPercentage := float64(campsite.daysCount) / float64(totalDays) * 100
		message += fmt.Sprintf("%-10s  %.2f%%  %s\n", campsite.campsiteID, availabilityPercentage, campsiteLink)
	}

	if remainingSites > 0 {
		message += fmt.Sprintf("... and %d more campsites with available days.", remainingSites)
	}

	message += "```"
	return message, nil
}

func GenerateDiscordMessageEmbed(sc *SchniffCollection, notification Notification) (*discordgo.MessageEmbed, error) {
	baseURL := "https://www.recreation.gov/camping/campsites/"
	schniff, err := sc.GetSchniff(notification.SchniffID)
	if err != nil {
		return nil, err
	}

	// Create an opening sentence with more details about the campground and the date range
	title := fmt.Sprintf("Successfully schniffed! %s from %s to %s",
		schniff.CampgroundName,
		schniff.StartDate.Format("2006-01-02"),
		schniff.EndDate.Format("2006-01-02"),
	)

	// Calculate total number of days in the date range
	totalDays := int(schniff.EndDate.Sub(schniff.StartDate).Hours()/24) + 1
	// fmt.Println(len(notification.AvailableCampsites))
	// fmt.Println(totalDays, schniff.EndDate, schniff.StartDate)
	// Create a map to hold campsite IDs and counts
	campsiteDayCount := make(map[string]int)

	// Populate the map with the count of days for each campsite
	for _, campsite := range notification.AvailableCampsites {
		campsiteDayCount[campsite.CampsiteID]++
	}

	// Convert map to a slice
	campsites := make([]campsiteWithDays, 0, len(campsiteDayCount))
	for id, count := range campsiteDayCount {
		campsites = append(campsites, campsiteWithDays{campsiteID: id, daysCount: count})
	}

	// Sort the slice in descending order by daysCount
	sort.Slice(campsites, func(i, j int) bool {
		return campsites[i].daysCount > campsites[j].daysCount
	})

	// Take top 10 campsites
	remainingSites := 0
	if len(campsites) > 10 {
		remainingSites = len(campsites) - 10
		campsites = campsites[:10]
	}

	// Prepare fields for the embed
	fields := make([]*discordgo.MessageEmbedField, len(campsites))

	// Add sorted campsites to the fields, with available days as percentage
	for i, campsite := range campsites {
		campsiteLink := baseURL + campsite.campsiteID
		availabilityPercentage := float64(campsite.daysCount) / float64(totalDays) * 100
		fields[i] = &discordgo.MessageEmbedField{
			Name:   campsite.campsiteID,
			Value:  fmt.Sprintf("[Covers %.0f%%](%s)", availabilityPercentage, campsiteLink),
			Inline: false,
		}
	}

	message := fmt.Sprintf(`Yo <@%s>, we found some availabilities open up for the time you're schniffing.
	Showing the top %d most available campsites. 
	Found %d in total.
	
	IMPORTANT: You must act fast to get one of these sites. Click the link and complete the form in the website.
	We have stopped monitoring the site. If you miss out on one of the availabilities below, please restart this schniff by typing '/restart-schniff'`,
		schniff.UserID,
		len(campsites),
		len(campsites)+remainingSites,
	)

	// Create the embed message
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: message,
		Fields:      fields,
		Color:       0x00ff00, // Change this to any color you want
	}

	return embed, nil
}
