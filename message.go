package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func sendMessageToChannelInAllGuilds(s *discordgo.Session, channelName string, message string) error {
	// Fetch all guilds (servers) the bot is a member of
	guilds := s.State.Guilds

	// If there are no guilds, return an error
	if len(guilds) == 0 {
		return fmt.Errorf("the bot is not a member of any guilds")
	}

	// Loop through each guild
	for _, guild := range guilds {
		// Get all channels for the guild
		channels, err := s.GuildChannels(guild.ID)
		if err != nil {
			return err
		}

		// Loop through to find the matching one
		var targetChannel *discordgo.Channel
		for _, channel := range channels {
			if channel.Name == channelName {
				targetChannel = channel
				break
			}
		}

		// If we didn't find the channel, continue to the next guild
		if targetChannel == nil {
			continue
		}

		// Send a message to the target channel
		_, err = s.ChannelMessageSend(targetChannel.ID, message)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendEmbedToChannelInAllGuilds(s *discordgo.Session, channelName string, embed *discordgo.MessageEmbed) error {
	// Fetch all guilds (servers) the bot is a member of
	guilds := s.State.Guilds

	// If there are no guilds, return an error
	if len(guilds) == 0 {
		return fmt.Errorf("the bot is not a member of any guilds")
	}

	// Loop through each guild
	for _, guild := range guilds {
		// Get all channels for the guild
		channels, err := s.GuildChannels(guild.ID)
		if err != nil {
			return err
		}

		// Loop through to find the matching one
		var targetChannel *discordgo.Channel
		for _, channel := range channels {
			if channel.Name == channelName {
				targetChannel = channel
				break
			}
		}

		// If we didn't find the channel, continue to the next guild
		if targetChannel == nil {
			continue
		}

		// Send a message to the target channel
		_, err = s.ChannelMessageSendEmbed(targetChannel.ID, embed)
		if err != nil {
			return err
		}
	}

	return nil
}
