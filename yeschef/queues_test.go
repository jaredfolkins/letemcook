package yeschef

import (
	"testing"
)

func TestNewQuartzQueue(t *testing.T) {
	q := NewQuartzQueue("test")
	if q.Name != "test" {
		t.Fatalf("unexpected name: %s", q.Name)
	}
}

func TestJobQueueSizeEmpty(t *testing.T) {
	q := &jobQueue{Path: t.TempDir(), Name: "temp"}
	if n, err := q.Size(); err != nil || n != 0 {
		t.Fatalf("expected empty queue, got n=%d err=%v", n, err)
	}
}
