package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Notification struct {
	AvailableCampsites []CampsiteAvailability
	Schniff            Schniff
}

type CampsiteAvailability struct {
	CampsiteID string
	Date       time.Time
}

type campsiteWithDays struct {
	campsiteID string
	daysCount  int
}

func GenerateDiscordMessage(notification Notification) string {
	var message string
	baseURL := "https://www.recreation.gov/camping/campsites/"

	// Create an opening sentence with more details about the campground and the date range
	message = fmt.Sprintf("Hello %s, here is the availability of campsites for campground %s (ID: %s) from %s to %s:\n",
		notification.Schniff.UserNick,
		notification.Schniff.CampgroundName,
		notification.Schniff.CampgroundID,
		notification.Schniff.StartDate.Format("2006-01-02"),
		notification.Schniff.EndDate.Format("2006-01-02"),
	)

	message += "```\nCampsite  Coverage  Link\n"

	// Calculate total number of days in the date range
	totalDays := int(notification.Schniff.EndDate.Sub(notification.Schniff.StartDate).Hours() / 24)

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
	return message
}

func GenerateDiscordMessageEmbed(notification Notification) *discordgo.MessageEmbed {
	baseURL := "https://www.recreation.gov/camping/campsites/"

	// Create an opening sentence with more details about the campground and the date range
	title := fmt.Sprintf("Availability of campsites for %s (ID: %s) from %s to %s",
		notification.Schniff.CampgroundName,
		notification.Schniff.CampgroundID,
		notification.Schniff.StartDate.Format("2006-01-02"),
		notification.Schniff.EndDate.Format("2006-01-02"),
	)

	// Calculate total number of days in the date range
	totalDays := int(notification.Schniff.EndDate.Sub(notification.Schniff.StartDate).Hours() / 24)

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
	if len(campsites) > 10 {
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
			Value:  fmt.Sprintf("Available for %.2f%% of the requested period. [Link](%s)", availabilityPercentage, campsiteLink),
			Inline: false,
		}
	}

	// Create the embed message
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: fmt.Sprintf("Hello %s, here is the availability of campsites:", notification.Schniff.UserNick),
		Fields:      fields,
		Color:       0x00ff00, // Change this to any color you want
	}

	return embed
}
