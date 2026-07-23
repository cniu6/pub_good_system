package models_test

import (
	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"testing"
)

func TestCreateAPIAccessLogsSkipsReplayedRequestIDs(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	items := []*models.APIAccessLog{
		{RequestID: "batch-request-1", Method: "GET", Path: "/one", RoutePath: "/one", StatusCode: 200},
		{RequestID: "batch-request-2", Method: "POST", Path: "/two", RoutePath: "/two", StatusCode: 200},
	}
	created, err := models.CreateAPIAccessLogs(items)
	if err != nil {
		t.Fatalf("first CreateAPIAccessLogs failed: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("first created count = %d, want 2", len(created))
	}

	replayed, err := models.CreateAPIAccessLogs(items)
	if err != nil {
		t.Fatalf("replay CreateAPIAccessLogs failed: %v", err)
	}
	if len(replayed) != 0 {
		t.Fatalf("replayed created count = %d, want 0", len(replayed))
	}

	list, total, err := models.GetAPIAccessLogList(&models.APIAccessLogQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetAPIAccessLogList failed: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("log count = total:%d len:%d, want 2", total, len(list))
	}
}
