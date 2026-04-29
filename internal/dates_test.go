package internal

import (
	"testing"
	"time"
)

func TestParseWhenSupportsTimeOnly(t *testing.T) {
	now := time.Date(2026, 4, 19, 8, 30, 0, 0, time.FixedZone("ART", -3*60*60))

	got, err := ParseWhen("10:45", now)
	if err != nil {
		t.Fatalf("ParseWhen returned error: %v", err)
	}

	want := time.Date(2026, 4, 19, 10, 45, 0, 0, now.Location())
	if !got.Equal(want) {
		t.Fatalf("ParseWhen mismatch: got %v want %v", got, want)
	}
}

func TestParseWhenSupportsDateOnlyAtNineAM(t *testing.T) {
	now := time.Date(2026, 4, 19, 8, 30, 0, 0, time.UTC)

	got, err := ParseWhen("2026-05-01", now)
	if err != nil {
		t.Fatalf("ParseWhen returned error: %v", err)
	}

	want := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("ParseWhen mismatch: got %v want %v", got, want)
	}
}

func TestParseWhenRejectsInvalidFormat(t *testing.T) {
	_, err := ParseWhen("tomorrow morning", time.Now())
	if err == nil {
		t.Fatal("ParseWhen unexpectedly succeeded")
	}
}

func TestParseDateOnly(t *testing.T) {
	got, err := ParseDateOnly("2026-04-22", time.UTC)
	if err != nil {
		t.Fatalf("ParseDateOnly returned error: %v", err)
	}

	want := time.Date(2026, 4, 22, 9, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("ParseDateOnly mismatch: got %v want %v", got, want)
	}
}

func TestHumanDayLabel(t *testing.T) {
	now := time.Date(2026, 4, 19, 8, 0, 0, 0, time.UTC)

	if got := HumanDayLabel(now, now); got != "Today" {
		t.Fatalf("HumanDayLabel Today mismatch: %q", got)
	}
	if got := HumanDayLabel(now.AddDate(0, 0, 1), now); got != "Tomorrow" {
		t.Fatalf("HumanDayLabel Tomorrow mismatch: %q", got)
	}
	if got := HumanDayLabel(now.AddDate(0, 0, 2), now); got != "Tue 2026-04-21" {
		t.Fatalf("HumanDayLabel future mismatch: %q", got)
	}
}
