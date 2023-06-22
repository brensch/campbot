package notifications

import (
	"fmt"
	"net/smtp"

	"github.com/brensch/campbot/pkg/models"
)

type Notifier struct {
	smtpHost  string
	smtpPort  string
	emailUser string
	emailPass string
}

func NewNotifier(smtpHost string, smtpPort string, emailUser string, emailPass string) *Notifier {
	return &Notifier{
		smtpHost:  smtpHost,
		smtpPort:  smtpPort,
		emailUser: emailUser,
		emailPass: emailPass,
	}
}

func (n *Notifier) SendNotifications(notifications []models.ChangeNotification) error {
	for _, notification := range notifications {
		// this needs to be changed to be email i assume
		user := notification.UserID
		availability := notification.NewStatus
		msg := fmt.Sprintf("Subject: Campsite Availability Changed\n\nThe availability of campsite %s has changed. It is now: %t.", availability.Site.SiteID, availability.Reserved)

		err := smtp.SendMail(
			n.smtpHost+":"+n.smtpPort,
			smtp.PlainAuth("", n.emailUser, n.emailPass, n.smtpHost),
			n.emailUser,
			[]string{user},
			[]byte(msg),
		)

		if err != nil {
			return err
		}
	}

	return nil
}
