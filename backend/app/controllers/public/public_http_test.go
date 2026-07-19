package public_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fst/backend/app/controllers/public"
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
