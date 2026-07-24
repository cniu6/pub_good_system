package services

import (
	"fmt"
	"fst/backend/utils"
	"log"
	"sync"
)

// SMSConfig 短信服务配置
type SMSConfig struct {
	Provider     string
	AccessKey    string
	SecretKey    string
	SignName     string
	TemplateCode string
	TemplateCodeEN string
	Region       string
	Endpoint     string
	SdkAppID     string
	BodyFormat   string
}

// SMSProvider 短信服务商接口
type SMSProvider interface {
	Name() string
	Send(phone, content string) error
	SendCode(phone, code string, expireMinutes int, templateParams map[string]string, lang string) error
	IsConfigured() bool
}

type SMSService struct {
	mu        sync.RWMutex
	providers map[string]SMSProvider
	config    SMSConfig
}

var GlobalSMSService *SMSService

func InitSMSService() {
	GlobalSMSService = &SMSService{
		providers: make(map[string]SMSProvider),
	}
	GlobalSMSService.providers["console"] = newConsoleProvider()
	log.Println("[SMSService] Initialized with console provider")
}

func (s *SMSService) RegisterProvider(name string, provider SMSProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[name] = provider
	log.Printf("[SMSService] Registered provider: %s", name)
}

func (s *SMSService) SetConfig(cfg SMSConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
}

func (s *SMSService) UseProvider(name string, provider SMSProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.providers {
		if key != "console" {
			delete(s.providers, key)
		}
	}

	if provider == nil {
		log.Printf("[SMSService] Runtime provider cleared")
		return
	}

	s.providers[name] = provider
	log.Printf("[SMSService] Runtime provider set: %s", provider.Name())
}

func (s *SMSService) GetConfig() SMSConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// snapshotProviders 在读锁下拷贝当前 provider 引用，避免在远程调用期间长时间持锁。
func (s *SMSService) snapshotProviders() (active []SMSProvider, console SMSProvider) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for name, provider := range s.providers {
		if name == "console" {
			console = provider
			continue
		}
		active = append(active, provider)
	}
	return active, console
}

func (s *SMSService) Send(phone, content string) error {
	active, console := s.snapshotProviders()
	if len(active) == 0 && console == nil {
		return fmt.Errorf("no SMS provider configured")
	}

	var lastErr error
	for _, provider := range active {
		if err := provider.Send(phone, content); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	if console != nil {
		return console.Send(phone, content)
	}
	return lastErr
}

func (s *SMSService) SendCode(phone, code string, expireMinutes int, templateParams map[string]string, lang string) error {
	// 通道侧一般不要前导 +（如 +86138... → 86138...）
	phone = utils.FormatPhoneForSMS(phone)
	active, console := s.snapshotProviders()
	if len(active) == 0 && console == nil {
		return fmt.Errorf("no SMS provider configured")
	}

	var lastErr error
	for _, provider := range active {
		if err := provider.SendCode(phone, code, expireMinutes, templateParams, lang); err == nil {
			return nil
		} else {
			lastErr = err
			log.Printf("[SMSService] Provider %s failed: %v, trying next...", provider.Name(), err)
		}
	}
	if console != nil {
		return console.SendCode(phone, code, expireMinutes, templateParams, lang)
	}
	return lastErr
}

func (s *SMSService) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for name, provider := range s.providers {
		if name != "console" && provider.IsConfigured() {
			return true
		}
	}
	return false
}

func (s *SMSService) GetProviderName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.providers) == 0 {
		return "none"
	}
	if s.config.Provider != "" {
		if p, ok := s.providers[s.config.Provider]; ok {
			return p.Name()
		}
	}
	for name, p := range s.providers {
		if name != "console" {
			return p.Name()
		}
	}
	return "console"
}

type consoleProvider struct{}

func newConsoleProvider() *consoleProvider { return &consoleProvider{} }

func (p *consoleProvider) Name() string { return "console" }

func (p *consoleProvider) IsConfigured() bool { return true }

func (p *consoleProvider) Send(phone, content string) error {
	// console 通道仅开发调试用：故意明文打印手机号/正文（含验证码），方便本地联调
	log.Printf("[SMS-Console] To: %s, Content: %s", phone, content)
	return nil
}

func (p *consoleProvider) SendCode(phone, code string, expireMinutes int, templateParams map[string]string, lang string) error {
	log.Printf("[SMS-Console] To: %s, Code: %s, Expires: %d min, Lang: %s, Params: %v",
		phone, code, expireMinutes, lang, templateParams)
	return nil
}

