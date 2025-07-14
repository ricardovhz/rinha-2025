package internal

import (
	"fmt"
	"time"
)

func ParseDateTime(dateStr string) time.Time {
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02",
		time.RFC3339Nano,
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z07:00",
	}
	for _, f := range formats {
		t, err := time.Parse(f, dateStr)
		if err == nil {
			return t
		}
	}
	fmt.Printf("Failed to parse date %s with formats %v\n", dateStr, formats)
	return time.Time{} // Return zero value if no format matches
}
