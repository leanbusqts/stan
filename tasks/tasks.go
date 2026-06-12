package tasks

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/option"
	gtasks "google.golang.org/api/tasks/v1"

	"stan/internal"
)

const DefaultListTitle = "Stan"

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
	Parent    string     `json:"parent,omitempty"`
	Position  string     `json:"position,omitempty"`
	Depth     int        `json:"depth,omitempty"`
}

// TaskDetail is a task plus its descendant subtasks.
type TaskDetail struct {
	Task     Task         `json:"task"`
	Subtasks []TaskDetail `json:"subtasks,omitempty"`
}

// JSONTask is the public JSON shape for task output.
type JSONTask struct {
	Title    string     `json:"title"`
	Notes    string     `json:"notes"`
	Subtasks []JSONTask `json:"subtasks,omitempty"`
}

// AddOptions contains supported flags for task creation.
type AddOptions struct {
	Title    string
	ListName string
	Due      string
	Notes    string
	Now      time.Time
}

// ShowOptions contains supported flags for task detail lookup.
type ShowOptions struct {
	Query    string
	ListName string
}

// ListOptions contains supported flags for task listing.
type ListOptions struct {
	IncludeSubtasks bool
}

// List returns tasks from the default Stan task list.
func List(ctx context.Context, client *http.Client, opts ListOptions) ([]Task, error) {
	listID, err := resolveListID(ctx, client, DefaultListTitle)
	if err != nil {
		return nil, err
	}
	items, err := listByID(ctx, client, listID)
	if err != nil {
		return nil, err
	}
	if opts.IncludeSubtasks {
		return items, nil
	}
	return topLevelTasks(items), nil
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

	listName := DefaultListTitle
	if opts.ListName != "" {
		listName = opts.ListName
	}
	listID, err := resolveListID(ctx, client, listName)
	if err != nil {
		return nil, err
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

	matches := matchingTaskLists(lists, title)
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("task list %q not found", title)
	case 1:
		return matches[0].ID, nil
	default:
		return "", fmt.Errorf("multiple task lists named %q; use a unique list title", title)
	}
}

func matchingTaskLists(lists []TaskList, title string) []TaskList {
	matches := make([]TaskList, 0)
	for _, item := range lists {
		if strings.EqualFold(item.Title, title) {
			matches = append(matches, item)
		}
	}
	return matches
}

func listByID(ctx context.Context, client *http.Client, listID string) ([]Task, error) {
	svc, err := gtasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	out := make([]Task, 0)
	pageToken := ""
	for {
		var response *gtasks.Tasks
		if err := withRetry(func() error {
			call := svc.Tasks.List(listID).ShowCompleted(true).ShowHidden(false).MaxResults(100)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			var inner error
			response, inner = call.Do()
			return inner
		}); err != nil {
			return nil, classifyGoogleError(err)
		}

		for _, item := range response.Items {
			task, err := fromGoogleTask(item)
			if err != nil {
				return nil, err
			}
			out = append(out, *task)
		}
		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}
	return orderTasks(out), nil
}

// Show returns one task and all of its descendant subtasks.
func Show(ctx context.Context, client *http.Client, opts ShowOptions) (*TaskDetail, error) {
	if strings.TrimSpace(opts.Query) == "" {
		return nil, fmt.Errorf("task query is required")
	}

	listName := DefaultListTitle
	if opts.ListName != "" {
		listName = opts.ListName
	}
	listID, err := resolveListID(ctx, client, listName)
	if err != nil {
		return nil, err
	}

	items, err := listByID(ctx, client, listID)
	if err != nil {
		return nil, err
	}

	root, err := findTask(items, opts.Query)
	if err != nil {
		return nil, err
	}
	detail := buildTaskDetail(items, root.ID)
	return &detail, nil
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
		Parent:    item.Parent,
		Position:  item.Position,
	}, nil
}

func orderTasks(items []Task) []Task {
	byParent := make(map[string][]Task)
	known := make(map[string]bool)
	for _, item := range items {
		known[item.ID] = true
	}
	for _, item := range items {
		parent := item.Parent
		if parent != "" && !known[parent] {
			parent = ""
		}
		byParent[parent] = append(byParent[parent], item)
	}
	for parent := range byParent {
		sort.SliceStable(byParent[parent], func(i, j int) bool {
			left := byParent[parent][i]
			right := byParent[parent][j]
			if left.Position != right.Position {
				return left.Position < right.Position
			}
			if left.Title != right.Title {
				return left.Title < right.Title
			}
			return left.ID < right.ID
		})
	}

	ordered := make([]Task, 0, len(items))
	visited := make(map[string]bool)
	var appendTree func(parent string, depth int)
	appendTree = func(parent string, depth int) {
		for _, item := range byParent[parent] {
			if visited[item.ID] {
				continue
			}
			visited[item.ID] = true
			item.Depth = depth
			ordered = append(ordered, item)
			appendTree(item.ID, depth+1)
		}
	}
	appendTree("", 0)
	for _, item := range items {
		if visited[item.ID] {
			continue
		}
		visited[item.ID] = true
		ordered = append(ordered, item)
	}
	return ordered
}

func topLevelTasks(items []Task) []Task {
	out := make([]Task, 0, len(items))
	for _, item := range items {
		if item.Depth == 0 {
			out = append(out, item)
		}
	}
	return out
}

func findTask(items []Task, query string) (Task, error) {
	query = strings.TrimSpace(query)
	for _, item := range items {
		if item.ID == query {
			return item, nil
		}
	}

	matches := make([]Task, 0)
	for _, item := range items {
		if strings.EqualFold(item.Title, query) {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return Task{}, fmt.Errorf("task %q not found", query)
	case 1:
		return matches[0], nil
	default:
		return Task{}, fmt.Errorf("multiple tasks named %q; use the task id", query)
	}
}

func buildTaskDetail(items []Task, rootID string) TaskDetail {
	byParent := make(map[string][]Task)
	byID := make(map[string]Task)
	for _, item := range items {
		byParent[item.Parent] = append(byParent[item.Parent], item)
		byID[item.ID] = item
	}

	var build func(Task, int) TaskDetail
	build = func(item Task, depth int) TaskDetail {
		item.Depth = depth
		detail := TaskDetail{Task: item}
		for _, child := range byParent[item.ID] {
			detail.Subtasks = append(detail.Subtasks, build(child, depth+1))
		}
		return detail
	}

	return build(byID[rootID], 0)
}

// ToJSONTasks converts internal task records to the public JSON shape.
func ToJSONTasks(items []Task) []JSONTask {
	out := make([]JSONTask, 0, len(items))
	for _, item := range items {
		out = append(out, JSONTask{
			Title: item.Title,
			Notes: item.Notes,
		})
	}
	return out
}

// ToJSONTaskDetail converts a task tree to the public JSON shape.
func ToJSONTaskDetail(detail *TaskDetail) JSONTask {
	out := JSONTask{
		Title: detail.Task.Title,
		Notes: detail.Task.Notes,
	}
	for _, subtask := range detail.Subtasks {
		child := subtask
		out.Subtasks = append(out.Subtasks, ToJSONTaskDetail(&child))
	}
	return out
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
