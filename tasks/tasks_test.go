package tasks

import (
	"errors"
	"testing"
	"time"

	gtasks "google.golang.org/api/tasks/v1"
)

func TestFromGoogleTask(t *testing.T) {
	item, err := fromGoogleTask(&gtasks.Task{
		Id:     "task-1",
		Title:  "Buy milk",
		Notes:  "semi skimmed",
		Status: "completed",
		Due:    "2026-04-20T09:00:00Z",
	})
	if err != nil {
		t.Fatalf("fromGoogleTask returned error: %v", err)
	}
	if item.ID != "task-1" || item.Title != "Buy milk" || item.Notes != "semi skimmed" {
		t.Fatalf("task mapping mismatch: %+v", item)
	}
	if !item.Completed {
		t.Fatalf("expected completed task: %+v", item)
	}
	if item.Due == nil || !item.Due.Equal(time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)) {
		t.Fatalf("task due mismatch: %+v", item.Due)
	}
}

func TestFromGoogleTaskRejectsInvalidDue(t *testing.T) {
	_, err := fromGoogleTask(&gtasks.Task{Due: "not-a-date"})
	if err == nil {
		t.Fatal("expected parsing error for invalid due date")
	}
}

func TestMatchingTaskListsIsCaseInsensitive(t *testing.T) {
	lists := []TaskList{
		{ID: "1", Title: "Inbox"},
		{ID: "2", Title: "sTaN"},
		{ID: "3", Title: "Work"},
	}

	matches := matchingTaskLists(lists, "Stan")
	if len(matches) != 1 || matches[0].ID != "2" {
		t.Fatalf("matchingTaskLists mismatch: %+v", matches)
	}
}

func TestMatchingTaskListsReturnsAllCaseInsensitiveMatches(t *testing.T) {
	lists := []TaskList{
		{ID: "1", Title: "Stan"},
		{ID: "2", Title: "stan"},
	}

	matches := matchingTaskLists(lists, "STAN")
	if len(matches) != 2 {
		t.Fatalf("matchingTaskLists count mismatch: %+v", matches)
	}
}

func TestOrderTasksUsesHierarchyAndPosition(t *testing.T) {
	items := []Task{
		{ID: "child-2", Title: "Child 2", Parent: "parent", Position: "00000000000000000002"},
		{ID: "root-2", Title: "Root 2", Position: "00000000000000000002"},
		{ID: "child-1", Title: "Child 1", Parent: "parent", Position: "00000000000000000001"},
		{ID: "parent", Title: "Parent", Position: "00000000000000000001"},
	}

	ordered := orderTasks(items)
	got := make([]string, 0, len(ordered))
	depths := make([]int, 0, len(ordered))
	for _, item := range ordered {
		got = append(got, item.ID)
		depths = append(depths, item.Depth)
	}

	want := []string{"parent", "child-1", "child-2", "root-2"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order mismatch: got %v want %v", got, want)
		}
	}
	if depths[0] != 0 || depths[1] != 1 || depths[2] != 1 || depths[3] != 0 {
		t.Fatalf("depth mismatch: %+v", ordered)
	}
}

func TestOrderTasksTreatsMissingParentsAsRoots(t *testing.T) {
	ordered := orderTasks([]Task{
		{ID: "orphan", Title: "Orphan", Parent: "missing", Position: "2"},
		{ID: "root", Title: "Root", Position: "1"},
	})

	got := []string{ordered[0].ID, ordered[1].ID}
	want := []string{"root", "orphan"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order mismatch: got %v want %v", got, want)
		}
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

	got := classifyGoogleError(errors.New("401 unauthorized"))
	if got == nil || got.Error() != "Session expired or revoked.\nRun: stan auth login" {
		t.Fatalf("classifyGoogleError mismatch: %v", got)
	}
}
