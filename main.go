package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log := zap.NewExample()

	botToken := os.Getenv("BOT_TOKEN")
	GuildID := os.Getenv("GUILD_ID")

	var s *discordgo.Session
	s, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal("Invalid bot parameters", zap.Error(err))
	}

	cc, err := NewCampgroundCollection(ctx, log, s, s.Client)
	if err != nil {
		log.Fatal("Cannot get campground collection", zap.Error(err))
	}

	sc := NewSchniffCollection("schniffs.json")
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) { log.Info("ready to schniff") })
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(log, s, i, sc, cc)
		}
	})
	s.AddHandler(HandleGuildMemberAdd)

	s.Identify.Intents = discordgo.IntentsGuildMembers

	err = s.Open()
	if err != nil {
		log.Fatal("Cannot open the session", zap.Error(err))
	}
	defer s.Close()

	// Register the commands
	_, err = s.ApplicationCommandBulkOverwrite(s.State.User.ID, GuildID, commands)
	if err != nil {
		log.Fatal("Cannot register commands", zap.Error(err))
	}

	// Notify that the bot is online
	err = sendMessageToChannelInAllGuilds(s, "announcements", "Schniffbot is online, ready to schniff.")
	if err != nil {
		log.Error("Unable to send message", zap.Error(err))
	}

	t := NewTracker()

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for {
			loop(ctx, log, s, sc, t)
			select {
			case <-ticker.C:
			case <-ctx.Done():
				log.Info("Context done")
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				log.Info("Context done")
				return
			}
			embed := t.CreateEmbedSummary(sc)
			err := sendEmbedToChannelInAllGuilds(s, "announcements", embed)
			if err != nil {
				log.Error("Unable to send tracker update", zap.Error(err))
			}
		}
	}()

	go func() {
		for {
			// Calculate next duration
			now := time.Now()
			location, _ := time.LoadLocation("America/Los_Angeles") // PST timezone
			now = now.In(location)
			next := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, location)
			if now.After(next) {
				// If time has passed 9 PM today, schedule for next day
				next = next.Add(24 * time.Hour)
			}
			duration := next.Sub(now)

			ticker := time.NewTimer(duration)
			select {
			case <-ticker.C:
				embed := t.CreateEmbedSummary(sc)
				err := sendEmbedToChannelInAllGuilds(s, "announcements", embed)
				if err != nil {
					log.Error("Unable to send tracker update", zap.Error(err))
				}
				ticker.Reset(24 * time.Hour) // Reset to 24 hours
			case <-ctx.Done():
				log.Info("Context done")
				return
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	cancel()
	log.Info("Gracefully shutting down")

	err = sendMessageToChannelInAllGuilds(s, "announcements", "Shutting down schniffbot")
	if err != nil {
		log.Error("Unable to send message", zap.Error(err))
	}

}

func nextDuration(hour, minute, second int) time.Duration {
	pst, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		// needs to never fail
		panic(err)
	}
	now := time.Now()
	nextTick := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, pst)
	if now.After(nextTick) {
		nextTick = nextTick.Add(24 * time.Hour)
	}
	return nextTick.Sub(now)
}

func loop(ctx context.Context, olog *zap.Logger, s *discordgo.Session, sc *SchniffCollection, t *tracker) {
	requests := ConstructAvailabilityRequests(ctx, olog, s.Client, sc)

	// Deduplicate requests
	deduplicatedRequests := DeduplicateAvailabilityRequests(requests)
	t.IncrementRequests(len(deduplicatedRequests))

	availabilities, err := DoRequests(ctx, olog, s.Client, deduplicatedRequests)
	if err != nil {
		olog.Error("Unable to get availability", zap.Error(err))
		sendMessageToChannelInAllGuilds(s, "problemos", fmt.Sprintf("Unable to get availability: %+v", err))
		return
	}

	notifications, err := GenerateNotifications(ctx, olog, availabilities, sc)
	if err != nil {
		sendMessageToChannelInAllGuilds(s, "problemos", fmt.Sprintf("Unable to generate notifications: %+v", err))
		olog.Error("Unable to generate notifications", zap.Error(err))
	}

	for _, notification := range notifications {
		schniff, err := sc.GetSchniff(notification.SchniffID)
		if err != nil {
			olog.Error("no such schniff", zap.Error(err))
			continue
		}
		t.AddActiveSchniff(schniff.SchniffID)
		t.AddActiveUser(schniff.UserID)
		currentDate := schniff.StartDate
		for currentDate.Before(schniff.EndDate) || currentDate.Equal(schniff.EndDate) {
			t.AddActiveDay(currentDate)
			currentDate = currentDate.AddDate(0, 0, 1)
		}

		dmChannel, err := s.UserChannelCreate(schniff.UserID)
		if err != nil {
			olog.Error("Unable to create dmChannel", zap.Error(err))
			continue
		}
		embeddedContents, err := GenerateDiscordMessageEmbed(sc, notification)
		if err != nil {
			olog.Error("Unable to generate embedded message", zap.Error(err))
			continue
		}
		_, err = s.ChannelMessageSendEmbeds(dmChannel.ID, []*discordgo.MessageEmbed{embeddedContents})
		if err != nil {
			olog.Error("Unable to send embeds", zap.Error(err))
			continue
		}

		// Mark the schniff as inactive
		err = sc.SetActive(schniff.SchniffID, false)
		if err != nil {
			olog.Error("Unable to mark as inactive", zap.Error(err))
			sendMessageToChannelInAllGuilds(s, "problemos", fmt.Sprintf("Unable to mark schniff as inactive: %+v", err))
			continue
		}

		sendMessageToChannelInAllGuilds(s, "announcements", RandomSillyBroadcast(schniff.UserID))

		// record we sent the notification
		t.AddNotification(notification)

	}
}
