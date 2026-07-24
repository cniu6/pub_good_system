package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"

	"github.com/gin-gonic/gin"
)

func TestIsReadOnlySQL(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		want bool
	}{
		{"SELECT", "SELECT * FROM users", true},
		{"带注释的 SELECT", "-- 查询用户\nSELECT * FROM users", true},
		{"SHOW", "SHOW TABLES", true},
		{"EXPLAIN", "EXPLAIN SELECT * FROM users", true},
		{"WITH 可能携带写操作", "WITH deleted AS (DELETE FROM users RETURNING id) SELECT * FROM deleted", false},
		{"PRAGMA 可能修改状态", "PRAGMA foreign_keys = OFF", false},
		{"DELETE", "DELETE FROM users", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isReadOnlySQL(tc.sql); got != tc.want {
				t.Fatalf("isReadOnlySQL(%q) = %v, want %v", tc.sql, got, tc.want)
			}
		})
	}
}

func TestQueryReadOnlySQL(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	cols, rows, err := queryReadOnlySQL(context.Background(), "SELECT 1 AS value")
	if err != nil {
		t.Fatalf("query read-only SQL: %v", err)
	}
	if len(cols) != 1 || cols[0] != "value" {
		t.Fatalf("columns=%v, want [value]", cols)
	}
	if len(rows) != 1 || rows[0]["value"] != int64(1) {
		t.Fatalf("rows=%v, want one value row", rows)
	}
}

func TestExecSQLParsedWritesOnce(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	if err := db.DB.Exec("CREATE TABLE db_console_write_test (id INTEGER PRIMARY KEY, value TEXT NOT NULL)").Error; err != nil {
		t.Fatalf("create test table: %v", err)
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/db/sql", nil)

	ctrl := NewDBConsoleController()
	ctrl.execSQLParsed(ctx, execSQLRequest{
		SQL:        "INSERT INTO db_console_write_test (id, value) VALUES (1, 'once')",
		AllowWrite: true,
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var count int64
	if err := db.DB.Table("db_console_write_test").Count(&count).Error; err != nil {
		t.Fatalf("count inserted rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("write executed %d times, want 1", count)
	}
}

func newDBConsoleTestContext(method, target string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}
	ctx.Request = httptest.NewRequest(method, target, reader)
	ctx.Request.Header.Set("Content-Type", "application/json")
	return ctx, recorder
}

func TestBuildTableMetaAndDDL(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	statements := []string{
		`CREATE TABLE db_console_parent (id INTEGER PRIMARY KEY)`,
		`CREATE TABLE db_console_meta (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER NOT NULL REFERENCES db_console_parent(id),
			name TEXT NOT NULL DEFAULT '',
			code TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX idx_db_console_meta_name_code ON db_console_meta(name, code)`,
	}
	for _, statement := range statements {
		if err := db.DB.Exec(statement).Error; err != nil {
			t.Fatalf("create metadata fixture: %v", err)
		}
	}

	meta, err := buildTableMeta("db_console_meta")
	if err != nil {
		t.Fatalf("build metadata: %v", err)
	}
	if len(meta.Columns) != 4 {
		t.Fatalf("columns=%d, want 4", len(meta.Columns))
	}
	if len(meta.ForeignKeys) != 1 || meta.ForeignKeys[0].RefTable != "db_console_parent" {
		t.Fatalf("foreign keys=%+v, want parent foreign key", meta.ForeignKeys)
	}
	var composite *dbIndexMeta
	for i := range meta.Indexes {
		if meta.Indexes[i].Name == "idx_db_console_meta_name_code" {
			composite = &meta.Indexes[i]
			break
		}
	}
	if composite == nil || !composite.Unique || len(composite.Columns) != 2 {
		t.Fatalf("indexes=%+v, want unique composite index", meta.Indexes)
	}
	ddl, err := buildDDL("db_console_meta", meta)
	if err != nil {
		t.Fatalf("build ddl: %v", err)
	}
	if !strings.Contains(ddl, "CREATE TABLE") || !strings.Contains(ddl, "db_console_meta") {
		t.Fatalf("ddl=%q, want create table statement", ddl)
	}
}

func TestTableRowCRUDAndSensitiveFieldGuard(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	if err := db.DB.Exec(`CREATE TABLE db_console_row (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		password TEXT
	)`).Error; err != nil {
		t.Fatalf("create row fixture: %v", err)
	}
	ctrl := NewDBConsoleController()

	createCtx, createRecorder := newDBConsoleTestContext(http.MethodPost, "/db/tables/db_console_row/rows", map[string]interface{}{
		"values": map[string]interface{}{"id": 1, "name": "before"},
	})
	createCtx.Params = gin.Params{{Key: "name", Value: "db_console_row"}}
	ctrl.CreateTableRow(createCtx)
	if createRecorder.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", createRecorder.Code, createRecorder.Body.String())
	}

	updateCtx, updateRecorder := newDBConsoleTestContext(http.MethodPatch, "/db/tables/db_console_row/rows", map[string]interface{}{
		"primary_key": map[string]interface{}{"id": 1},
		"values":      map[string]interface{}{"name": "after"},
	})
	updateCtx.Params = gin.Params{{Key: "name", Value: "db_console_row"}}
	ctrl.UpdateTableRow(updateCtx)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updateRecorder.Code, updateRecorder.Body.String())
	}
	var name string
	if err := db.DB.Raw(`SELECT name FROM db_console_row WHERE id = 1`).Scan(&name).Error; err != nil || name != "after" {
		t.Fatalf("updated name=%q err=%v", name, err)
	}

	sensitiveCtx, sensitiveRecorder := newDBConsoleTestContext(http.MethodPatch, "/db/tables/db_console_row/rows", map[string]interface{}{
		"primary_key": map[string]interface{}{"id": 1},
		"values":      map[string]interface{}{"password": "should-not-write"},
	})
	sensitiveCtx.Params = gin.Params{{Key: "name", Value: "db_console_row"}}
	ctrl.UpdateTableRow(sensitiveCtx)
	if sensitiveRecorder.Code != http.StatusBadRequest {
		t.Fatalf("sensitive update status=%d body=%s", sensitiveRecorder.Code, sensitiveRecorder.Body.String())
	}

	deleteCtx, deleteRecorder := newDBConsoleTestContext(http.MethodDelete, "/db/tables/db_console_row/rows", map[string]interface{}{
		"primary_key": map[string]interface{}{"id": 1},
	})
	deleteCtx.Params = gin.Params{{Key: "name", Value: "db_console_row"}}
	ctrl.DeleteTableRow(deleteCtx)
	if deleteRecorder.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleteRecorder.Code, deleteRecorder.Body.String())
	}
	var count int64
	if err := db.DB.Table("db_console_row").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("row count=%d err=%v, want 0", count, err)
	}
}

