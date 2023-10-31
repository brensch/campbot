package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func HandleNewSchniff(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection, cc *CampgroundCollection) {
	data := i.ApplicationCommandData()

	var campground SummarisedCampground
	var startDate, endDate time.Time
	var campsiteList []string
	minConsecutiveDays := int64(1)
	var err error
	for _, option := range data.Options {
		switch option.Name {
		case "campground":
			campground, err = cc.GetCampground(option.StringValue())
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Campground not found: %v", option.StringValue()),
					},
				})
				return
			}
		case "start":
			startDate, err = time.Parse("2006-01-02", option.StringValue())
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Invalid start date: %v", err),
					},
				})
				return
			}
		case "end":
			endDate, err = time.Parse("2006-01-02", option.StringValue())
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Invalid end date: %v", err),
					},
				})
				return
			}
		case "campsite-list":
			campsiteList = strings.Split(option.StringValue(), ",")
		case "minimum-consecutive-days":
			minConsecutiveDays = option.IntValue()
		}
	}

	if startDate.After(endDate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Start date must be before end date",
			},
		})
		return
	}

	_ = campsiteList
	_ = minConsecutiveDays

	var user *discordgo.User
	if i.Member == nil {
		user = i.User
	} else {
		user = i.Member.User
	}

	schniff := &Schniff{
		CampgroundID:           data.Options[0].StringValue(),
		CampgroundName:         campground.Name,
		StartDate:              startDate,
		EndDate:                endDate,
		UserID:                 user.ID,
		UserNick:               user.Username,
		SchniffID:              uuid.New().String(),
		Active:                 true,
		CreationTime:           time.Now(),
		CampsiteIDs:            campsiteList,
		MinimumConsecutiveDays: minConsecutiveDays,
	}

	err = sc.Add(schniff)
	if err != nil {
		log.Error("Cannot add schniff", zap.Error(err))
	}

	duration := endDate.Sub(startDate).Hours() / 24 // calculates duration in days

	embed := &discordgo.MessageEmbed{
		Title: "New Schniff Created",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Schniff ID",
				Value:  schniff.SchniffID,
				Inline: true,
			},
			{
				Name:   "Campground ID",
				Value:  schniff.CampgroundID,
				Inline: true,
			},
			{
				Name:   "Campground Name",
				Value:  schniff.CampgroundName,
				Inline: true,
			},
			{
				Name:   "Start Date",
				Value:  schniff.StartDate.Format("2006-01-02"),
				Inline: true,
			},
			{
				Name:   "End Date",
				Value:  schniff.EndDate.Format("2006-01-02"),
				Inline: true,
			},
			{
				Name:   "Duration",
				Value:  fmt.Sprintf("%.0f days", duration),
				Inline: true,
			},
			{
				Name:   "User ID",
				Value:  schniff.UserID,
				Inline: true,
			},
			{
				Name:   "User Nickname",
				Value:  schniff.UserNick,
				Inline: true,
			},
			{
				Name:   "Active",
				Value:  fmt.Sprintf("%v", schniff.Active),
				Inline: true,
			},
		},
		Color: 0x009900, // Green color
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.Error("Cannot respond to interaction", zap.Error(err))
	}
}

func HandleNewSchniffAutocomplete(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection, cc *CampgroundCollection) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice
	switch {
	// In this case there are multiple autocomplete options. The Focused field shows which option user is focused on.
	case data.Options[0].Focused:
		allChoices := cc.GetCampgrounds()
		userInput := data.Options[0].StringValue()
		choices = suggestBestMatchesForCampground(allChoices, userInput)
	}

	if len(choices) > 10 {
		choices = choices[:10]
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Error("Cannot respond to interaction", zap.Error(err))
	}
}

func HandleViewSchniffs(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection) {
	// get all this user's schniffs
	var user *discordgo.User
	if i.Member == nil {
		user = i.User
	} else {
		user = i.Member.User
	}
	schniffs := sc.GetSchniffsForUser(user.ID)
	table := GenerateEmbedMessage(schniffs)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{table},
		},
	})
	if err != nil {
		log.Error("Cannot respond to interaction", zap.Error(err))
	}
}

