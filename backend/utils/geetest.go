package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type GeetestValidateRequest struct {
	LotNumber     string `json:"lot_number"`
	CaptchaOutput string `json:"captcha_output"`
	PassToken     string `json:"pass_token"`
	GenTime       string `json:"gen_time"`
	CaptchaID     string `json:"captcha_id"`
}

type GeetestValidateResponse struct {
	Result string `json:"result"`
	Reason string `json:"reason"`
}

func ValidateGeetest(captchaID, captchaKey string, req GeetestValidateRequest) (bool, error) {
	if req.CaptchaID != captchaID {
		return false, errors.New("captcha_id mismatch")
	}

	// Generate sign_token
	h := hmac.New(sha256.New, []byte(captchaKey))
	h.Write([]byte(req.LotNumber))
	signToken := hex.EncodeToString(h.Sum(nil))

	// Prepare request
	apiURL := "https://gcaptcha4.geetest.com/validate"
	params := url.Values{}
	params.Set("lot_number", req.LotNumber)
	params.Set("captcha_output", req.CaptchaOutput)
	params.Set("pass_token", req.PassToken)
	params.Set("gen_time", req.GenTime)
	params.Set("captcha_id", req.CaptchaID)
	params.Set("sign_token", signToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.PostForm(apiURL, params)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result GeetestValidateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}

	if result.Result == "success" {
		return true, nil
	}

	return false, errors.New(result.Reason)
}
