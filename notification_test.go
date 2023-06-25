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

	sc := NewSchniffCollection("schniffs.json")

	fmt.Print(GenerateDiscordMessage(sc, notification))
}
