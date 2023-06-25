package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

const (
	CommandNewSchniff  = "new-schniff"
	CommandViewSchniff = "view-schniff"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        CommandNewSchniff,
			Description: "Set up a new schniff",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "campground",
					Description:  "Campground Name",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:         "start",
					Description:  "Start (YYYY-MM-DD)",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: false,
				},
				{
					Name:         "end",
					Description:  "End (YYYY-MM-DD)",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: false,
				},
				{
					Name:         "campsite-list",
					Description:  "List of campsite IDs (separated by comma)",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     false,
					Autocomplete: false,
				},
			},
		},
		{
			Name:        CommandViewSchniff,
			Description: "See all schniffs belonging to you.",
			Type:        discordgo.ChatApplicationCommand,
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection){
		CommandViewSchniff: func(s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:

				// get all this user's schniffs
				schniffs := sc.GetSchniffsForUser(i.Member.User.ID)
				table := GenerateTableMessage(schniffs)

				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: table,
					},
				})
				if err != nil {
					panic(err)
				}
			}
		},
		CommandNewSchniff: func(s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				data := i.ApplicationCommandData()
				startDate, err := time.Parse("2006-01-02", data.Options[1].StringValue())
				if err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("Invalid start date: %v", err),
						},
					})
					return
				}

				endDate, err := time.Parse("2006-01-02", data.Options[2].StringValue())
				if err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("Invalid end date: %v", err),
						},
					})
					return
				}

				schniff := &Schniff{
					CampgroundID:   data.Options[0].StringValue(),
					CampgroundName: data.Options[0].Name,
					StartDate:      startDate,
					EndDate:        endDate,
					UserID:         i.Member.User.ID,
					UserNick:       i.Member.User.Username,
					SchniffID:      uuid.New().String(),
					Active:         true,
				}

				err = sc.Add(schniff)
				if err != nil {
					panic(err)
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Selected: %+v", schniff),
					},
				})
				if err != nil {
					panic(err)
				}
			case discordgo.InteractionApplicationCommandAutocomplete:
				data := i.ApplicationCommandData()
				var choices []*discordgo.ApplicationCommandOptionChoice
				switch {
				// In this case there are multiple autocomplete options. The Focused field shows which option user is focused on.
				case data.Options[0].Focused:
					choices = []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Yosemite",
							Value: "10083840",
						},
						{
							Name:  "Arroyo",
							Value: "231958",
						},
					}
				}

				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionApplicationCommandAutocompleteResult,
					Data: &discordgo.InteractionResponseData{
						Choices: choices,
					},
				})
				if err != nil {
					panic(err)
				}
			}
		},
	}
)
