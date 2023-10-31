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
	title := fmt.Sprintf("%s\n%s\n%s to %s",
		RandomSillyHeader(),
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
	fields := make([]*discordgo.MessageEmbedField, len(campsites)+1)

	// Add sorted campsites to the fields, with available days as percentage
	for i, campsite := range campsites {
		campsiteLink := baseURL + campsite.campsiteID
		daysAvailableString := ""
		daysCount := 0
		sort.Slice(notification.AvailableCampsites, func(i, j int) bool {
			return notification.AvailableCampsites[i].Date.Before(notification.AvailableCampsites[j].Date)
		})
		for _, availableCampsite := range notification.AvailableCampsites {
			if campsite.campsiteID != availableCampsite.CampsiteID {
				continue
			}
			dayOfWeek := availableCampsite.Date.Weekday().String()
			dateString := availableCampsite.Date.Format("2006-01-02")
			daysAvailableString += fmt.Sprintf("%s (%s)\n", dayOfWeek, dateString)
			daysCount++
			if daysCount == 10 {
				break
			}

		}
		if daysCount < campsite.daysCount {
			daysAvailableString += fmt.Sprintf("...and %d more", campsite.daysCount-daysCount)
		}
		fields[i] = &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Campsite %s", campsite.campsiteID),
			Value:  fmt.Sprintf("[%d of %d days available](%s)\n%s", campsite.daysCount, totalDays, campsiteLink, daysAvailableString),
			Inline: false,
		}
	}

	message := fmt.Sprintf(`<@%s>, I just schniffed some available campsites for you.
Showing the top %d campsites by days available.
%d total campsites with availabilities.`,
		schniff.UserID,
		len(campsites),
		len(campsites)+remainingSites,
	)

	fields[len(campsites)] = &discordgo.MessageEmbedField{
		Name: "Remember",
		Value: `- You must act fast to get one of these sites. 
- The links above take you directly to the campsite page to book. Find the availability and click it.
- If there are no availabilities when you clicked the link, it is because you were too slow. I do not make mistakes.
- The recreation.gov mobile app will sometimes open to the last page you were looking at despite clicking a link to a different page. Double check the link has opened correctly.`,
	}

	// Create the embed message
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: message,
		Fields:      fields,
		Color:       0x009900, // Change this to any color you want
	}

	return embed, nil
}
