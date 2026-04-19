package sms_plugin

import (
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AliyunProvider struct {
	config services.SMSConfig
}

func NewAliyunProvider(cfg services.SMSConfig) *AliyunProvider {
	return &AliyunProvider{config: cfg}
}

func (p *AliyunProvider) Name() string { return "aliyun" }

func (p *AliyunProvider) IsConfigured() bool {
	return p.config.AccessKey != "" && p.config.SecretKey != "" && p.config.SignName != ""
}

func (p *AliyunProvider) Send(phone, content string) error {
	return fmt.Errorf("use SendCode for templated SMS")
}

func (p *AliyunProvider) SendCode(phone, code string, expireMinutes int, templateParams map[string]string, lang string) error {
	if !p.IsConfigured() {
		return fmt.Errorf("aliyun SMS not configured")
	}

	region := p.config.Region
	if region == "" {
		region = "cn-hangzhou"
	}

	templateName, templatePayload, _ := normalizeTemplateParams(code, expireMinutes, templateParams)
	templateCode := p.templateCodeForLang(lang)
	templateJSON, err := json.Marshal(templatePayload)
	if err != nil {
		p.log(phone, templateName, lang, templateCode, 0, err.Error(), "", "")
		return err
	}

	cfg := &openapi.Config{
		AccessKeyId:     tea.String(p.config.AccessKey),
		AccessKeySecret: tea.String(p.config.SecretKey),
		RegionId:        tea.String(region),
	}
	cfg.Endpoint = tea.String("dysmsapi.aliyuncs.com")

	client, err := dysmsapi20170525.NewClient(cfg)
	if err != nil {
		p.log(phone, templateName, lang, templateCode, 0, err.Error(), "", "")
		return err
	}

	request := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(p.config.SignName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(string(templateJSON)),
	}

	resp, err := client.SendSms(request)
	if err != nil {
		p.log(phone, templateName, lang, templateCode, 0, err.Error(), "", "")
		return err
	}

	respBytes, _ := json.Marshal(resp)
	respStr := string(respBytes)
	requestID := tea.StringValue(resp.Body.RequestId)
	if strings.EqualFold(tea.StringValue(resp.Body.Code), "OK") {
		p.log(phone, templateName, lang, templateCode, 1, "", requestID, respStr)
		return nil
	}

	msg := tea.StringValue(resp.Body.Message)
	if msg == "" {
		msg = tea.StringValue(resp.Body.Code)
	}
	p.log(phone, templateName, lang, templateCode, 0, msg, requestID, respStr)
	return fmt.Errorf("aliyun SMS failed: %s", msg)
}

func (p *AliyunProvider) templateCodeForLang(lang string) string {
	normalized := normalizeSMSLang(lang)
	if normalized == "en-US" && strings.TrimSpace(p.config.TemplateCodeEN) != "" {
		return p.config.TemplateCodeEN
	}
	return p.config.TemplateCode
}

func (p *AliyunProvider) log(phone, templateName, lang, templateCode string, status uint8, errMsg, requestID, resp string) {
	content := fmt.Sprintf("code sent to %s", models.MaskPhone(phone))
	models.CreateSMSLog(&models.SMSLog{
		Phone:        models.MaskPhone(phone),
		Provider:     "aliyun",
		TemplateCode: templateCode,
		TemplateName: templateName,
		Lang:         lang,
		Content:      content,
		Status:       status,
		ErrorMsg:     errMsg,
		RequestID:    requestID,
		Response:     resp,
	})
}

