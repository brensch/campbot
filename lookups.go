package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type ScoredCampground struct {
	campground SummarisedCampground
	score      int
}

func suggestBestMatchesForCampground(campgrounds []SummarisedCampground, userInput string) []*discordgo.ApplicationCommandOptionChoice {
	scoredCampgrounds := make([]ScoredCampground, 0, len(campgrounds))

	for _, campground := range campgrounds {
		lowerName := strings.ToLower(campground.Name)
		lowerParentName := strings.ToLower(campground.ParentName)
		lowerInput := strings.ToLower(userInput)

		nameScore, parentNameScore := 0, 0

		if strings.HasPrefix(lowerName, lowerInput) {
			// If name starts with user input, give a high score
			nameScore = 10000
		} else if strings.Contains(lowerName, lowerInput) {
			// If name contains user input, give a lower score
			nameScore = 5000
		} else {
			// Calculate Levenshtein distance as score
			distance := levenshtein.DistanceForStrings([]rune(lowerName), []rune(lowerInput), levenshtein.DefaultOptions)
			// We subtract the distance from a high number to get higher scores for closer matches
			nameScore = 10000 - distance
		}

		if strings.HasPrefix(lowerParentName, lowerInput) {
			// If parent name starts with user input, give a high score
			parentNameScore = 10000
		} else if strings.Contains(lowerParentName, lowerInput) {
			// If parent name contains user input, give a lower score
			parentNameScore = 5000
		} else {
			// Calculate Levenshtein distance as score
			distance := levenshtein.DistanceForStrings([]rune(lowerParentName), []rune(lowerInput), levenshtein.DefaultOptions)
			// We subtract the distance from a high number to get higher scores for closer matches
			parentNameScore = 10000 - distance
		}

		// We take the max score between nameScore and parentNameScore and add the campground's rating to it
		score := max(nameScore, parentNameScore)
		scoredCampgrounds = append(scoredCampgrounds, ScoredCampground{campground: campground, score: score})
	}

	// Sort the campgrounds by score in descending order and then by rating if scores are equal
	sort.Slice(scoredCampgrounds, func(i, j int) bool {
		if scoredCampgrounds[i].score == scoredCampgrounds[j].score {
			return scoredCampgrounds[i].campground.Rating > scoredCampgrounds[j].campground.Rating
		}
		return scoredCampgrounds[i].score > scoredCampgrounds[j].score
	})

	// Generate Discord options and add score to the option description
	var bestMatches []*discordgo.ApplicationCommandOptionChoice
	for i := 0; i < len(scoredCampgrounds) && i < 10; i++ {
		campground := scoredCampgrounds[i].campground
		// score := scoredCampgrounds[i].score
		description := fmt.Sprintf("%s [%s]", campground.Name, campground.ParentName)
		// need to truncate to 100 characters because of Discord's limit
		limit := 85
		if len(description) > limit {
			description = description[:limit]
			description += "..."
		}
		description = fmt.Sprintf("%s %.2f", description, campground.Rating)
		option := &discordgo.ApplicationCommandOptionChoice{
			Name:  description,
			Value: campground.ID,
		}
		bestMatches = append(bestMatches, option)
	}

	return bestMatches
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func suggestBestMatchesForSchniff(schniffs []*Schniff, userInput string) []*discordgo.ApplicationCommandOptionChoice {
	scoredSchniffs := make([]ScoredSchniff, 0, len(schniffs))

	for _, schniff := range schniffs {
		lowerName := strings.ToLower(schniff.CampgroundName)
		lowerInput := strings.ToLower(userInput)

		nameScore := 0

		if strings.HasPrefix(lowerName, lowerInput) {
			// If name starts with user input, give a high score
			nameScore = 10000
		} else if strings.Contains(lowerName, lowerInput) {
			// If name contains user input, give a lower score
			nameScore = 5000
		} else {
			// Calculate Levenshtein distance as score
			distance := levenshtein.DistanceForStrings([]rune(lowerName), []rune(lowerInput), levenshtein.DefaultOptions)
			// We subtract the distance from a high number to get higher scores for closer matches
			nameScore = 10000 - distance
		}

		// For schniffs, we are just using nameScore
		score := nameScore
		scoredSchniffs = append(scoredSchniffs, ScoredSchniff{schniff: schniff, score: score})
	}

	// Sort the schniffs by score in descending order
	sort.Slice(scoredSchniffs, func(i, j int) bool {
		return scoredSchniffs[i].score > scoredSchniffs[j].score
	})

	// Generate Discord options
	var bestMatches []*discordgo.ApplicationCommandOptionChoice
	for i := 0; i < len(scoredSchniffs) && i < 10; i++ {
		schniff := scoredSchniffs[i].schniff
		description := fmt.Sprintf("%s [%s -> %s]",
			schniff.CampgroundName,
			schniff.StartDate.Format("2006-01-02"),
			schniff.EndDate.Format("2006-01-02"),
		)
		// need to truncate to 100 characters because of Discord's limit
		limit := 85
		if len(description) > limit {
			description = description[:limit]
			description += "..."
		}
		option := &discordgo.ApplicationCommandOptionChoice{
			Name:  description,
			Value: schniff.SchniffID,
		}
		bestMatches = append(bestMatches, option)
	}

	return bestMatches
}

type ScoredSchniff struct {
	schniff *Schniff
	score   int
}
