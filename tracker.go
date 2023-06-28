package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type tracker struct {
	Notifications  []Notification
	Requests       int
	LastReset      time.Time
	ActiveSchniffs map[string]struct{}
	ActiveUsers    map[string]struct{}
	ActiveDays     map[time.Time]struct{}
	ActiveSites    map[string]struct{}
	mu             sync.Mutex
}

func NewTracker() *tracker {
	return &tracker{
		Notifications:  []Notification{},
		Requests:       0,
		LastReset:      time.Now(),
		ActiveSchniffs: make(map[string]struct{}),
		ActiveUsers:    make(map[string]struct{}),
		ActiveDays:     make(map[time.Time]struct{}),
		ActiveSites:    make(map[string]struct{}),

		mu: sync.Mutex{},
	}
}

func (t *tracker) AddNotification(notification Notification) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Notifications = append(t.Notifications, notification)
}

func (t *tracker) IncrementRequests(count int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Requests += count
}

func (t *tracker) AddActiveSchniff(schniffID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ActiveSchniffs[schniffID] = struct{}{}
}

func (t *tracker) AddActiveUser(userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ActiveUsers[userID] = struct{}{}
}

func (t *tracker) AddActiveDay(date time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ActiveDays[date] = struct{}{}
}

func (t *tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Notifications = []Notification{}
	t.Requests = 0
	t.LastReset = time.Now()
}

// CreateEmbedSummary creates a summary of the tracker state in a discordgo.MessageEmbed.
func (t *tracker) CreateEmbedSummary(sc *SchniffCollection) *discordgo.MessageEmbed {
	t.mu.Lock()
	defer t.mu.Unlock()

	usernames := make(map[string]struct{})
	for _, n := range t.Notifications {
		schniff, err := sc.GetSchniff(n.SchniffID)
		if err != nil {
			continue
		}
		usernames[schniff.UserNick] = struct{}{}
	}

	// Convert the map to a slice for easier formatting
	usernamesSlice := make([]string, 0, len(usernames))
	for username := range usernames {
		usernamesSlice = append(usernamesSlice, username)
	}

	// Join usernames with a comma and a space
	usernamesString := strings.Join(usernamesSlice, ", ")

	// Convert the map to a slice for easier formatting
	activeUsernamesSlice := make([]string, 0, len(t.ActiveUsers))
	for username := range t.ActiveUsers {
		activeUsernamesSlice = append(activeUsernamesSlice, username)
	}

	// Join usernames with a comma and a space
	activeUsernameString := strings.Join(activeUsernamesSlice, ", ")

	totalRequests := t.Requests
	totalNotifications := len(t.Notifications)
	elapsed := time.Since(t.LastReset).Hours()
	requestsPerHour := float64(totalRequests) / elapsed

	return &discordgo.MessageEmbed{
		Title: "Schniffer summary:\nLast " + fmt.Sprintf("%.2f", elapsed) + " hours",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Total Notifications",
				Value:  fmt.Sprintf("%d", totalNotifications),
				Inline: true,
			},
			{
				Name:   "Times checked",
				Value:  fmt.Sprintf("%d", totalRequests),
				Inline: true,
			},
			{
				Name:   "Requests per Hour",
				Value:  fmt.Sprintf("%.2f", requestsPerHour),
				Inline: true,
			},
			{
				Name:   "Schniffists Notified",
				Value:  usernamesString,
				Inline: false,
			},
			{
				Name:   "Schniffists with active schniffs",
				Value:  activeUsernameString,
				Inline: false,
			},
			{
				Name:   "Active Schniffs",
				Value:  fmt.Sprintf("%d", len(t.ActiveSchniffs)),
				Inline: false,
			},
			{
				Name:   "Days being tracked",
				Value:  fmt.Sprintf("%d", len(t.ActiveDays)),
				Inline: false,
			},
			{
				Name:   "Sites being tracked",
				Value:  fmt.Sprintf("%d", len(t.ActiveSites)),
				Inline: false,
			},
		},
		Color: 0x009900, // Green color
	}
}
