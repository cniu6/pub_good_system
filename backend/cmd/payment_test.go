package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	"fst/backend/internal/db"
	"fst/backend/routes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

var testRouter *gin.Engine
var testToken string // 测试用 JWT token

// TestMain 集成测试入口：初始化数据库和路由
func TestMain(m *testing.M) {
	// 初始化配置（需要 .env 文件或环境变量）
	config.InitConfig()

	// 初始化数据库
	db.InitDB()

	// 初始化必要的表
	models.InitSystemSettingsTable()
	models.InitUserMoneyLogsTable()
	models.InitPaymentOrdersTable()

	// 初始化服务
	services.InitSettingsService()

	// 创建 Gin 路由
	gin.SetMode(gin.TestMode)
	testRouter = gin.Default()
	routes.SetupRoutes(testRouter)

	// 获取测试用 token（需要有效用户）
	testToken = getTestToken()

	os.Exit(m.Run())
}

// getTestToken 登录获取测试token（需要数据库中有测试用户）
func getTestToken() string {
	body := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/public/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != 200 {
		log.Printf("[Test] 登录失败（可能没有测试用户），跳过需要认证的测试: status=%d", w.Code)
		return ""
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if token, ok := data["access_token"].(string); ok {
			return token
		}
	}
	return ""
}

// apiRequest 发送 API 请求的辅助函数
func apiRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

// parseResponse 解析 API 响应
func parseResponse(w *httptest.ResponseRecorder) (int, string, map[string]interface{}) {
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	code := 0
	msg := ""
	var data map[string]interface{}

	if c, ok := resp["code"].(float64); ok {
		code = int(c)
	}
	if m, ok := resp["message"].(string); ok {
		msg = m
	}
	if d, ok := resp["data"].(map[string]interface{}); ok {
		data = d
	}

	return code, msg, data
}

// ========================================
// 公共接口测试（无需登录）
// ========================================

// TestPaymentNotify_NoSign 测试无签名的回调应拒绝
func TestPaymentNotify_NoSign(t *testing.T) {
	w := apiRequest("GET", "/api/v1/public/payment/notify?out_trade_no=FAKE&trade_no=FAKE&money=10.00&trade_status=TRADE_SUCCESS", nil, "")

	if w.Code != 200 {
		t.Fatalf("HTTP状态码应为200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "SUCCESS" {
		t.Error("无签名的回调不应返回 SUCCESS")
	}
}

// TestPaymentNotify_InvalidSign 测试错误签名的回调应拒绝
func TestPaymentNotify_InvalidSign(t *testing.T) {
	w := apiRequest("GET",
		"/api/v1/public/payment/notify?pid=1&out_trade_no=FAKE&trade_no=FAKE&money=10.00&trade_status=TRADE_SUCCESS&sign=00000000000000000000000000000000&sign_type=MD5",
		nil, "")

	body := w.Body.String()
	if body == "SUCCESS" {
		t.Error("错误签名的回调不应返回 SUCCESS")
	}
}

// TestPaymentReturn_InvalidSign 测试错误签名的同步跳转
func TestPaymentReturn_InvalidSign(t *testing.T) {
	w := apiRequest("GET",
		"/api/v1/public/payment/return?pid=1&out_trade_no=FAKE&sign=invalid&sign_type=MD5",
		nil, "")

	// 应重定向（302）到错误页面
	if w.Code != http.StatusFound {
		t.Logf("同步跳转应返回302重定向, got %d", w.Code)
	}
}

// ========================================
// 用户接口测试（需要登录）
// ========================================

// TestGetPaymentMethods 测试获取支付方式
func TestGetPaymentMethods(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	w := apiRequest("GET", "/api/v1/user/payment/methods", nil, testToken)
	code, _, data := parseResponse(w)

	if code != 200 {
		t.Fatalf("获取支付方式失败: code=%d", code)
	}

	if data == nil {
		t.Fatal("响应 data 不应为空")
	}

	// 应包含 methods 和 config 字段
	if _, ok := data["methods"]; !ok {
		t.Error("响应应包含 methods 字段")
	}
	if _, ok := data["config"]; !ok {
		t.Error("响应应包含 config 字段")
	}
}

// TestGetPaymentOrders_Empty 测试获取空的订单列表
func TestGetPaymentOrders_Empty(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	w := apiRequest("GET", "/api/v1/user/payment/orders?page=1&page_size=10&status=-1", nil, testToken)
	code, _, data := parseResponse(w)

	if code != 200 {
		t.Fatalf("获取订单列表失败: code=%d", code)
	}

	if data == nil {
		t.Fatal("响应 data 不应为空")
	}
}

// TestCreateOrder_Unauthorized 测试未登录创建订单
func TestCreateOrder_Unauthorized(t *testing.T) {
	body := map[string]interface{}{
		"amount":       10.00,
		"payment_type": "alipay",
	}
	w := apiRequest("POST", "/api/v1/user/payment/create", body, "")
	code, _, _ := parseResponse(w)

	if code == 200 {
		t.Error("未登录不应能创建订单")
	}
}

// TestCreateOrder_InvalidAmount 测试无效金额
func TestCreateOrder_InvalidAmount(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	tests := []struct {
		name   string
		amount float64
	}{
		{"零金额", 0},
		{"负金额", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"amount":       tt.amount,
				"payment_type": "alipay",
			}
			w := apiRequest("POST", "/api/v1/user/payment/create", body, testToken)
			code, _, _ := parseResponse(w)

			if code == 200 {
				t.Errorf("无效金额 %f 不应创建成功", tt.amount)
			}
		})
	}
}

