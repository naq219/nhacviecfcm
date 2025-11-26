package utils

import "time"

func IsTimeValid(t time.Time) bool {
	return !t.IsZero() && t.Year() >= 2000
}

// func IsBeforeNow(t time.Time) bool {
// 	return !t.IsZero() && (t.Before(time.Now()) || t.Equal(time.Now()))
// }
