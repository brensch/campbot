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
	olog := zap.NewExample()
	sc := NewSchniffCollection("schniffs.json")
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) { log.Println("Bot is up!") })
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i, sc)
		}
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	createdCommands, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, *GuildID, commands)
	if err != nil {
		log.Fatalf("Cannot register commands: %v", err)
	}

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for {
			loop(ctx, olog, s, sc)
			select {
			case <-ticker.C:
			case <-ctx.Done():
				olog.Info("Context done")
				return
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Gracefully shutting down")

	for _, cmd := range createdCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, cmd.ID)
		if err != nil {
			log.Fatalf("Cannot delete %q command: %v", cmd.Name, err)
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
