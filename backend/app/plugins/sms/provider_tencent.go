package sms_plugin

import (
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"net/http"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type TencentProvider struct {
	config     services.SMSConfig
	httpClient *http.Client
}

func NewTencentProvider(cfg services.SMSConfig) *TencentProvider {
	return &TencentProvider{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *TencentProvider) Name() string { return "tencent" }

func (p *TencentProvider) IsConfigured() bool {
	return p.config.AccessKey != "" && p.config.SecretKey != "" && p.config.SignName != "" && p.config.SdkAppID != ""
}

func (p *TencentProvider) Send(phone, content string) error {
	return fmt.Errorf("use SendCode for templated SMS")
}

func (p *TencentProvider) SendCode(phone, code string, expireMinutes int, templateParams map[string]string, lang string) error {
	if !p.IsConfigured() {
		return fmt.Errorf("tencent SMS not configured")
	}

	templateName, payload, order := normalizeTemplateParams(code, expireMinutes, templateParams)
	templateCode := p.templateCodeForLang(lang)

	credential := common.NewCredential(p.config.AccessKey, p.config.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 10
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	cpf.SignMethod = "HmacSHA1"

	region := p.config.Region
	if region == "" {
		region = "ap-guangzhou"
	}

	client, err := tencentSms.NewClient(credential, region, cpf)
	if err != nil {
		p.log(phone, templateName, lang, templateCode, 0, err.Error(), "", "")
		return err
	}

	request := tencentSms.NewSendSmsRequest()
	request.SmsSdkAppId = common.StringPtr(p.config.SdkAppID)
	request.SignName = common.StringPtr(p.config.SignName)
	request.TemplateId = common.StringPtr(templateCode)

	templateValues := make([]string, 0, len(order))
	for _, key := range order {
		templateValues = append(templateValues, payload[key])
	}
	request.TemplateParamSet = common.StringPtrs(templateValues)

	e164phone := phone
	if !strings.HasPrefix(e164phone, "+") {
		e164phone = "+86" + phone
	}
	request.PhoneNumberSet = common.StringPtrs([]string{e164phone})

	resp, err := client.SendSms(request)
	if err != nil {
		p.log(phone, templateName, lang, templateCode, 0, err.Error(), "", "")
		return err
	}

	respBytes, _ := json.Marshal(resp)
	respStr := string(respBytes)

	if resp.Response != nil && resp.Response.RequestId != nil {
		requestId := *resp.Response.RequestId
		if len(resp.Response.SendStatusSet) > 0 {
			status := resp.Response.SendStatusSet[0]
			if status.Code != nil && *status.Code == "Ok" {
				p.log(phone, templateName, lang, templateCode, 1, "", requestId, respStr)
				return nil
			}
			var codeStr, msg string
			if status.Code != nil {
				codeStr = *status.Code
			}
			if status.Message != nil {
				msg = *status.Message
			}
			p.log(phone, templateName, lang, templateCode, 0, codeStr+": "+msg, requestId, respStr)
			return fmt.Errorf("tencent SMS failed: %s - %s", codeStr, msg)
		}
	}

	p.log(phone, templateName, lang, templateCode, 0, "unknown error", "", respStr)
	return fmt.Errorf("tencent SMS unknown error")
}

func (p *TencentProvider) templateCodeForLang(lang string) string {
	normalized := normalizeSMSLang(lang)
	if normalized == "en-US" && strings.TrimSpace(p.config.TemplateCodeEN) != "" {
		return p.config.TemplateCodeEN
	}
	return p.config.TemplateCode
}

func (p *TencentProvider) log(phone, templateName, lang, templateCode string, status uint8, errMsg, requestID, resp string) {
	content := fmt.Sprintf("code sent to %s", models.MaskPhone(phone))
	models.CreateSMSLog(&models.SMSLog{
		Phone:        models.MaskPhone(phone),
		Provider:     "tencent",
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

