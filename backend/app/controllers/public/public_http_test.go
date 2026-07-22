package public_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fst/backend/app/controllers/public"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestPublicAuthAndSettings(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	g := r.Group("/api/v1/public")
	public.NewAuthController().RegisterRoutes(g)
	public.NewSettingsController().RegisterRoutes(g)
	public.NewGeoController().RegisterRoutes(g)
	public.NewPaymentCallbackController().RegisterRoutes(g)

	t.Run("app-config可读", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/app-config", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp["code"] != float64(0) && resp["code"] != float64(200) {
			t.Fatalf("code=%v body=%s", resp["code"], w.Body.String())
		}
	})

	t.Run("phone-country探测可读", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/geo/phone-country?lang=zhCN", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp["code"] != float64(0) && resp["code"] != float64(200) {
			t.Fatalf("code=%v body=%s", resp["code"], w.Body.String())
		}
		data, _ := resp["data"].(map[string]any)
		if data == nil {
			t.Fatalf("data empty: %s", w.Body.String())
		}
		if data["country_code"] == nil || data["dial_code"] == nil {
			t.Fatalf("missing country fields: %v", data)
		}
	})

	t.Run("登录缺密码失败", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"username": "nobody"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["code"] == float64(0) {
			t.Fatalf("缺密码不应成功: %s", w.Body.String())
		}
	})

	t.Run("注册缺验证码失败", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": "u1", "password": "password1", "email": "a@b.com",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["code"] == float64(0) {
			t.Fatalf("缺验证码不应成功: %s", w.Body.String())
		}
	})

	// TestBugFix: 修复前，Register 会在极验校验之前就先查用户名/邮箱是否存在，
	// 意味着攻击者不用过人机验证就能批量探测账号是否已注册。修复后极验必须先通过。
	t.Run("注册撞用户名时应先过极验再暴露是否已存在", func(t *testing.T) {
		services.InitSettingsService()
		if err := models.UpdateSetting("geetest_enabled", "true"); err != nil {
			t.Fatalf("UpdateSetting geetest_enabled: %v", err)
		}
		if err := models.UpdateSetting("geetest_captcha_id", "test-captcha-id"); err != nil {
			t.Fatalf("UpdateSetting geetest_captcha_id: %v", err)
		}
		if err := models.UpdateSetting("geetest_captcha_key", "test-captcha-key"); err != nil {
			t.Fatalf("UpdateSetting geetest_captcha_key: %v", err)
		}
		if err := services.GlobalSettingsService.RefreshCache(); err != nil {
			t.Fatalf("RefreshCache: %v", err)
		}
		defer func() {
			_ = models.UpdateSetting("geetest_enabled", "false")
			_ = services.GlobalSettingsService.RefreshCache()
		}()

		// 造一个已存在的用户名，用来触发「存在性检查」分支
		existingUser := &models.User{
			Username: "captcha_order_user", Password: "x",
			Email: "captcha-order@example.test", Role: "user", Status: 1,
		}
		if err := models.CreateUser(existingUser); err != nil {
			t.Fatalf("CreateUser: %v", err)
		}

		body, _ := json.Marshal(map[string]string{
			"username": "captcha_order_user", // 已存在，会触发存在性检查
			"password": "password1",
			"email":    "brand-new-email@example.test",
			"code":     "000000",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// 故意不带任何 X-Geetest-* 头，模拟未过极验
		r.ServeHTTP(w, req)

		if w.Code != 403 {
			t.Fatalf("未过极验时应返回 403，实际 status=%d body=%s", w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		msg, _ := resp["message"].(string)
		if msg != "Captcha validation failed" {
			t.Fatalf("应先被极验拦下而不是暴露用户名是否已存在；实际 message=%q body=%s", msg, w.Body.String())
		}
	})

	t.Run("未知用户登录失败", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": "no-such-user-xyz", "password": "password1",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["code"] == float64(0) {
			t.Fatalf("未知用户不应登录成功: %s", w.Body.String())
		}
	})
}
