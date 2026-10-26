package middleware

import (
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsPaymentCallbackPath(t *testing.T) {
	for _, path := range []string{
		"/api/v1/public/payment/notify",
		"/api/v1/public/payment/notify/epay",
		"/api/v1/public/payment/return",
	} {
		if !isPaymentCallbackPath(path) {
			t.Errorf("%s 应判定为支付回调", path)
		}
	}
	for _, path := range []string{
		"/api/v1/public/payment",
		"/api/v1/public/settings",
		"/api/v1/user/payment/orders",
		"/api/v1/public/payment/notify-other",
		"/api/v1/public/payment/return-other",
	} {
		if isPaymentCallbackPath(path) {
			t.Errorf("%s 不应判定为支付回调", path)
		}
	}
}

// 回调 query 里的签名必须打码，订单号/金额/状态这类对账字段要保持原样。
func TestRedactPaymentQueryMasksSignature(t *testing.T) {
	raw := "out_trade_no=ORDER123&trade_no=T1&money=9.90&sign=deadbeef&sign_type=MD5&trade_status=TRADE_SUCCESS"

	got := redactPaymentQuery(raw)
	if strings.Contains(got, "deadbeef") {
		t.Fatalf("签名未打码: %s", got)
	}

	values, err := url.ParseQuery(got)
	if err != nil {
		t.Fatalf("打码后 query 不可解析: %v", err)
	}
	if values.Get("sign") != "***" {
		t.Errorf("sign = %q, want ***", values.Get("sign"))
	}
	if values.Get("out_trade_no") != "ORDER123" || values.Get("money") != "9.90" || values.Get("sign_type") != "MD5" {
		t.Errorf("对账字段被误改: %s", got)
	}
}

func TestRedactPaymentQueryEdgeCases(t *testing.T) {
	for _, raw := range []string{"", "   "} {
		if got := redactPaymentQuery(raw); got != "" {
			t.Errorf("空 query 应返回空字符串，得到 %q", got)
		}
	}
	if got := redactPaymentQuery("sign=%zz&money=9.90"); got != "[query omitted: payment callback]" {
		t.Errorf("无效编码不能回退到明文: %q", got)
	}

	raw := "sign=first&sign=second&SIGN=upper&%73ign=encoded&token=token-value&custom_signature=new-provider-value"
	values, err := url.ParseQuery(redactPaymentQuery(raw))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"sign", "SIGN", "token", "custom_signature"} {
		if items := values[key]; len(items) != 1 || items[0] != "***" {
			t.Errorf("%s 的所有值都应被打码: %v", key, items)
		}
	}
}

func TestPaymentCallbackLogMetadataPreservesRequest(t *testing.T) {
	for _, path := range []string{
		"/api/v1/public/payment/notify",
		"/api/v1/public/payment/notify/stripe",
		"/api/v1/public/payment/return",
	} {
		t.Run(path, func(t *testing.T) {
			const body = "out_trade_no=ORDER123&sign=body-signature"
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", path+"?sign=query-signature&money=9.90", strings.NewReader(body))
			c.Request.Header.Set("Stripe-Signature", "header-signature")
			c.Request.Header.Set("Referer", "https://example.test/callback?sign=referer-signature")

			headers, referer := accessLogRequestMetadata(c)
			if headers != nil || referer != "" {
				t.Fatal("支付回调头部与 Referer 不应进入访问日志")
			}
			query, err := url.ParseQuery(accessLogQueryString(c))
			if err != nil || query.Get("sign") != "***" || query.Get("money") != "9.90" {
				t.Fatalf("回调 query 脱敏错误: %v %v", query, err)
			}
			gotBody, err := io.ReadAll(c.Request.Body)
			if err != nil || string(gotBody) != body || c.Query("sign") != "query-signature" ||
				c.GetHeader("Stripe-Signature") != "header-signature" || c.Request.Referer() == "" {
				t.Fatal("日志脱敏不能修改验签使用的请求")
			}
		})
	}
}

func TestNonPaymentLogMetadataUnchanged(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/api/v1/public/settings?sign=ordinary-value", nil)
	c.Request.Header.Set("X-Test", "ordinary-header")
	c.Request.Header.Set("Referer", "https://example.test/settings")
	wantHeaders := captureRequestHeaders(c.Request.Header)
	headers, referer := accessLogRequestMetadata(c)
	if headers == nil || wantHeaders == nil || *headers != *wantHeaders || referer != c.Request.Referer() {
		t.Fatal("普通接口的请求头与 Referer 日志行为不能改变")
	}
	if got := accessLogQueryString(c); got != c.Request.URL.RawQuery {
		t.Errorf("普通 query 被修改: %q", got)
	}
}

func TestAPIAccessLogSkipRulesPreserved(t *testing.T) {
	for _, path := range []string{"/health", adminAPIBasePath() + "/debug/pprof", adminAPIBasePath() + "/debug/pprof/profile"} {
		if !shouldSkipAPIAccessLog(path) {
			t.Errorf("该路径应跳过访问日志: %s", path)
		}
	}
	for _, path := range []string{"/api/v1/public/payment/notify", "/api/v1/public/payment/return", adminAPIBasePath() + "/dashboard"} {
		if shouldSkipAPIAccessLog(path) {
			t.Errorf("该路径仍应保留访问记录: %s", path)
		}
	}
}
