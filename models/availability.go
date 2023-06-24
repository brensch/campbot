package models

import "time"

type Availability struct {
	Date     time.Time
	Reserved bool
	SiteID   string
}

// Equal checks if the current availability is equal to another one.
func (a *Availability) Equal(other Availability) bool {
	return a.Date.Equal(other.Date) && a.Reserved == other.Reserved && a.SiteID == other.SiteID
}

// IsAvailable checks if a site is available.
func (a *Availability) IsAvailable() bool {
	return !a.Reserved
}

// NewAvailability is a constructor for the Availability struct
func NewAvailability(date time.Time, reserved bool, siteID string) Availability {
	return Availability{
		Date:     date,
		Reserved: reserved,
		SiteID:   siteID,
	}
}
