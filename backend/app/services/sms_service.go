package services

import (
	"fmt"
	"log"
	"sync"
)

// ========================================
// SMS Provider 接口定义
// ========================================

// SMSProvider 短信服务商接口
// 对接新的短信服务只需实现此接口
type SMSProvider interface {
	// Name 返回服务商名称
	Name() string
	// Send 发送短信
	Send(phone, content string) error
	// SendCode 发送验证码短信
	SendCode(phone, code string, expireMinutes int) error
	// IsConfigured 检查是否已正确配置
	IsConfigured() bool
}

// SMSConfig 短信服务配置
type SMSConfig struct {
	Provider  string // 服务商标识: aliyun, tencent, custom
	AccessKey string // AccessKey / API Key
	SecretKey string // SecretKey / API Secret
	SignName  string // 短信签名
	TemplateCode string // 验证码模板ID
	Region   string // 区域（部分服务商需要）
	Endpoint string // 自定义端点（自定义服务商）
}

// ========================================
// SMS 服务管理
// ========================================

// SMSService 短信服务（管理多个 Provider）
type SMSService struct {
	mu       sync.RWMutex
	provider SMSProvider
	config   SMSConfig
}

// GlobalSMSService 全局短信服务实例
var GlobalSMSService *SMSService

// InitSMSService 初始化全局短信服务
func InitSMSService() {
	GlobalSMSService = &SMSService{}
	smsConfig := GetGlobalSMSRuntimeConfig()
	GlobalSMSService.SetConfig(SMSConfig{
		Provider:     smsConfig.Provider,
		AccessKey:    smsConfig.AccessKey,
		SecretKey:    smsConfig.SecretKey,
		SignName:     smsConfig.SignName,
		TemplateCode: smsConfig.TemplateCode,
		Region:       smsConfig.Region,
	})
	log.Println("[SMSService] Initialized (no provider configured)")
}

// SetProvider 设置当前使用的短信服务商
func (s *SMSService) SetProvider(provider SMSProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.provider = provider
	if provider != nil {
		log.Printf("[SMSService] Provider set to: %s\n", provider.Name())
	}
}

// SetConfig 设置短信配置并自动选择 Provider
func (s *SMSService) SetConfig(cfg SMSConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg

	// 根据 provider 标识自动选择实现
	switch cfg.Provider {
	case "aliyun":
		s.provider = &AliyunSMSProvider{config: cfg}
	case "tencent":
		s.provider = &TencentSMSProvider{config: cfg}
	default:
		// 未配置或未知的 provider，使用日志占位
		s.provider = &ConsoleSMSProvider{}
	}

	log.Printf("[SMSService] Config updated, provider: %s\n", s.provider.Name())
}

// GetConfig 获取当前配置
func (s *SMSService) GetConfig() SMSConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// Send 发送短信
func (s *SMSService) Send(phone, content string) error {
	s.mu.RLock()
	provider := s.provider
	s.mu.RUnlock()

	if provider == nil {
		log.Printf("[SMSService] No provider configured, skipping SMS to %s\n", phone)
		return nil
	}

	return provider.Send(phone, content)
}

// SendCode 发送验证码
func (s *SMSService) SendCode(phone, code string, expireMinutes int) error {
	s.mu.RLock()
	provider := s.provider
	s.mu.RUnlock()

	if provider == nil {
		log.Printf("[SMSService] No provider configured, skipping code SMS to %s\n", phone)
		return nil
	}

	return provider.SendCode(phone, code, expireMinutes)
}

// IsConfigured 检查短信服务是否已配置
func (s *SMSService) IsConfigured() bool {
	s.mu.RLock()
	provider := s.provider
	s.mu.RUnlock()

	if provider == nil {
		return false
	}
	return provider.IsConfigured()
}

// GetProviderName 获取当前 Provider 名称
func (s *SMSService) GetProviderName() string {
	s.mu.RLock()
	provider := s.provider
	s.mu.RUnlock()

	if provider == nil {
		return "none"
	}
	return provider.Name()
}

