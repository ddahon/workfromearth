package components

import (
	"fmt"
	"time"
)

// FormatRelativeDate formats a date string as a relative time (e.g., "3d ago", "2h ago")
// Returns empty string if the date cannot be parsed
func FormatRelativeDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC1123,
		time.RFC1123Z,
	}

	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return ""
	}

	duration := time.Since(t)

	if duration < 0 {
		return ""
	}

	hours := duration.Hours()
	days := hours / 24

	switch {
	case hours < 1:
		minutes := duration.Minutes()
		if minutes < 1 {
			return "just now"
		}
		return fmt.Sprintf("%.0fm ago", minutes)
	case hours < 24:
		return fmt.Sprintf("%.0fh ago", hours)
	default:
		return fmt.Sprintf("%.0fd ago", days)
	}
}
