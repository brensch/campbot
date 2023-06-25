package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// func main() {
// 	log := zap.NewExample()
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			Dial: (&net.Dialer{
// 				Timeout:   60 * time.Second,
// 				KeepAlive: 60 * time.Second,
// 			}).Dial,
// 			TLSHandshakeTimeout:   30 * time.Second,
// 			ResponseHeaderTimeout: 30 * time.Second,
// 			ExpectContinueTimeout: 10 * time.Second,
// 		},
// 	}

// 	availability, err := GetAvailability(context.Background(), log, client, "10083840", time.Now())
// 	if err != nil {
// 		log.Error("couldn't get availability", zap.Error(err))
// 		return
// 	}

// 	TOKEN := os.Getenv("DISCORD_TOKEN")

// 	dg, _ := discordgo.New("Bot " + TOKEN)

// 	dg.AddHandler(messageCreate)

// 	dg.Open()
// 	<-make(chan struct{})

// 	log.Info("got availability", zap.Any("availability", availability))
// }

// func main() {
// 	token := os.Getenv("DISCORD_TOKEN")

// 	discord, err := discordgo.New("Bot " + token)
// 	if err != nil {
// 		fmt.Println("Error creating Discord session: ", err)
// 		return
// 	}

// 	http.HandleFunc("/", InteractionHandler(discord))

// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "3000" // Default to port 3000 if no PORT variable is set
// 	}

// 	fmt.Println("Server is running on Port:", port)
// 	err = http.ListenAndServe(":"+port, nil)
// 	if err != nil {
// 		fmt.Println("Error starting server: ", err)
// 		return
// 	}
// }

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

	log.Println(createdCommands)

	requests := ConstructAvailabilityRequests(ctx, olog, s.Client, sc)
	fmt.Println(requests)

	availabilities, err := DoRequests(ctx, olog, s.Client, requests)
	if err != nil {
		olog.Error("Unable to get availability", zap.Error(err))
	}

	// for _, availability := range availabilities {
	// 	json, _ := json.Marshal(availability)
	// 	fmt.Println(string(json))
	// }

	notifications, err := GenerateNotifications(ctx, olog, availabilities, sc)
	if err != nil {
		olog.Error("Unable to generate notifications", zap.Error(err))
	}

	guilds, err := s.UserGuilds(100, "", "")
	if err != nil {
		log.Fatalf("Cannot get guilds: %v", err)
	}

	for _, guild := range guilds {
		// Get channels of each guild
		fmt.Println(guild.ID)
		channels, err := s.GuildChannels(guild.ID)
		if err != nil {
			fmt.Println("Error getting channels: ", err)
			return
		}

		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildText {
				fmt.Println("Channel ID: ", channel.ID, channel.Name)
				for _, notification := range notifications {
					_, err = s.ChannelMessageSend(channel.ID, GenerateDiscordMessage(notification))
					if err != nil {
						olog.Error("Unable to send message", zap.Error(err))
					}
					s.ChannelMessageSendEmbeds(channel.ID, []*discordgo.MessageEmbed{GenerateDiscordMessageEmbed(notification)})

				}

			}
		}
	}

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

func DoNotification(ctx context.Context, olog *zap.Logger, userID string) error {
	olog.Info("got thing", zap.String("userID", userID))
	return nil
}

var (
	// This seems sufficient. There are fancier ways to do it but i'm not seeing any problems doing it like this so yolo
	userAgents = []string{
		"Opera/9.89 (X11; Linux i686; en-US) Presto/2.9.160 Version/10.00",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows 98; Win 9x 4.90; Trident/5.0)",
		"Mozilla/5.0 (Macintosh; U; PPC Mac OS X 10_7_1 rv:3.0) Gecko/20140828 Firefox/35.0",
		"Opera/9.75 (X11; Linux i686; en-US) Presto/2.11.287 Version/10.00",
		"Mozilla/5.0 (Windows; U; Windows 98; Win 9x 4.90) AppleWebKit/535.32.3 (KHTML, like Gecko) Version/4.0.1 Safari/535.32.3",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1) AppleWebKit/534.38.1 (KHTML, like Gecko) Version/4.0.5 Safari/534.38.1",
		"Mozilla/5.0 (Windows; U; Windows 98) AppleWebKit/534.46.6 (KHTML, like Gecko) Version/4.1 Safari/534.46.6",
		"Mozilla/5.0 (compatible; MSIE 5.0; Windows NT 5.01; Trident/3.0)",
		"Opera/9.41 (Windows NT 4.0; en-US) Presto/2.12.197 Version/10.00",
		"Mozilla/5.0 (Macintosh; U; PPC Mac OS X 10_6_1) AppleWebKit/5320 (KHTML, like Gecko) Chrome/36.0.828.0 Mobile Safari/5320",
		"Mozilla/5.0 (Windows CE; sl-SI; rv:1.9.2.20) Gecko/20120730 Firefox/35.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:5.0) Gecko/20190718 Firefox/35.0",
		"Mozilla/5.0 (compatible; MSIE 6.0; Windows NT 5.0; Trident/4.1)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 7_2_2 like Mac OS X; en-US) AppleWebKit/535.12.1 (KHTML, like Gecko) Version/4.0.5 Mobile/8B118 Safari/6535.12.1",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/5331 (KHTML, like Gecko) Chrome/39.0.814.0 Mobile Safari/5331",
		"Mozilla/5.0 (Windows 98) AppleWebKit/5331 (KHTML, like Gecko) Chrome/37.0.821.0 Mobile Safari/5331",
		"Opera/9.22 (Windows 98; Win 9x 4.90; en-US) Presto/2.12.176 Version/12.00",
		"Mozilla/5.0 (Macintosh; PPC Mac OS X 10_8_8 rv:6.0) Gecko/20160922 Firefox/35.0",
		"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_8_0) AppleWebKit/5350 (KHTML, like Gecko) Chrome/37.0.875.0 Mobile Safari/5350",
		"Opera/9.10 (Windows NT 6.2; sl-SI) Presto/2.9.276 Version/12.00",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1) AppleWebKit/531.39.4 (KHTML, like Gecko) Version/5.1 Safari/531.39.4",
		"Mozilla/5.0 (X11; Linux i686; rv:7.0) Gecko/20140323 Firefox/37.0",
		"Mozilla/5.0 (Windows NT 5.01; en-US; rv:1.9.1.20) Gecko/20170530 Firefox/35.0",
		"Mozilla/5.0 (X11; Linux i686) AppleWebKit/5362 (KHTML, like Gecko) Chrome/40.0.840.0 Mobile Safari/5362",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows CE; Trident/5.1)",
		"Mozilla/5.0 (Windows; U; Windows NT 6.0) AppleWebKit/531.2.3 (KHTML, like Gecko) Version/5.1 Safari/531.2.3",
		"Mozilla/5.0 (compatible; MSIE 5.0; Windows 95; Trident/4.0)",
		"Mozilla/5.0 (compatible; MSIE 5.0; Windows NT 4.0; Trident/5.1)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 8_1_2 like Mac OS X; en-US) AppleWebKit/532.19.4 (KHTML, like Gecko) Version/4.0.5 Mobile/8B119 Safari/6532.19.4",
		"Opera/8.92 (X11; Linux i686; sl-SI) Presto/2.11.179 Version/12.00",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows 98; Trident/3.1)",
		"Mozilla/5.0 (compatible; MSIE 5.0; Windows NT 5.2; Trident/3.1)",
		"Mozilla/5.0 (Macintosh; PPC Mac OS X 10_7_4 rv:6.0; sl-SI) AppleWebKit/533.22.1 (KHTML, like Gecko) Version/4.1 Safari/533.22.1",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/5332 (KHTML, like Gecko) Chrome/39.0.882.0 Mobile Safari/5332",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows 98; Win 9x 4.90; Trident/5.1)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_5_0 rv:6.0) Gecko/20110829 Firefox/36.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 7_0_2 like Mac OS X; en-US) AppleWebKit/531.12.6 (KHTML, like Gecko) Version/3.0.5 Mobile/8B118 Safari/6531.12.6",
		"Opera/9.61 (X11; Linux x86_64; sl-SI) Presto/2.11.287 Version/12.00",
		"Opera/9.55 (X11; Linux i686; sl-SI) Presto/2.9.288 Version/12.00",
		"Mozilla/5.0 (compatible; MSIE 11.0; Windows NT 5.1; Trident/4.0)",
	}
)

func RandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}
