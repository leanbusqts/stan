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
