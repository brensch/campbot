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

func RandomSillyBroadcast(userID string) string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	greetings := []string{
		"I came, I saw, I schniffed <@%s> a campsite. ",
		"If schniffing were an olympic sport, <@%s> would be Steven Bradbury since I just found them a campsite.",
		"When you stare into the schniff, the schniff stares back. Is what <@%s> is saying right now because I found them a campsite.",
		"These messages are not generated by chatgpt. Neither is the campsite I just found for <@%s>.",
		"Schniffer's the name, schniffing's the game. <@%s> is one campsite richer.",
		"<@%s> is thinking, why am I getting so many notifications? It's because I just successfully schniffed for them.",
		"Can <@%s> remember their recreation.gov login credentials? They'll need them to book the campsite I just found for them in time.",
		"That's one small schniff for <@%s>, one giant leap for schniffkind.",
		"The schniff will set <@%s> free. Free to book the campsite I just found for them. But not free, you have to pay.",
		"The only thing we have to fear is fear itself. And not booking the campsite I just found for <@%s>.",
		"80%% of success is showing up. The other 20%% is schniffing. <@%s> is 100%% successful.",
		"Frankly, my dear, I don't give a schniff. But I do give a campsite to <@%s>.",
		"Hell is other people. But heaven is a campsite I just found for <@%s>.",
		"I love the smell of schniff in the morning. It smells like <@%s>'s dms.",
		"If you want something done right, you have to do it yourself. Or you can just use schniffer and I'll do it for you, like I just did for <@%s>.",
		"I'm gonna schniff <@%s> a campsite they can't refuse.",
		"Go ahead, make my schniff. I just found <@%s> a campsite.",
		"Tis better to have schniffed and lost than never to have schniffed at all. But <@%s> didn't lose, I just found them a campsite.",
		"What doesn't schniff you makes you stronger. <@%s> must be very weak since I just schniffed them a campsite.",
	}

	// Choose a random greeting template
	template := greetings[r.Intn(len(greetings))]
	return fmt.Sprintf(template, userID)

}
