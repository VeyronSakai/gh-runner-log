package usecase

import (
	"fmt"
	"time"
)

// ParseSince parses the --since flag value and returns the corresponding time
// Supports formats like "24h", "2d", "1w" or RFC3339 timestamps
func ParseSince(since string) (time.Time, error) {
	if since == "" {
		// Default to 24 hours ago
		return time.Now().Add(-24 * time.Hour), nil
	}

	// Try parsing as duration (e.g., "24h", "2d", "1w")
	if duration, err := parseDuration(since); err == nil {
		return time.Now().Add(-duration), nil
	}

	// Try parsing as RFC3339 timestamp
	if t, err := time.Parse(time.RFC3339, since); err == nil {
		return t, nil
	}

	// Try parsing as date only (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", since); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s (expected format: duration like '24h', '2d', '1w' or date like '2024-01-01')", since)
}

// parseDuration extends time.ParseDuration to support days (d) and weeks (w)
func parseDuration(s string) (time.Duration, error) {
	// Try standard Go duration first
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Handle days (d) and weeks (w)
	var value int
	var unit string
	if n, err := fmt.Sscanf(s, "%d%s", &value, &unit); err == nil && n == 2 {
		switch unit {
		case "d":
			return time.Duration(value) * 24 * time.Hour, nil
		case "w":
			return time.Duration(value) * 7 * 24 * time.Hour, nil
		}
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
}
