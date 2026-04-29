package calendar

import (
	"errors"
	"testing"
	"time"

	gcalendar "google.golang.org/api/calendar/v3"
)

func TestParseListOptionsDefaultRange(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

	opts, err := ParseListOptions(7, "", "", now)
	if err != nil {
		t.Fatalf("ParseListOptions returned error: %v", err)
	}
	if !opts.Start.Equal(now) {
		t.Fatalf("start mismatch: got %v want %v", opts.Start, now)
	}
	if got, want := opts.End, now.Add(7*24*time.Hour); !got.Equal(want) {
		t.Fatalf("end mismatch: got %v want %v", got, want)
	}
}

func TestParseListOptionsStartAndEnd(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

	opts, err := ParseListOptions(7, "2026-04-20", "2026-04-22", now)
	if err != nil {
		t.Fatalf("ParseListOptions returned error: %v", err)
	}
	if got, want := opts.Start, time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("start mismatch: got %v want %v", got, want)
	}
	if got, want := opts.End, time.Date(2026, 4, 23, 9, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("end mismatch: got %v want %v", got, want)
	}
}

func TestParseListOptionsRejectsInvalidRanges(t *testing.T) {
	now := time.Now()

	if _, err := ParseListOptions(7, "", "2026-04-22", now); err == nil {
		t.Fatal("expected error when end is provided without start")
	}
	if _, err := ParseListOptions(7, "2026-04-22", "2026-04-20", now); err == nil {
		t.Fatal("expected error for inverted range")
	}
}

func TestParseEventTime(t *testing.T) {
	got, err := parseEventTime(&gcalendar.EventDateTime{DateTime: "2026-04-19T10:00:00Z"})
	if err != nil {
		t.Fatalf("parseEventTime datetime returned error: %v", err)
	}
	if want := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("datetime mismatch: got %v want %v", got, want)
	}

	got, err = parseEventTime(&gcalendar.EventDateTime{Date: "2026-04-19"})
	if err != nil {
		t.Fatalf("parseEventTime date returned error: %v", err)
	}
	if got.Year() != 2026 || got.Month() != 4 || got.Day() != 19 {
		t.Fatalf("date mismatch: got %v", got)
	}
}

func TestWithRetryAndClassifyGoogleError(t *testing.T) {
	attempts := 0
	err := withRetry(func() error {
		attempts++
		if attempts == 1 {
			return errors.New("temporary failure")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("withRetry returned error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("withRetry attempts mismatch: got %d want %d", attempts, 2)
	}

	got := classifyGoogleError(errors.New("oauth invalid_grant"))
	if got == nil || got.Error() != "Session expired or revoked.\nRun: stan auth login" {
		t.Fatalf("classifyGoogleError mismatch: %v", got)
	}
}
