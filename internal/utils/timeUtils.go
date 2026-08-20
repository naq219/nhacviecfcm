package utils

import "time"

func IsTimeValid(t time.Time) bool {
	return !t.IsZero() && t.Year() >= 2000
}