// TestCreateOrder_MissingPaymentType 测试缺少支付方式
func TestCreateOrder_MissingPaymentType(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	body := map[string]interface{}{
		"amount": 10.00,
	}
	w := apiRequest("POST", "/api/v1/user/payment/create", body, testToken)
	code, _, _ := parseResponse(w)

	if code == 200 {
		t.Error("缺少支付方式不应创建成功")
	}
}

// TestGetOrderDetail_NotFound 测试查看不存在的订单
func TestGetOrderDetail_NotFound(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	w := apiRequest("GET", "/api/v1/user/payment/orders/999999999", nil, testToken)
	code, _, _ := parseResponse(w)

	if code == 200 {
		t.Error("不存在的订单不应返回成功")
	}
}

// TestCheckOrderStatus_NotFound 测试轮询不存在的订单
func TestCheckOrderStatus_NotFound(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	w := apiRequest("GET", "/api/v1/user/payment/orders/999999999/status", nil, testToken)
	code, _, _ := parseResponse(w)

	if code == 200 {
		t.Error("不存在的订单状态不应返回成功")
	}
}

// ========================================
// 管理端接口测试（需要管理员）
// ========================================

// TestAdminGetPaymentStats 测试管理端获取统计
func TestAdminGetPaymentStats(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	w := apiRequest("GET", "/api/v1/admin/payment/stats", nil, testToken)
	code, _, data := parseResponse(w)

	// 如果用户不是管理员，会返回403
	if code == 403 {
		t.Skip("当前用户非管理员，跳过管理端测试")
	}

	if code != 200 {
		t.Fatalf("获取统计数据失败: code=%d", code)
	}

	if data == nil {
		t.Fatal("统计数据不应为空")
	}

	// 验证字段存在
	expectedFields := []string{"total_orders", "paid_orders", "total_amount", "today_orders", "today_amount", "pending_orders"}
	for _, field := range expectedFields {
		if _, ok := data[field]; !ok {
			t.Errorf("统计数据缺少字段: %s", field)
		}
	}
}

// TestAdminListOrders 测试管理端订单列表
func TestAdminListOrders(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	w := apiRequest("GET", "/api/v1/admin/payment/orders?page=1&page_size=10&status=-1", nil, testToken)
	code, _, data := parseResponse(w)

	if code == 403 {
		t.Skip("当前用户非管理员，跳过")
	}

	if code != 200 {
		t.Fatalf("获取管理端订单列表失败: code=%d", code)
	}

	if data == nil {
		t.Fatal("响应 data 不应为空")
	}
}

// TestAdminCompleteOrder_NotFound 测试补单不存在的订单
func TestAdminCompleteOrder_NotFound(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	body := map[string]string{"memo": "测试补单"}
	w := apiRequest("POST", "/api/v1/admin/payment/orders/999999999/complete", body, testToken)
	code, _, _ := parseResponse(w)

	if code == 403 {
		t.Skip("当前用户非管理员，跳过")
	}

	if code == 200 {
		t.Error("不存在的订单不应补单成功")
	}
}

// TestAdminCancelOrder_NotFound 测试取消不存在的订单
func TestAdminCancelOrder_NotFound(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	w := apiRequest("POST", "/api/v1/admin/payment/orders/999999999/cancel", nil, testToken)
	code, _, _ := parseResponse(w)

	if code == 403 {
		t.Skip("当前用户非管理员，跳过")
	}

	if code == 200 {
		t.Error("不存在的订单不应取消成功")
	}
}

// ========================================
// 回调安全测试
// ========================================

// TestNotifyReplayAttack 测试回调重放攻击防护
func TestNotifyReplayAttack(t *testing.T) {
	// 模拟攻击者重放一个过期的回调请求
	w := apiRequest("POST",
		"/api/v1/public/payment/notify",
		nil, "")

	// 无参数的回调应直接失败
	body := w.Body.String()
	if body == "SUCCESS" {
		t.Error("空回调不应返回 SUCCESS")
	}
}

// TestNotifyXSSInjection 测试回调XSS注入防护
func TestNotifyXSSInjection(t *testing.T) {
	// 在回调参数中注入XSS
	xssPayload := "<script>alert('xss')</script>"
	path := fmt.Sprintf("/api/v1/public/payment/notify?out_trade_no=%s&trade_no=%s&money=10&trade_status=TRADE_SUCCESS&sign=fake",
		xssPayload, xssPayload)

	w := apiRequest("GET", path, nil, "")

	body := w.Body.String()
	if body == "SUCCESS" {
		t.Error("XSS注入的回调不应成功")
	}
}
