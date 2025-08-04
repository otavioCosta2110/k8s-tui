package utils

import (
	"fmt"
	"time"
)

func FormatTime(t time.Time) string {
	duration := time.Since(t)
	seconds := int(duration.Seconds())

	switch {
	case seconds < 60:
		return fmt.Sprintf("%ds", seconds)
	case seconds < 3600:
		minutes := seconds / 60
		remainingSeconds := seconds % 60
		return fmt.Sprintf("%dm%ds", minutes, remainingSeconds)
	case seconds < 86400:
		hours := seconds / 3600
		remainingMinutes := (seconds % 3600) / 60
		return fmt.Sprintf("%dh%dm", hours, remainingMinutes)
	case seconds < 31536000: 
		days := seconds / 86400
		remainingHours := (seconds % 86400) / 3600
		return fmt.Sprintf("%dd%dh", days, remainingHours)
	default:
		years := seconds / 31536000
		remainingDays := (seconds % 31536000) / 86400
		if remainingDays > 30 {
			months := remainingDays / 30
			return fmt.Sprintf("%dy%dmo", years, months)
		}
		return fmt.Sprintf("%dy", years)
	}
}