func TestDBConsoleWritesDisabledInProduction(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	if err := db.DB.Exec("CREATE TABLE db_console_prod (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("create test table: %v", err)
	}
	previous := config.CloneGlobalConfig()
	config.UpdateGlobalConfig(func(cfg *config.Config) {
		cfg.Environment = "production"
		cfg.EnableAdminDBWrite = true
	})
	t.Cleanup(func() {
		config.SetGlobalConfig(previous)
	})

	ctx, recorder := newDBConsoleTestContext(http.MethodPost, "/db/sql", nil)
	NewDBConsoleController().execSQLParsed(ctx, execSQLRequest{
		SQL:        "INSERT INTO db_console_prod (id, name) VALUES (1, 'blocked')",
		AllowWrite: true,
	})
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var count int64
	if err := db.DB.Table("db_console_prod").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("row count=%d err=%v, want 0", count, err)
	}
}

func TestDBConsoleAuditKeepsPlaintext(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	ctx, _ := newDBConsoleTestContext(http.MethodPost, "/db/sql", nil)
	writeDBConsoleAudit(ctx, "test_audit_plain", `{"password":"super-secret","name":"safe"}`, "ok", http.StatusOK)

	var logItem models.OperationLog
	if err := db.DB.Where("action = ?", "test_audit_plain").Order("id DESC").First(&logItem).Error; err != nil {
		t.Fatalf("load operation log: %v", err)
	}
	// 操作日志按产品要求明文落库，不做字段脱敏
	if logItem.RequestBody == nil || !strings.Contains(*logItem.RequestBody, "super-secret") {
		t.Fatalf("audit request body=%v, want plaintext password kept", logItem.RequestBody)
	}
}
