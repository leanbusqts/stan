package calendar

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	gcalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"stan/internal"
)

// ListOptions controls the calendar list range.
type ListOptions struct {
	Start time.Time
	End   time.Time
}

// Event is the structured Stan representation of a calendar event.
type Event struct {
	ID    string    `json:"id"`
	Title string    `json:"title"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Link  string    `json:"html_link,omitempty"`
}

// ParseListOptions resolves the supported calendar range flags.
func ParseListOptions(days int, start string, end string, now time.Time) (ListOptions, error) {
	loc := now.Location()
	opts := ListOptions{}

	switch {
	case start == "" && end == "":
		if days <= 0 {
			days = 7
		}
		opts.Start = now
		opts.End = now.Add(time.Duration(days) * 24 * time.Hour)
	case start != "" && end == "":
		startTime, err := internal.ParseDateOnly(start, loc)
		if err != nil {
			return ListOptions{}, err
		}
		opts.Start = startTime
		opts.End = startTime.Add(time.Duration(days) * 24 * time.Hour)
	case start == "" && end != "":
		return ListOptions{}, fmt.Errorf("--end requires --start")
	default:
		startTime, err := internal.ParseDateOnly(start, loc)
		if err != nil {
			return ListOptions{}, err
		}
		endTime, err := internal.ParseDateOnly(end, loc)
		if err != nil {
			return ListOptions{}, err
		}
		opts.Start = startTime
		opts.End = endTime.Add(24 * time.Hour)
	}

	if !opts.End.After(opts.Start) {
		return ListOptions{}, fmt.Errorf("end must be after start")
	}
	return opts, nil
}

// List returns primary calendar events for the requested time range.
func List(ctx context.Context, client *http.Client, opts ListOptions) ([]Event, error) {
	svc, err := gcalendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	call := svc.Events.List("primary").
		SingleEvents(true).
		OrderBy("startTime").
		TimeMin(opts.Start.Format(time.RFC3339)).
		TimeMax(opts.End.Format(time.RFC3339))

	var response *gcalendar.Events
	if err := withRetry(func() error {
		var inner error
		response, inner = call.Do()
		return inner
	}); err != nil {
		return nil, classifyGoogleError(err)
	}

	events := make([]Event, 0, len(response.Items))
	for _, item := range response.Items {
		start, err := parseEventTime(item.Start)
		if err != nil {
			return nil, err
		}
		end, err := parseEventTime(item.End)
		if err != nil {
			return nil, err
		}
		events = append(events, Event{
			ID:    item.Id,
			Title: item.Summary,
			Start: start,
			End:   end,
			Link:  item.HtmlLink,
		})
	}
	return events, nil
}

// Add creates a calendar event in the primary calendar.
func Add(ctx context.Context, client *http.Client, title string, when string, duration time.Duration, end string, now time.Time) (*Event, error) {
	start, err := internal.ParseWhen(when, now)
	if err != nil {
		return nil, err
	}
	finish := start.Add(duration)
	if end != "" {
		finish, err = internal.ParseWhen(end, now)
		if err != nil {
			return nil, err
		}
	}
	if !finish.After(start) {
		return nil, fmt.Errorf("end must be after start")
	}

	svc, err := gcalendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	raw := &gcalendar.Event{
		Summary: title,
		Start:   &gcalendar.EventDateTime{DateTime: start.Format(time.RFC3339)},
		End:     &gcalendar.EventDateTime{DateTime: finish.Format(time.RFC3339)},
	}

	var created *gcalendar.Event
	if err := withRetry(func() error {
		var inner error
		created, inner = svc.Events.Insert("primary", raw).Do()
		return inner
	}); err != nil {
		return nil, classifyGoogleError(err)
	}

	return &Event{
		ID:    created.Id,
		Title: created.Summary,
		Start: start,
		End:   finish,
		Link:  created.HtmlLink,
	}, nil
}

func parseEventTime(value *gcalendar.EventDateTime) (time.Time, error) {
	if value == nil {
		return time.Time{}, fmt.Errorf("event time missing")
	}
	if value.DateTime != "" {
		return time.Parse(time.RFC3339, value.DateTime)
	}
	if value.Date != "" {
		return time.ParseInLocation("2006-01-02", value.Date, time.Local)
	}
	return time.Time{}, fmt.Errorf("event time missing")
}

func withRetry(fn func() error) error {
	var err error
	wait := 200 * time.Millisecond
	for i := 0; i < 2; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(wait)
		wait *= 2
	}
	return err
}

func classifyGoogleError(err error) error {
	text := err.Error()
	if containsAny(text, "invalid_grant", "401", "400") {
		return fmt.Errorf("Session expired or revoked.\nRun: stan auth login")
	}
	return err
}

func containsAny(text string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(text, part) {
			return true
		}
	}
	return false
}
