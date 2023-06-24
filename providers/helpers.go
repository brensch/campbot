package providers

import "time"

func GetStartOfMonth(input time.Time) time.Time {
	return time.Date(input.Year(), input.Month(), 1, 0, 0, 0, 0, time.UTC)
}
