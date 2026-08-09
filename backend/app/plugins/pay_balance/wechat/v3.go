package wechat

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fst/backend/pkg/payment"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// v3Refund 微信 V3 退款
func v3Refund(ctx context.Context, extConfig map[string]string, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	mchID := strings.TrimSpace(extConfig["mch_id"])
	serialNo := strings.TrimSpace(extConfig["serial_no"])
	privateKey := strings.TrimSpace(extConfig["private_key"])
	if mchID == "" || serialNo == "" || privateKey == "" {
		return nil, fmt.Errorf("wechat v3 mch_id / serial_no / private_key missing")
	}
	if req.Money == "" {
		return nil, fmt.Errorf("wechat v3 refund money missing")
	}
	if req.OrderNo == "" {
		return nil, fmt.Errorf("wechat v3 refund requires out_trade_no")
	}

	refundMinor, err := payment.ParseMoneyMinor(req.Money)
	if err != nil {
		return nil, fmt.Errorf("invalid refund money: %w", err)
	}
	totalMinor := refundMinor
	if req.ExtConfig["total_money"] != "" {
		totalMinor, err = payment.ParseMoneyMinor(req.ExtConfig["total_money"])
		if err != nil {
			totalMinor = refundMinor
		}
	}

	outRefundNo := req.ExtConfig["out_refund_no"]
	if outRefundNo == "" {
		outRefundNo = fmt.Sprintf("R%s%d", req.OrderNo, time.Now().Unix())
	}

	body := map[string]interface{}{
		"out_refund_no": outRefundNo,
		"out_trade_no":  req.OrderNo,
		"amount": map[string]interface{}{
			"refund":   refundMinor,
			"total":    totalMinor,
			"currency": "CNY",
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	apiURL := "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds"
	headers, err := buildV3AuthHeader(http.MethodPost, "/v3/refund/domestic/refunds", bodyBytes, mchID, serialNo, privateKey)
	if err != nil {
		return nil, fmt.Errorf("wechat v3 refund auth failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("wechat v3 refund failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var refundResp struct {
		RefundID    string `json:"refund_id"`
		OutRefundNo string `json:"out_refund_no"`
		OutTradeNo  string `json:"out_trade_no"`
		Amount      struct {
			Refund int64 `json:"refund"`
		} `json:"amount"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &refundResp); err != nil {
		log.Printf("[WeChat] v3 refund parse failed: %s", string(respBody))
		return nil, fmt.Errorf("wechat v3 refund parse failed: %w", err)
	}

	code := 1
	if refundResp.Status != "SUCCESS" && refundResp.Status != "PROCESSING" {
		code = 0
	}

	return &payment.RefundResponse{
		Code:        code,
		Msg:         refundResp.Status,
		RefundNo:    refundResp.RefundID,
		OutRefundNo: refundResp.OutRefundNo,
		TradeNo:     refundResp.OutTradeNo,
		Money:       payment.FormatMoneyYuan(refundResp.Amount.Refund),
	}, nil
}

// buildV3AuthHeader 构造微信支付 V3 Authorization 头
func buildV3AuthHeader(method, requestURL string, body []byte, mchID, serialNo, privateKey string) (map[string]string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce, err := randomNonce(32)
	if err != nil {
		return nil, err
	}

	signPayload := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n",
		strings.ToUpper(method),
		requestURL,
		timestamp,
		nonce,
		string(body),
	)

	signature, err := payment.RSASign(signPayload, privateKey)
	if err != nil {
		return nil, err
	}

	auth := fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%s\",serial_no=\"%s\"",
		mchID, nonce, signature, timestamp, serialNo)

	return map[string]string{
		"Authorization":    auth,
		"Wechatpay-Serial": "",
		"Accept":           "application/json",
	}, nil
}

func randomNonce(n int) (string, error) {
	b := make([]byte, n/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
