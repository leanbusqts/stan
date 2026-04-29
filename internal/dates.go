package internal

import (
	"fmt"
	"time"
)

const (
	dateLayout     = "2006-01-02"
	timeOnlyLayout = "15:04"
	dateTimeLayout = "2006-01-02T15:04"
)

// ParseWhen parses the supported Stan date and time inputs in local time.
func ParseWhen(input string, now time.Time) (time.Time, error) {
	loc := now.Location()

	if t, err := time.ParseInLocation(time.RFC3339, input, loc); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation(dateTimeLayout, input, loc); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation(dateLayout, input, loc); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, loc), nil
	}
	if t, err := time.ParseInLocation(timeOnlyLayout, input, loc); err == nil {
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
	}

	return time.Time{}, fmt.Errorf("invalid date/time %q; supported formats: 15:04, 2006-01-02, 2006-01-02T15:04, RFC3339", input)
}

// ParseDateOnly parses a YYYY-MM-DD value in local time at 09:00.
func ParseDateOnly(input string, loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation(dateLayout, input, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q; expected YYYY-MM-DD", input)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, loc), nil
}

// HumanDayLabel formats a date header for calendar output.
func HumanDayLabel(day time.Time, now time.Time) string {
	today := truncateDay(now)
	target := truncateDay(day)
	switch {
	case target.Equal(today):
		return "Today"
	case target.Equal(today.AddDate(0, 0, 1)):
		return "Tomorrow"
	default:
		return target.Format("Mon 2006-01-02")
	}
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
