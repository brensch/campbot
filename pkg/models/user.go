package models

type User struct {
	ID                string
	Email             string
	NotificationTypes []string // could be "email", "sms", etc.
	Campsites         []string // list of campsites that the user wants to monitor
}

func NewUser(id string, email string, notificationTypes []string, campsites []string) User {
	return User{
		ID:                id,
		Email:             email,
		NotificationTypes: notificationTypes,
		Campsites:         campsites,
	}
}

func (u *User) AddCampsite(campsite string) {
	u.Campsites = append(u.Campsites, campsite)
}

func (u *User) RemoveCampsite(campsite string) {
	for i, site := range u.Campsites {
		if site == campsite {
			u.Campsites = append(u.Campsites[:i], u.Campsites[i+1:]...)
			break
		}
	}
}

func (u *User) AddNotificationType(notificationType string) {
	u.NotificationTypes = append(u.NotificationTypes, notificationType)
}

func (u *User) RemoveNotificationType(notificationType string) {
	for i, nType := range u.NotificationTypes {
		if nType == notificationType {
			u.NotificationTypes = append(u.NotificationTypes[:i], u.NotificationTypes[i+1:]...)
			break
		}
	}
}
