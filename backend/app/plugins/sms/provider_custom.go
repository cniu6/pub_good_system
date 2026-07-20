package sms_plugin

import (
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CustomProvider struct {
	config     services.SMSConfig
	httpClient *http.Client
}

func NewCustomProvider(cfg services.SMSConfig) *CustomProvider {
	return &CustomProvider{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *CustomProvider) Name() string { return "custom" }

func (p *CustomProvider) IsConfigured() bool {
	return p.config.Endpoint != ""
}

func (p *CustomProvider) Send(phone, content string) error {
	if !p.IsConfigured() {
		return fmt.Errorf("custom SMS endpoint not configured")
	}

	body, err := p.buildBody(phone, content, nil)
	if err != nil {
		p.log(0, phone, "send", "", 0, err.Error(), "", "")
		return err
	}

	resp, err := p.doPost(body)
	if err != nil {
		p.log(0, phone, "send", "", 0, err.Error(), "", "")
		return err
	}

	if p.isSuccess(resp) {
		p.log(0, phone, "send", "", 1, "", "", resp)
		return nil
	}
	p.log(0, phone, "send", "", 0, "provider returned failure", "", resp)
	return fmt.Errorf("custom SMS provider returned failure: %s", resp)
}

func (p *CustomProvider) SendCode(phone, code string, expireMinutes int, templateParams map[string]string, lang string) error {
	if !p.IsConfigured() {
		return fmt.Errorf("custom SMS endpoint not configured")
	}

	userID := ExtractMetaUserID(templateParams)
	templateName, params, _ := normalizeTemplateParams(code, expireMinutes, templateParams)
	templateMgr := getTemplateManager()
	content, signName := templateMgr.Render(templateName, normalizeSMSLang(lang), code, fmt.Sprintf("%d", expireMinutes))
	if signName != "" {
		p.config.SignName = signName
	}

	body, err := p.buildBody(phone, content, params)
	if err != nil {
		p.log(userID, phone, templateName, lang, 0, err.Error(), "", "")
		return err
	}

	resp, err := p.doPost(body)
	if err != nil {
		p.log(userID, phone, templateName, lang, 0, err.Error(), "", "")
		return err
	}

	if p.isSuccess(resp) {
		p.log(userID, phone, templateName, lang, 1, "", "", resp)
		return nil
	}
	p.log(userID, phone, templateName, lang, 0, "provider returned failure", "", resp)
	return fmt.Errorf("custom SMS provider returned failure: %s", resp)
}

func (p *CustomProvider) buildBody(phone, content string, params map[string]string) (string, error) {
	if p.getBodyFormat() == "json" {
		payload := map[string]interface{}{
			"phone":   phone,
			"content": content,
			"sign":    p.config.SignName,
		}
		if params != nil {
			payload["params"] = params
		}
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	}

	form := url.Values{}
	form.Set("phone", phone)
	form.Set("content", content)
	form.Set("sign", p.config.SignName)
	for k, v := range params {
		if strings.TrimSpace(k) != "" {
			form.Set(k, v)
		}
	}
	return form.Encode(), nil
}

func (p *CustomProvider) doPost(body string) (string, error) {
	req, err := http.NewRequest("POST", p.config.Endpoint, strings.NewReader(body))
	if err != nil {
		return "", err
	}

	if p.getBodyFormat() == "json" {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if p.config.SecretKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.SecretKey)
	}
	if p.config.AccessKey != "" {
		req.Header.Set("X-Api-Key", p.config.AccessKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return string(respBody), nil
}

func (p *CustomProvider) getBodyFormat() string {
	format := strings.ToLower(strings.TrimSpace(p.config.BodyFormat))
	if format == "form" {
		return "form"
	}
	return "json"
}

func (p *CustomProvider) isSuccess(resp string) bool {
	resp = strings.TrimSpace(resp)
	if resp == "" {
		return true
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(resp), &payload); err == nil {
		if code, ok := payload["code"]; ok {
			switch v := code.(type) {
			case float64:
				if v == 0 {
					return true
				}
			case string:
				lowerCode := strings.ToLower(strings.TrimSpace(v))
				if lowerCode == "0" || lowerCode == "ok" || lowerCode == "success" {
					return true
				}
			}
		}
		if errCode, ok := payload["errcode"]; ok {
			switch v := errCode.(type) {
			case float64:
				if v == 0 {
					return true
				}
			case string:
				lowerErrCode := strings.ToLower(strings.TrimSpace(v))
				if lowerErrCode == "0" || lowerErrCode == "ok" || lowerErrCode == "success" {
					return true
				}
			}
		}
	}

	lower := strings.ToLower(resp)
	return strings.HasPrefix(lower, "success") ||
		strings.HasPrefix(lower, "ok") ||
		strings.HasPrefix(lower, `"success`) ||
		strings.HasPrefix(lower, `"ok`) ||
		strings.HasPrefix(lower, "0") ||
		(strings.Contains(lower, `"code"`) && strings.Contains(lower, `"0"`)) ||
		(strings.Contains(lower, `"errcode"`) && (strings.Contains(lower, `"0"`) || strings.Contains(lower, `"ok"`)))
}

func (p *CustomProvider) log(userID uint64, phone, templateName, lang string, status uint8, errMsg, requestID, resp string) {
	content := fmt.Sprintf("code sent to %s", models.MaskPhone(phone))
	models.CreateSMSLog(&models.SMSLog{
		UserID:       userID,
		Phone:        models.MaskPhone(phone),
		Provider:     "custom",
		TemplateCode: p.config.TemplateCode,
		TemplateName: templateName,
		Lang:         lang,
		Content:      content,
		Status:       status,
		ErrorMsg:     errMsg,
		RequestID:    requestID,
		Response:     resp,
	})
}