// ========================================
// 控制台占位 Provider（开发环境使用）
// ========================================

// ConsoleSMSProvider 控制台日志占位（不实际发送短信）
type ConsoleSMSProvider struct{}

func (p *ConsoleSMSProvider) Name() string { return "console" }

func (p *ConsoleSMSProvider) Send(phone, content string) error {
	fmt.Printf("[SMS-Console] To: %s, Content: %s\n", phone, content)
	return nil
}

func (p *ConsoleSMSProvider) SendCode(phone, code string, expireMinutes int) error {
	fmt.Printf("[SMS-Console] To: %s, Code: %s, Expires: %d min\n", phone, code, expireMinutes)
	return nil
}

func (p *ConsoleSMSProvider) IsConfigured() bool { return true }

// ========================================
// 阿里云短信 Provider（预留）
// ========================================

// AliyunSMSProvider 阿里云短信服务
type AliyunSMSProvider struct {
	config SMSConfig
}

func (p *AliyunSMSProvider) Name() string { return "aliyun" }

func (p *AliyunSMSProvider) Send(phone, content string) error {
	// TODO: 对接阿里云短信 SDK
	// sdk: github.com/alibabacloud-go/dysmsapi-20170525/v3/client
	// 1. 创建 client: dysmsapi.NewClient(...)
	// 2. 构建请求: dysmsapi.SendSmsRequest{PhoneNumbers: phone, SignName: p.config.SignName, ...}
	// 3. 发送: client.SendSms(request)
	fmt.Printf("[SMS-Aliyun] TODO: Send to %s, content: %s\n", phone, content)
	return fmt.Errorf("aliyun SMS provider not implemented yet")
}

func (p *AliyunSMSProvider) SendCode(phone, code string, expireMinutes int) error {
	// TODO: 对接阿里云验证码模板
	// templateParam := fmt.Sprintf(`{"code":"%s"}`, code)
	// 使用 p.config.TemplateCode 作为模板ID
	fmt.Printf("[SMS-Aliyun] TODO: SendCode to %s, code: %s, template: %s\n", phone, code, p.config.TemplateCode)
	return fmt.Errorf("aliyun SMS provider not implemented yet")
}

func (p *AliyunSMSProvider) IsConfigured() bool {
	return p.config.AccessKey != "" && p.config.SecretKey != "" && p.config.SignName != "" && p.config.TemplateCode != ""
}

// ========================================
// 腾讯云短信 Provider（预留）
// ========================================

// TencentSMSProvider 腾讯云短信服务
type TencentSMSProvider struct {
	config SMSConfig
}

func (p *TencentSMSProvider) Name() string { return "tencent" }

func (p *TencentSMSProvider) Send(phone, content string) error {
	// TODO: 对接腾讯云短信 SDK
	// sdk: github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111
	// 1. 创建 credential: common.NewCredential(p.config.AccessKey, p.config.SecretKey)
	// 2. 创建 client: sms.NewClient(credential, p.config.Region, ...)
	// 3. 构建请求: sms.SendSmsRequest{...}
	// 4. 发送: client.SendSms(request)
	fmt.Printf("[SMS-Tencent] TODO: Send to %s, content: %s\n", phone, content)
	return fmt.Errorf("tencent SMS provider not implemented yet")
}

func (p *TencentSMSProvider) SendCode(phone, code string, expireMinutes int) error {
	// TODO: 对接腾讯云验证码模板
	// templateParamSet := []*string{&code, fmt.Sprintf("%d", expireMinutes)}
	fmt.Printf("[SMS-Tencent] TODO: SendCode to %s, code: %s, template: %s\n", phone, code, p.config.TemplateCode)
	return fmt.Errorf("tencent SMS provider not implemented yet")
}

func (p *TencentSMSProvider) IsConfigured() bool {
	return p.config.AccessKey != "" && p.config.SecretKey != "" && p.config.SignName != "" && p.config.TemplateCode != ""
}
