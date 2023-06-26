package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type tracker struct {
	Notifications []Notification
	Requests      int
	LastReset     time.Time
	mu            sync.Mutex
}

func NewTracker() *tracker {
	return &tracker{
		Notifications: []Notification{},
		Requests:      0,
		LastReset:     time.Now(),
		mu:            sync.Mutex{},
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

	totalRequests := t.Requests
	totalNotifications := len(t.Notifications)
	elapsed := time.Since(t.LastReset).Hours()
	requestsPerHour := float64(totalRequests) / elapsed

	return &discordgo.MessageEmbed{
		Title: "Schniffbot Activity Summary over last " + fmt.Sprintf("%.2f", elapsed) + " hours",
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
				Name:   "Users Notified",
				Value:  usernamesString,
				Inline: false,
			},
		},
		Color: 0x00ff00, // Green color
	}
}
