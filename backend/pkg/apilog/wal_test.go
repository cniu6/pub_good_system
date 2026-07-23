package apilog

import (
	"context"
	"errors"
	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"testing"
	"time"
)

func TestWALReplayRetainsFileAfterFailure(t *testing.T) {
	wal, err := NewWAL(t.TempDir())
	if err != nil {
		t.Fatalf("NewWAL failed: %v", err)
	}
	if err := wal.Append(&models.APIAccessLog{RequestID: "wal-retry-1", Method: "GET"}); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	if err := wal.Replay(100, func([]*models.APIAccessLog) error {
		return errors.New("database unavailable")
	}); err == nil {
		t.Fatal("Replay should return handler error")
	}

	var replayed []string
	if err := wal.Replay(100, func(items []*models.APIAccessLog) error {
		for _, item := range items {
			replayed = append(replayed, item.RequestID)
		}
		return nil
	}); err != nil {
		t.Fatalf("Replay after recovery failed: %v", err)
	}
	if len(replayed) != 1 || replayed[0] != "wal-retry-1" {
		t.Fatalf("replayed = %#v, want [wal-retry-1]", replayed)
	}
}

func TestWriterQueueOverflowFallsBackToWAL(t *testing.T) {
	writer, err := NewWriter(Options{
		QueueCapacity: 1,
		QueueMaxBytes: 1 << 20,
		BatchSize:     100,
		WALDir:        t.TempDir(),
	})
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	if err := writer.Enqueue(&models.APIAccessLog{RequestID: "queued-1", Method: "GET"}); err != nil {
		t.Fatalf("first Enqueue failed: %v", err)
	}
	if err := writer.Enqueue(&models.APIAccessLog{RequestID: "wal-overflow-2", Method: "GET"}); err != nil {
		t.Fatalf("overflow Enqueue failed: %v", err)
	}

	var replayed []string
	if err := writer.wal.Replay(100, func(items []*models.APIAccessLog) error {
		for _, item := range items {
			replayed = append(replayed, item.RequestID)
		}
		return nil
	}); err != nil {
		t.Fatalf("WAL Replay failed: %v", err)
	}
	if len(replayed) != 1 || replayed[0] != "wal-overflow-2" {
		t.Fatalf("replayed = %#v, want [wal-overflow-2]", replayed)
	}
}

func TestWriterFlushesBatchToDatabase(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	writer, err := NewWriter(Options{
		QueueCapacity: 10,
		QueueMaxBytes: 1 << 20,
		BatchSize:     2,
		FlushInterval: time.Second,
		WALDir:        t.TempDir(),
	})
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	go writer.loop()

	for _, requestID := range []string{"batch-flush-1", "batch-flush-2"} {
		if err := writer.Enqueue(&models.APIAccessLog{RequestID: requestID, Method: "GET", Path: "/test", RoutePath: "/test", StatusCode: 200}); err != nil {
			t.Fatalf("Enqueue(%s) failed: %v", requestID, err)
		}
	}

	deadline := time.Now().Add(time.Second)
	for {
		_, total, err := models.GetAPIAccessLogList(&models.APIAccessLogQuery{Page: 1, PageSize: 10})
		if err == nil && total == 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("batch was not flushed before deadline, total=%d err=%v", total, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := writer.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}
