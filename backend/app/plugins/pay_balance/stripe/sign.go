package stripe

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// verifyWebhookSignature 使用 endpoint_secret 验证 Stripe-Signature 头
// 签名格式：t=<timestamp>,v1=<signature>
// 计算：HMAC-SHA256(endpoint_secret, timestamp + "." + raw_body)
func verifyWebhookSignature(rawBody []byte, signatureHeader, secret string, tolerance time.Duration) (bool, error) {
	secret = strings.TrimSpace(secret)
	signatureHeader = strings.TrimSpace(signatureHeader)
	if secret == "" || signatureHeader == "" {
		return false, errors.New("stripe secret or signature header missing")
	}

	// 解析 header
	pairs := strings.Split(signatureHeader, ",")
	var timestamp int64
	var signatures []string
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp, _ = strconv.ParseInt(kv[1], 10, 64)
		case "v1":
			signatures = append(signatures, kv[1])
		}
	}
	if timestamp == 0 || len(signatures) == 0 {
		return false, errors.New("invalid stripe-signature header")
	}

	// 时间窗口校验
	now := time.Now().Unix()
	if now-timestamp > int64(tolerance.Seconds()) || timestamp-now > 60 {
		return false, fmt.Errorf("stripe webhook timestamp out of tolerance: %d", timestamp)
	}

	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(rawBody))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expected := hex.EncodeToString(mac.Sum(nil))

	for _, sig := range signatures {
		if hmac.Equal([]byte(sig), []byte(expected)) {
			return true, nil
		}
	}
	return false, nil
}
