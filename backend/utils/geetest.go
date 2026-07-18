package utils

// geetest 验证码验证工具
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

	"github.com/gin-gonic/gin"
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

// ValidateGeetestFromHeaders 从请求头提取极验参数并校验。
// enabled=false 时直接通过；失败返回 error。
func ValidateGeetestFromHeaders(c *gin.Context, captchaID, captchaKey string, enabled bool) error {
	if !enabled {
		return nil
	}

	req := GeetestValidateRequest{
		LotNumber:     c.GetHeader("X-Geetest-Lot-Number"),
		CaptchaOutput: c.GetHeader("X-Geetest-Captcha-Output"),
		PassToken:     c.GetHeader("X-Geetest-Pass-Token"),
		GenTime:       c.GetHeader("X-Geetest-Gen-Time"),
		CaptchaID:     c.GetHeader("X-Geetest-Captcha-Id"),
	}

	valid, err := ValidateGeetest(captchaID, captchaKey, req)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("captcha validation failed")
	}
	return nil
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
	apiURL := "https://gcaptcha4.geetest.com/validate" //geetest 验证地址,所以机器必须要连外网
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
