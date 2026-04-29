package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// OutputOptions are the supported global CLI output flags.
type OutputOptions struct {
	JSON    bool
	Quiet   bool
	NoColor bool
}

// PrintJSON writes structured JSON without additional formatting noise.
func PrintJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

// Colorize wraps text in an ANSI color code when enabled.
func Colorize(enabled bool, code string, text string) string {
	if !enabled {
		return text
	}
	return code + text + "\033[0m"
}

// CalendarLineColor returns a color based on event proximity.
func CalendarLineColor(enabled bool, when time.Time, now time.Time, text string) string {
	switch {
	case sameDay(when, now):
		return Colorize(enabled, "\033[32m", text)
	case when.Before(now):
		return Colorize(enabled, "\033[31m", text)
	case when.Before(now.Add(48 * time.Hour)):
		return Colorize(enabled, "\033[33m", text)
	default:
		return text
	}
}

// TaskLineColor returns a color based on task status and due date.
func TaskLineColor(enabled bool, completed bool, due *time.Time, now time.Time, text string) string {
	switch {
	case completed:
		return Colorize(enabled, "\033[32m", text)
	case due != nil && truncateDay(*due).Before(truncateDay(now)):
		return Colorize(enabled, "\033[31m", text)
	case due != nil && sameDay(*due, now):
		return Colorize(enabled, "\033[32m", text)
	case due != nil:
		return Colorize(enabled, "\033[33m", text)
	default:
		return text
	}
}

// CompactLine trims blank values for quiet output.
func CompactLine(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, " ")
}

// RelativeExpiry formats a token expiry in coarse human-readable units.
func RelativeExpiry(expiry time.Time, now time.Time) string {
	if expiry.IsZero() {
		return "unknown"
	}
	delta := expiry.Sub(now).Round(time.Minute)
	if delta < 0 {
		return "expired"
	}
	if delta < time.Hour {
		return fmt.Sprintf("%d minutes", int(delta.Minutes()))
	}
	return fmt.Sprintf("%.1f hours", delta.Hours())
}

func sameDay(a time.Time, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
