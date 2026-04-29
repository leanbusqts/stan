package tasks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/option"
	gtasks "google.golang.org/api/tasks/v1"

	"stan/internal"
)

const defaultListID = "@default"

// TaskList is the structured Stan representation of a Google task list.
type TaskList struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Task is the structured Stan representation of a Google task.
type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Notes     string     `json:"notes,omitempty"`
	Completed bool       `json:"completed"`
	Due       *time.Time `json:"due,omitempty"`
}

// AddOptions contains supported flags for task creation.
type AddOptions struct {
	Title    string
	ListName string
	Due      string
	Notes    string
	Now      time.Time
}

// List returns tasks from the default task list.
func List(ctx context.Context, client *http.Client) ([]Task, error) {
	return listByID(ctx, client, defaultListID)
}

// ListTaskLists returns available Google task lists.
func ListTaskLists(ctx context.Context, client *http.Client) ([]TaskList, error) {
	svc, err := gtasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	var response *gtasks.TaskLists
	if err := withRetry(func() error {
		var inner error
		response, inner = svc.Tasklists.List().MaxResults(100).Do()
		return inner
	}); err != nil {
		return nil, classifyGoogleError(err)
	}

	out := make([]TaskList, 0, len(response.Items))
	for _, item := range response.Items {
		out = append(out, TaskList{ID: item.Id, Title: item.Title})
	}
	return out, nil
}

// Add creates a task in the default or named task list.
func Add(ctx context.Context, client *http.Client, opts AddOptions) (*Task, error) {
	svc, err := gtasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	listID := defaultListID
	if opts.ListName != "" {
		listID, err = resolveListID(ctx, client, opts.ListName)
		if err != nil {
			return nil, err
		}
	}

	item := &gtasks.Task{
		Title: opts.Title,
		Notes: opts.Notes,
	}
	if opts.Due != "" {
		due, err := internal.ParseDateOnly(opts.Due, opts.Now.Location())
		if err != nil {
			return nil, err
		}
		item.Due = due.Format(time.RFC3339)
	}

	var created *gtasks.Task
	if err := withRetry(func() error {
		var inner error
		created, inner = svc.Tasks.Insert(listID, item).Do()
		return inner
	}); err != nil {
		return nil, classifyGoogleError(err)
	}
	return fromGoogleTask(created)
}

func resolveListID(ctx context.Context, client *http.Client, title string) (string, error) {
	lists, err := ListTaskLists(ctx, client)
	if err != nil {
		return "", err
	}

	matches := make([]TaskList, 0)
	for _, item := range lists {
		if item.Title == title {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("task list %q not found", title)
	case 1:
		return matches[0].ID, nil
	default:
		return "", fmt.Errorf("multiple task lists named %q; use a unique list title", title)
	}
}

func listByID(ctx context.Context, client *http.Client, listID string) ([]Task, error) {
	svc, err := gtasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	var response *gtasks.Tasks
	if err := withRetry(func() error {
		var inner error
		response, inner = svc.Tasks.List(listID).ShowCompleted(true).ShowHidden(false).Do()
		return inner
	}); err != nil {
		return nil, classifyGoogleError(err)
	}

	out := make([]Task, 0, len(response.Items))
	for _, item := range response.Items {
		task, err := fromGoogleTask(item)
		if err != nil {
			return nil, err
		}
		out = append(out, *task)
	}
	return out, nil
}

func fromGoogleTask(item *gtasks.Task) (*Task, error) {
	var due *time.Time
	if item.Due != "" {
		parsed, err := time.Parse(time.RFC3339, item.Due)
		if err != nil {
			return nil, err
		}
		due = &parsed
	}
	return &Task{
		ID:        item.Id,
		Title:     item.Title,
		Notes:     item.Notes,
		Completed: strings.EqualFold(item.Status, "completed"),
		Due:       due,
	}, nil
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
	if strings.Contains(text, "invalid_grant") || strings.Contains(text, "401") || strings.Contains(text, "400") {
		return fmt.Errorf("Session expired or revoked.\nRun: stan auth login")
	}
	return err
}
