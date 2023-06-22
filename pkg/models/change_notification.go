package models

import "time"

type ChangeNotification struct {
	UserID    string
	Campsite  Campsite
	OldStatus Availability
	NewStatus Availability
	SentAt    time.Time
}

func NewChangeNotification(userID string, campsite Campsite, oldStatus, newStatus Availability, sentAt time.Time) ChangeNotification {
	return ChangeNotification{
		UserID:    userID,
		Campsite:  campsite,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		SentAt:    sentAt,
	}
}

func (n *ChangeNotification) IsSentSuccessfully() bool {
	return !n.SentAt.IsZero()
}
