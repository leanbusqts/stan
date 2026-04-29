package internal

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer

	err := PrintJSON(&buf, map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "\"hello\": \"world\"") {
		t.Fatalf("PrintJSON output mismatch: %q", got)
	}
}

func TestColorizeDisabled(t *testing.T) {
	if got := Colorize(false, "\033[31m", "value"); got != "value" {
		t.Fatalf("Colorize disabled mismatch: %q", got)
	}
}

func TestCalendarLineColor(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

	if got := CalendarLineColor(true, now, now, "today"); !strings.Contains(got, "\033[32m") {
		t.Fatalf("expected green line for today, got %q", got)
	}
	if got := CalendarLineColor(true, now.Add(24*time.Hour), now, "soon"); !strings.Contains(got, "\033[33m") {
		t.Fatalf("expected yellow line for upcoming, got %q", got)
	}
	if got := CalendarLineColor(true, now.Add(72*time.Hour), now, "later"); got != "later" {
		t.Fatalf("expected no color for distant event, got %q", got)
	}
}

func TestTaskLineColor(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	overdue := now.AddDate(0, 0, -1)
	today := now
	upcoming := now.AddDate(0, 0, 2)

	if got := TaskLineColor(true, true, nil, now, "done"); !strings.Contains(got, "\033[32m") {
		t.Fatalf("expected green line for completed task, got %q", got)
	}
	if got := TaskLineColor(true, false, &overdue, now, "late"); !strings.Contains(got, "\033[31m") {
		t.Fatalf("expected red line for overdue task, got %q", got)
	}
	if got := TaskLineColor(true, false, &today, now, "today"); !strings.Contains(got, "\033[32m") {
		t.Fatalf("expected green line for same-day task, got %q", got)
	}
	if got := TaskLineColor(true, false, &upcoming, now, "soon"); !strings.Contains(got, "\033[33m") {
		t.Fatalf("expected yellow line for upcoming task, got %q", got)
	}
}

func TestCompactLineAndRelativeExpiry(t *testing.T) {
	if got := CompactLine("id-1", "", "title"); got != "id-1 title" {
		t.Fatalf("CompactLine mismatch: %q", got)
	}

	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	if got := RelativeExpiry(time.Time{}, now); got != "unknown" {
		t.Fatalf("RelativeExpiry zero mismatch: %q", got)
	}
	if got := RelativeExpiry(now.Add(30*time.Minute), now); got != "30 minutes" {
		t.Fatalf("RelativeExpiry minutes mismatch: %q", got)
	}
	if got := RelativeExpiry(now.Add(90*time.Minute), now); got != "1.5 hours" {
		t.Fatalf("RelativeExpiry hours mismatch: %q", got)
	}
	if got := RelativeExpiry(now.Add(-time.Minute), now); got != "expired" {
		t.Fatalf("RelativeExpiry expired mismatch: %q", got)
	}
}
