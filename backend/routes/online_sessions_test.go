package routes_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/testutil"
	"fst/backend/routes"

	"github.com/gin-gonic/gin"
)

// TestAdminOnlineSessionsReturnsUsername 管理端在线列表接口必须返回 username/nickname。
// 回归：UserSession.Username/Nickname 曾标 gorm:"-"，JOIN users 的列被 Scan 丢弃，前端只能显示 #ID。
func TestAdminOnlineSessionsReturnsUsername(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	services.InitSettingsService()
	if err := models.UpdateSetting("presence_enabled", "true"); err != nil {
		t.Fatalf("开启 presence_enabled: %v", err)
	}
	if err := services.GlobalSettingsService.RefreshCache(); err != nil {
		t.Fatalf("刷新设置缓存: %v", err)
	}

	_, token := setupAdminSelfProtection(t, "online_list_admin")
	user := testutil.CreateTestUser(t, "route-online")
	now := time.Now().Unix()
	if err := models.CreateUserSession(
		user.ID, "user", "route-online-token", "route-online-refresh",
		"127.0.0.1", "test-agent", "pc", "web", "",
		now+3600, now+7200,
	); err != nil {
		t.Fatalf("创建用户会话: %v", err)
	}

	r := gin.New()
	routes.SetupRoutes(r)
	w := performAdminJSON(t, r, http.MethodGet, "/api/v1/admin/online/sessions", token, "")

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				UserID   uint64 `json:"user_id"`
				Username string `json:"username"`
				Nickname string `json:"nickname"`
				Devices  []struct {
					Username string `json:"username"`
				} `json:"devices"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应不是 JSON: %v", err)
	}
	if w.Code != http.StatusOK || resp.Code != http.StatusOK {
		t.Fatalf("在线列表请求失败 HTTP %d: %s", w.Code, w.Body.String())
	}

	var found bool
	for _, row := range resp.Data.List {
		if row.UserID != user.ID {
			continue
		}
		found = true
		if row.Username != user.Username || row.Nickname != user.Nickname {
			t.Fatalf("用户名未返回: username=%q nickname=%q", row.Username, row.Nickname)
		}
		if len(row.Devices) != 1 || row.Devices[0].Username != user.Username {
			t.Fatalf("设备行用户名未返回: %+v", row.Devices)
		}
	}
	if !found {
		t.Fatalf("在线列表未包含目标用户 %d: %s", user.ID, w.Body.String())
	}
}
