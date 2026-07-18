//go:build integration

package admin_stats

import (
	"fst/backend/app/models"
	"testing"
)

func TestGetWithdrawRequestStatsRegression(t *testing.T) {
	models.InitWithdrawRequestsTable()

	stats, err := models.GetWithdrawRequestStats(&models.WithdrawListQuery{})
	if err != nil {
		t.Fatalf("获取提现统计失败: %v", err)
	}
	if stats == nil {
		t.Fatal("提现统计不应为空")
	}
}

func TestAdminGetWithdrawStats(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	models.InitWithdrawRequestsTable()

	w := apiRequest("GET", "/api/v1/admin/withdraw/stats", nil, testToken)
	code, msg, data := parseResponse(w)

	if code == 403 {
		t.Skip("当前用户非管理员，跳过管理端测试")
	}

	if code != 200 {
		t.Fatalf("获取提现统计失败: code=%d, msg=%s, body=%s", code, msg, w.Body.String())
	}
	if data == nil {
		t.Fatal("提现统计响应 data 不应为空")
	}

	expectedFields := []string{"pending_count", "approved_count", "rejected_count", "paid_count", "paid_amount"}
	for _, field := range expectedFields {
		if _, ok := data[field]; !ok {
			t.Errorf("提现统计缺少字段: %s", field)
		}
	}
}

func TestAdminGetDashboard(t *testing.T) {
	if testToken == "" {
		t.Skip("无测试 token，跳过")
	}

	w := apiRequest("GET", "/api/v1/admin/dashboard", nil, testToken)
	code, msg, data := parseResponse(w)

	if code == 403 {
		t.Skip("当前用户非管理员，跳过管理端测试")
	}

	if code != 200 {
		t.Fatalf("获取后台仪表盘失败: code=%d, msg=%s, body=%s", code, msg, w.Body.String())
	}
	if data == nil {
		t.Fatal("后台仪表盘响应 data 不应为空")
	}

	statisticsValue, ok := data["statistics"]
	if !ok {
		t.Fatalf("后台仪表盘响应缺少 statistics 字段: body=%s", w.Body.String())
	}
	statistics, ok := statisticsValue.(map[string]interface{})
	if !ok {
		t.Fatalf("statistics 字段类型错误: %T", statisticsValue)
	}

	topLevelFields := []string{"recent_users", "recent_login_users", "trends"}
	for _, field := range topLevelFields {
		if _, ok := data[field]; !ok {
			t.Errorf("后台仪表盘缺少字段: %s", field)
		}
	}

	expectedStatFields := []string{
		"total_payment_orders",
		"paid_payment_orders",
		"pending_payment_orders",
		"pending_withdraw_count",
		"approved_withdraw_count",
		"paid_withdraw_count",
		"paid_withdraw_amount",
		"total_realname_requests",
		"pending_realname_count",
		"approved_realname_count",
		"rejected_realname_count",
	}
	for _, field := range expectedStatFields {
		if _, ok := statistics[field]; !ok {
			t.Errorf("后台仪表盘 statistics 缺少字段: %s", field)
		}
	}
}