func HandleRestartSchniff(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection) {
	data := i.ApplicationCommandData()

	schniffID := data.Options[0].StringValue()

	err := sc.SetActive(schniffID, true)
	if err != nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		if err != nil {
			log.Error("Cannot respond to interaction", zap.Error(err))
			return
		}
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Successfully restarted your schniff.",
		},
	})
	if err != nil {
		log.Error("Cannot respond to interaction", zap.Error(err))
		return
	}
}

func HandleStopSchniff(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection) {
	data := i.ApplicationCommandData()

	schniffID := data.Options[0].StringValue()

	err := sc.SetActive(schniffID, false)
	if err != nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		if err != nil {
			log.Error("Cannot respond to interaction", zap.Error(err))
			return
		}
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Successfully stopped your schniff.",
		},
	})
	if err != nil {
		log.Error("Cannot respond to interaction", zap.Error(err))
		return
	}
}

func HandleRestartSchniffAutocomplete(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice
	var user *discordgo.User
	if i.Member == nil {
		user = i.User
	} else {
		user = i.Member.User
	}
	switch {
	// In this case there are multiple autocomplete options. The Focused field shows which option user is focused on.
	case data.Options[0].Focused:
		allChoices := sc.GetSchniffsForUser(user.ID)
		var stoppedChoices []*Schniff
		for _, schniff := range allChoices {
			if schniff.Active {
				continue
			}
			stoppedChoices = append(stoppedChoices, schniff)
		}
		userInput := data.Options[0].StringValue()
		choices = suggestBestMatchesForSchniff(stoppedChoices, userInput)
	}

	if len(choices) > 10 {
		choices = choices[:10]
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Error("Cannot respond to interaction", zap.Error(err))
	}
}

func HandleStopSchniffAutocomplete(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice
	var user *discordgo.User
	if i.Member == nil {
		user = i.User
	} else {
		user = i.Member.User
	}
	switch {
	// In this case there are multiple autocomplete options. The Focused field shows which option user is focused on.
	case data.Options[0].Focused:
		allChoices := sc.GetSchniffsForUser(user.ID)
		var runningChoices []*Schniff
		for _, schniff := range allChoices {
			if !schniff.Active {
				continue
			}
			runningChoices = append(runningChoices, schniff)
		}
		userInput := data.Options[0].StringValue()
		choices = suggestBestMatchesForSchniff(runningChoices, userInput)
	}

	if len(choices) > 10 {
		choices = choices[:10]
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Error("Cannot respond to interaction", zap.Error(err))
	}
}

func HandleGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	embed := &discordgo.MessageEmbed{
		Color:       0x009900, // Green
		Title:       "Let's get Schniffing",
		Description: "Hello <@" + m.Member.User.ID + ">. Follow the 3 step schniff plan to experience the joys of schniffing.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "1",
				Value:  "Type `/new-schniff` to create a new schniff.",
				Inline: false,
			},
			{
				Name:   "2",
				Value:  "Fill out the fields that pop up, and press enter.",
				Inline: false,
			},
			{
				Name: "3",
				Value: `I will dutifully schniff until I find something, and message you here when I do.
				Go to the link in the message I send you and book the site ASAP.`,
				Inline: false,
			},
			{
				Name: "Notes",
				Value: `- We will notify you of even partial availability for a period you are schniffing. 
- It is important to think about schniffer whilst enjoying your schniffed camping experience.
- If you don't book the site fast enough, restart the schniff by typing '/restart-schniff' and selecting the schniff you want to restart.
- I will notify you 15 seconds after it becomes available. It will normally be rebooked in about 10 minutes. Please panic when I message you.`,
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "To err is human, to schniff is divine.",
		},
	}

	dmChannel, err := s.UserChannelCreate(m.User.ID)
	if err != nil {
		return
	}

	_, err = s.ChannelMessageSendEmbed(dmChannel.ID, embed)
	if err != nil {
		return
	}

	sendMessageToChannelInAllGuilds(s, "announcements", RandomSillyGreeting(m.Member.User.ID))
}
