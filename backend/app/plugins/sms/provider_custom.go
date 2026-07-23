package sms_plugin

import (
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CustomProvider struct {
	config     services.SMSConfig
	httpClient *http.Client
}

// newSSRFSafeHTTPClient 创建带 SSRF 防护的 HTTP 客户端：
// 最多跟随 2 次重定向，且每次重定向目标都重新走 ValidateOutboundURL。
func newSSRFSafeHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 限制跳转次数，避免无限重定向与跳转链绕过
			if len(via) >= 2 {
				return fmt.Errorf("stopped after 2 redirects (SSRF protection)")
			}
			if req == nil || req.URL == nil {
				return fmt.Errorf("redirect URL is empty")
			}
			if err := ValidateOutboundURL(req.URL.String()); err != nil {
				return fmt.Errorf("redirect blocked: %w", err)
			}
			return nil
		},
	}
}

func NewCustomProvider(cfg services.SMSConfig) *CustomProvider {
	return &CustomProvider{
		config:     cfg,
		httpClient: newSSRFSafeHTTPClient(10 * time.Second),
	}
}

// ValidateOutboundURL 校验自定义短信网关等出站 URL，防止 SSRF。
// 规则：仅允许 http/https；解析 IP 或 DNS；拒绝回环/私网/链路本地/云元数据地址。
func ValidateOutboundURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("endpoint URL is empty")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q: only http/https allowed", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("URL host is empty")
	}
	// 显式拒绝常见本机主机名（即便 DNS 被劫持也先挡一层）
	if strings.EqualFold(host, "localhost") {
		return fmt.Errorf("URL host %q is blocked (loopback)", host)
	}

	ips, err := resolveOutboundHostIPs(host)
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if isBlockedOutboundIP(ip) {
			return fmt.Errorf("URL resolves to blocked address %s", ip.String())
		}
	}
	return nil
}

// resolveOutboundHostIPs 将主机名解析为 IP 列表；已是字面量 IP 则直接返回。
func resolveOutboundHostIPs(host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup failed for %q: %w", host, err)
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no IP addresses for host %q", host)
	}
	return addrs, nil
}

// isBlockedOutboundIP 判断 IP 是否属于禁止出站访问的地址段（SSRF 黑名单）。
func isBlockedOutboundIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	// 回环 / 私网 / 链路本地 / 未指定 / 组播
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	// 云元数据地址 169.254.169.254（已被链路本地覆盖，此处再显式强调）
	if ip4 := ip.To4(); ip4 != nil && ip4[0] == 169 && ip4[1] == 254 {
		return true
	}
	return false
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
	// 发请求前强制 SSRF 校验，避免自定义 Endpoint 打到内网/元数据
	if err := ValidateOutboundURL(p.config.Endpoint); err != nil {
		return "", fmt.Errorf("SSRF protection: %w", err)
	}

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

