package main

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	CommandNewSchniff     = "new-schniff"
	CommandViewSchniffs   = "view-schniffs"
	CommandRestartSchniff = "restart-schniff"
	CommandStopSchniff    = "stop-schniff"
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
				{
					Name:         "minimum-consecutive-days",
					Description:  "Minimum number of consecutive available days required to trigger a notification. (Default: 1)",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     false,
					Autocomplete: false,
				},
			},
		},
		{
			Name:        CommandViewSchniffs,
			Description: "See all schniffs belonging to you.",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        CommandRestartSchniff,
			Description: "Start a schniff running again",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "schniff-id",
					Description:  "Schniff",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        CommandStopSchniff,
			Description: "Stop a schniff",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "schniff-id",
					Description:  "Schniff",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
			},
		},
	}

	commandHandlers = map[string]func(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection, cc *CampgroundCollection){
		CommandViewSchniffs: func(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection, cc *CampgroundCollection) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				HandleViewSchniffs(log, s, i, sc)

			}
		},
		CommandNewSchniff: func(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection, cc *CampgroundCollection) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				HandleNewSchniff(log, s, i, sc, cc)
			case discordgo.InteractionApplicationCommandAutocomplete:
				HandleNewSchniffAutocomplete(log, s, i, sc, cc)
			}
		},
		CommandRestartSchniff: func(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection, cc *CampgroundCollection) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				HandleRestartSchniff(log, s, i, sc)
			case discordgo.InteractionApplicationCommandAutocomplete:
				HandleRestartSchniffAutocomplete(log, s, i, sc)
			}
		},
		CommandStopSchniff: func(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection, cc *CampgroundCollection) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				HandleStopSchniff(log, s, i, sc)
			case discordgo.InteractionApplicationCommandAutocomplete:
				HandleStopSchniffAutocomplete(log, s, i, sc)
			}
		},
	}
)
