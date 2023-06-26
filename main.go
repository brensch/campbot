package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "1122072835985244201", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	ctx := context.Background()
	log := zap.NewExample()

	cc, err := NewCampgroundCollection(ctx, log, s.Client)
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
	err = s.Open()
	if err != nil {
		log.Fatal("Cannot open the session", zap.Error(err))
	}
	defer s.Close()

	createdCommands, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, *GuildID, commands)
	if err != nil {
		log.Fatal("Cannot register commands", zap.Error(err))
	}

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for {
			loop(ctx, log, s, sc)
			select {
			case <-ticker.C:
			case <-ctx.Done():
				log.Info("Context done")
				return
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Info("Gracefully shutting down")

	for _, cmd := range createdCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, cmd.ID)
		if err != nil {
			log.Error("Cannot delete command: %v", zap.String("command", cmd.Name), zap.Error(err))
		}
	}

}

func loop(ctx context.Context, olog *zap.Logger, s *discordgo.Session, sc *SchniffCollection) {
	requests := ConstructAvailabilityRequests(ctx, olog, s.Client, sc)

	availabilities, err := DoRequests(ctx, olog, s.Client, requests)
	if err != nil {
		olog.Error("Unable to get availability", zap.Error(err))
	}

	notifications, err := GenerateNotifications(ctx, olog, availabilities, sc)
	if err != nil {
		olog.Error("Unable to generate notifications", zap.Error(err))
	}

	for _, notification := range notifications {
		schniff, err := sc.GetSchniff(notification.SchniffID)
		if err != nil {
			olog.Error("no such schniff", zap.Error(err))
			continue
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
		fmt.Println(schniff.SchniffID)
		err = sc.SetActive(schniff.SchniffID, false)
		if err != nil {
			olog.Error("Unable to mark as inactive", zap.Error(err))
			continue
		}

	}
}
