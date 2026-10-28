package routes_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/routes"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

func setupAdminSelfProtection(t *testing.T, name string) (*models.User, string) {
	t.Helper()

	admin := testutil.CreateTestAdmin(t, name)
	token, err := utils.GenerateTokenForGuardWithTTL(admin.ID, "admin", utils.AdminAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("生成管理员 token 失败: %v", err)
	}
	if err := models.CreateUserSession(
		admin.ID,
		utils.AdminAuthGuard,
		utils.HashToken(token),
		utils.HashToken("refresh-token"),
		"127.0.0.1",
		"test",
		"pc",
		"web",
		"",
		time.Now().Add(time.Hour).Unix(),
		time.Now().Add(2*time.Hour).Unix(),
	); err != nil {
		t.Fatalf("创建管理员会话失败: %v", err)
	}
	return admin, token
}

func performAdminJSON(t *testing.T, handler http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestAdminSelfProtection_BatchStatusSkipsSelf(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	admin, token := setupAdminSelfProtection(t, "self_guard_batch_status_admin")
	user := testutil.CreateTestUser(t, "self_guard_batch_status_user")

	r := gin.New()
	routes.SetupRoutes(r)

	body := fmt.Sprintf(`{"ids":[%d,%d],"status":0}`, admin.ID, user.ID)
	w := performAdminJSON(t, r, http.MethodPost, "/api/v1/admin/users/batch-status", token, body)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Succeeded int `json:"succeeded"`
			Failed    int `json:"failed"`
			Errors    []struct {
				ID      uint64 `json:"id"`
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("批量状态响应不是 JSON: %v", err)
	}
	if w.Code != http.StatusOK || resp.Code != http.StatusOK || resp.Data.Succeeded != 1 || resp.Data.Failed != 1 {
		t.Fatalf("批量禁用应逐条跳过自己，HTTP %d: %s", w.Code, w.Body.String())
	}
	if len(resp.Data.Errors) != 1 || resp.Data.Errors[0].ID != admin.ID {
		t.Fatalf("批量禁用错误项应包含当前管理员: %s", w.Body.String())
	}

	adminAfter, err := models.GetUserByID(admin.ID)
	if err != nil || adminAfter.Status != 1 {
		t.Fatalf("当前管理员不应被禁用: status=%v err=%v", adminAfter, err)
	}
	userAfter, err := models.GetUserByID(user.ID)
	if err != nil || userAfter.Status != 0 {
		t.Fatalf("普通用户应被禁用: status=%v err=%v", userAfter, err)
	}
}

func TestAdminSelfProtection_BatchDeleteSkipsSelf(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	admin, token := setupAdminSelfProtection(t, "self_guard_batch_delete_admin")
	user := testutil.CreateTestUser(t, "self_guard_batch_delete_user")

	r := gin.New()
	routes.SetupRoutes(r)

	body := fmt.Sprintf(`{"ids":[%d,%d]}`, admin.ID, user.ID)
	w := performAdminJSON(t, r, http.MethodPost, "/api/v1/admin/users/batch-delete", token, body)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Succeeded int `json:"succeeded"`
			Failed    int `json:"failed"`
			Errors    []struct {
				ID      uint64 `json:"id"`
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("批量删除响应不是 JSON: %v", err)
	}
	if w.Code != http.StatusOK || resp.Code != http.StatusOK || resp.Data.Succeeded != 1 || resp.Data.Failed != 1 {
		t.Fatalf("批量删除应逐条跳过自己，HTTP %d: %s", w.Code, w.Body.String())
	}
	if len(resp.Data.Errors) != 1 || resp.Data.Errors[0].ID != admin.ID {
		t.Fatalf("批量删除错误项应包含当前管理员: %s", w.Body.String())
	}

	if _, err := models.GetUserByID(admin.ID); err != nil {
		t.Fatalf("当前管理员不应被删除: %v", err)
	}
	if _, err := models.GetUserByID(user.ID); err == nil {
		t.Fatal("普通用户应已被删除")
	}
}

func TestAdminSelfProtection_SingleStatusCannotDisableSelf(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	admin, token := setupAdminSelfProtection(t, "self_guard_single_status_admin")
	user := testutil.CreateTestUser(t, "self_guard_single_status_user")

	r := gin.New()
	routes.SetupRoutes(r)

	w := performAdminJSON(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d/status", admin.ID), token, `{"status":0}`)
	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("单条禁用响应不是 JSON: %v", err)
	}
	if w.Code != http.StatusBadRequest || resp.Code != http.StatusBadRequest || resp.Message != "不能禁用当前登录的管理员账号" {
		t.Fatalf("自禁用应被拒绝，HTTP %d: %s", w.Code, w.Body.String())
	}
	adminAfter, err := models.GetUserByID(admin.ID)
	if err != nil || adminAfter.Status != 1 {
		t.Fatalf("当前管理员状态不应变化: status=%v err=%v", adminAfter, err)
	}

	// status=1 对自身的幂等启用不应被自保护误伤
	w = performAdminJSON(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d/status", admin.ID), token, `{"status":1}`)
	var enableResp struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &enableResp); err != nil {
		t.Fatalf("自身启用响应不是 JSON: %v", err)
	}
	if w.Code != http.StatusOK || enableResp.Code != http.StatusOK {
		t.Fatalf("自身 status=1 不应被拦截，HTTP %d: %s", w.Code, w.Body.String())
	}
	adminAfterEnable, err := models.GetUserByID(admin.ID)
	if err != nil || adminAfterEnable.Status != 1 {
		t.Fatalf("自身 status=1 后管理员状态应保持启用: status=%v err=%v", adminAfterEnable, err)
	}

	w = performAdminJSON(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d/status", user.ID), token, `{"status":0}`)
	var normalResp struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &normalResp); err != nil {
		t.Fatalf("普通用户禁用响应不是 JSON: %v", err)
	}
	if w.Code != http.StatusOK || normalResp.Code != http.StatusOK {
		t.Fatalf("普通用户禁用不应被误伤，HTTP %d: %s", w.Code, w.Body.String())
	}
	userAfter, err := models.GetUserByID(user.ID)
	if err != nil || userAfter.Status != 0 {
		t.Fatalf("普通用户应已被禁用: status=%v err=%v", userAfter, err)
	}
}

func TestAdminSelfProtection_SingleDeleteCannotDeleteSelf(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	admin, token := setupAdminSelfProtection(t, "self_guard_single_delete_admin")

	r := gin.New()
	routes.SetupRoutes(r)

	w := performAdminJSON(t, r, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", admin.ID), token, "")
	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("单条删除响应不是 JSON: %v", err)
	}
	if w.Code != http.StatusBadRequest || resp.Code != http.StatusBadRequest || resp.Message != "不能删除当前登录的管理员账号" {
		t.Fatalf("自删除应被拒绝，HTTP %d: %s", w.Code, w.Body.String())
	}
	if _, err := models.GetUserByID(admin.ID); err != nil {
		t.Fatalf("当前管理员不应被删除: %v", err)
	}
}
