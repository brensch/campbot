package main

import (
	"fmt"
	"math/rand"
	"time"
)

func RandomSillyGreeting(userID string) string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	greetings := []string{
		"Welcome to schnifftown, <@%s>.",
		"It's schniff o'clock for <@%s>.",
		"I was just thinking, today's a nice day to schniff, especially for <@%s>.",
		"If you're <@%s>, it's time to schniff.",
		"I was just talking about you, <@%s>. I said, \"I bet they're ready to schniff.\".",
		"If <@%s> were a verb, it would be \"schniff\".",
		"The humble consequence of carbon, <@%s> has arrived to schniff.",
		"There will be a day that is the end. The collapse of time and all that stood within it. A day of nothing. This is not that day for <@%s>. It's a day to schniff.",
		"In their last will and testament there is a codicil memorializing their appreciation for the schniffer and all those who serve it. <@%s> is ready to schniff.",
		"<@%s> was first seen standing at the edge of the shore between the ancient marks of the high and low tide, a place that is neither land nor sea. But as the moonlight filtered through the darkness, it revealed a schniffer who has been to the beyond and witnessed the secrets of life and death.",
	}

	// Choose a random greeting template
	template := greetings[r.Intn(len(greetings))]

	// Substitute the user ID into the template
	greeting := fmt.Sprintf(template, userID)
	greeting = fmt.Sprintf("%s\n\n%s", greeting, "Please check your DMs for instructions on how to use Schniffer.")

	return greeting
}

func RandomSillyHeader() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	greetings := []string{
		"You've got schniff!",
		"Schnifffff!",
		"Another day, another schniff.",
		"Oh, what a schniff!",
		"Schniff, schniff, hooray!",
		"Look what the schniffer dragged in!",
		"Sch-sch-sch-sch-schniff!",
	}

	// Choose a random greeting template
	return greetings[r.Intn(len(greetings))]

}
