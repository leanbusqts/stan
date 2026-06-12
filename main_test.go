package main

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"stan/calendar"
	"stan/internal"
	"stan/tasks"
)

func TestExtractGlobalFlags(t *testing.T) {
	args, output, err := extractGlobalFlags([]string{"calendar", "list", "--json", "--no-color"})
	if err != nil {
		t.Fatalf("extractGlobalFlags returned error: %v", err)
	}
	if output.JSON != true || output.NoColor != true || output.Quiet {
		t.Fatalf("unexpected output flags: %+v", output)
	}
	if got := strings.Join(args, " "); got != "calendar list" {
		t.Fatalf("filtered args mismatch: %q", got)
	}
}

func TestExtractGlobalFlagsRejectsConflicts(t *testing.T) {
	_, _, err := extractGlobalFlags([]string{"tasks", "list", "--json", "-q"})
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestVersionCommand(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run(t.Context(), []string{"version"}); err != nil {
			t.Fatalf("run version returned error: %v", err)
		}
	})
	if strings.TrimSpace(out) != currentVersion() {
		t.Fatalf("version output mismatch: %q", out)
	}
}

func TestRootHelpIncludesUtilityCommands(t *testing.T) {
	out := captureStdout(t, func() {
		printRootHelp()
	})
	for _, want := range []string{"stan version", "stan doctor", "stan help"} {
		if !strings.Contains(out, want) {
			t.Fatalf("root help missing %q: %q", want, out)
		}
	}
}

func TestPrintCalendarEventsAndTasks(t *testing.T) {
	calendarOut := captureStdout(t, func() {
		printCalendarEvents([]calendar.Event{
			{ID: "evt-1", Title: "Meeting", Start: time.Now(), End: time.Now().Add(time.Hour)},
		}, internal.OutputOptions{NoColor: true})
	})
	if !strings.Contains(calendarOut, "📅 Today") || !strings.Contains(calendarOut, "Meeting") {
		t.Fatalf("calendar output mismatch: %q", calendarOut)
	}

	tasksOut := captureStdout(t, func() {
		printTasks([]tasks.Task{
			{ID: "tsk-1", Title: "Buy milk"},
			{ID: "tsk-2", Title: "Pay bills", Completed: true, Depth: 1},
		}, tasks.DefaultListTitle, internal.OutputOptions{NoColor: true})
	})
	if !strings.Contains(tasksOut, "Tasks (Stan)") || !strings.Contains(tasksOut, "  [x] Pay bills") {
		t.Fatalf("tasks output mismatch: %q", tasksOut)
	}
}

func TestPrintCalendarEventsQuietMode(t *testing.T) {
	out := captureStdout(t, func() {
		printCalendarEvents([]calendar.Event{
			{ID: "evt-1", Title: "Meeting"},
		}, internal.OutputOptions{Quiet: true})
	})
	if strings.TrimSpace(out) != "evt-1 Meeting" {
		t.Fatalf("quiet calendar output mismatch: %q", out)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	raw, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	return string(raw)
}
