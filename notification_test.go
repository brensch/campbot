package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestGenerateNotificationMessage(t *testing.T) {
	notificationBytes, err := os.ReadFile("notification.json")
	if err != nil {
		t.Error(err)
		return
	}

	var notification Notification
	err = json.Unmarshal(notificationBytes, &notification)
	if err != nil {
		t.Error(err)
		return
	}

	sc := NewSchniffCollection("../example_schniffs.json")

	fmt.Print(GenerateDiscordMessage(sc, notification))
}

func TestGenerateNotificationMessageEmbed(t *testing.T) {
	notificationBytes, err := os.ReadFile("notification.json")
	if err != nil {
		t.Error(err)
		return
	}

	var notification Notification
	err = json.Unmarshal(notificationBytes, &notification)
	if err != nil {
		t.Error(err)
		return
	}

	sc := NewSchniffCollection("../example_schniffs.json")

	message, err := GenerateDiscordMessageEmbed(sc, notification)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(message.Title)
	fmt.Println(message.Description)
	for _, field := range message.Fields {
		fmt.Println(field.Name)
		fmt.Println(field.Value)
	}

}
