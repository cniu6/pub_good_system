package paypal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// getAccessToken 用 client_id / client_secret 换 OAuth2 access token
func getAccessToken(clientID, clientSecret, apiURL string, timeout time.Duration) (string, error) {
	apiURL = strings.TrimRight(apiURL, "/")
	if apiURL == "" {
		apiURL = "https://api-m.sandbox.paypal.com"
	}

	req, err := http.NewRequest("POST", apiURL+"/v1/oauth2/token", strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("paypal oauth status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("paypal oauth empty token")
	}
	return result.AccessToken, nil
}

// VerifyWebhookSignature 调用 PayPal 官方接口验证 webhook 签名
func VerifyWebhookSignature(apiURL, accessToken string, payload verifyWebhookPayload, timeout time.Duration) (bool, error) {
	apiURL = strings.TrimRight(apiURL, "/")
	if apiURL == "" {
		apiURL = "https://api-m.sandbox.paypal.com"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", apiURL+"/v1/notifications/verify-webhook-signature", strings.NewReader(string(body)))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result struct {
		VerificationStatus string `json:"verification_status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.VerificationStatus == "SUCCESS", nil
}

type verifyWebhookPayload struct {
	AuthAlgo       string          `json:"auth_algo"`
	CertURL        string          `json:"cert_url"`
	TransmissionID string          `json:"transmission_id"`
	TransmissionSig string         `json:"transmission_sig"`
	TransmissionTime string        `json:"transmission_time"`
	WebhookID      string          `json:"webhook_id"`
	WebhookEvent   json.RawMessage `json:"webhook_event"`
}
