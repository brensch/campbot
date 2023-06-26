package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func HandleNewSchniff(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection, cc *CampgroundCollection) {
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

	campground, err := cc.GetCampground(data.Options[0].StringValue())
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Campground not found: %v", data.Options[0].StringValue()),
			},
		})
		return
	}

	schniff := &Schniff{
		CampgroundID:   data.Options[0].StringValue(),
		CampgroundName: campground.Name,
		StartDate:      startDate,
		EndDate:        endDate,
		UserID:         i.Member.User.ID,
		UserNick:       i.Member.User.Username,
		SchniffID:      uuid.New().String(),
		Active:         true,
	}

	err = sc.Add(schniff)
	if err != nil {
		log.Error("Cannot add schniff", zap.Error(err))
	}

	duration := endDate.Sub(startDate).Hours() / 24 // calculates duration in days

	embed := &discordgo.MessageEmbed{
		Title: "New Schniff Created!",
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
		Color: 0x00ff00, // Green color
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
	schniffs := sc.GetSchniffsForUser(i.Member.User.ID)
	table := GenerateTableMessage(schniffs)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: table,
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

func HandleRestartSchniffAutocomplete(log *zap.Logger, s *discordgo.Session, i *discordgo.InteractionCreate, sc *SchniffCollection) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice
	switch {
	// In this case there are multiple autocomplete options. The Focused field shows which option user is focused on.
	case data.Options[0].Focused:
		allChoices := sc.GetSchniffsForUser(i.Member.User.ID)
		userInput := data.Options[0].StringValue()
		choices = suggestBestMatchesForSchniff(allChoices, userInput)
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
