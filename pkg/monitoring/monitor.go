package monitoring

import (
	"github.com/brensch/campbot/pkg/models"
	"github.com/brensch/campbot/pkg/providers"
)

type Monitor struct {
	providers []providers.Provider
}

// NewMonitor initializes a new Monitor with the given parameters
func NewMonitor(providers []providers.Provider) *Monitor {
	return &Monitor{
		providers: providers,
	}
}

// CheckAvailabilities checks the availability of campsites for all users
func (m *Monitor) CheckAvailabilities(users []models.User) []models.ChangeNotification {
	// 1. Iterate over all users
	// 2. For each user, iterate over their campsites and call the appropriate provider's CheckAvailability method
	// 3. Compare the new availability with the old availability (probably stored in some database)
	// 4. If there's a difference, create a new ChangeNotification and add it to a slice of ChangeNotifications
	// 5. Update the old availability in the database to the new availability
	// 6. Return the slice of ChangeNotifications

	var changes []models.ChangeNotification
	// ... Code to implement steps 1-6 above
	return changes
}
