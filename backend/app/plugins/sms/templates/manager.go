package templates

import (
	"fmt"
	"sync"
)

// Template 短信模板
type Template struct {
	Name        string
	Lang        string
	SignName    string
	Content     string
	Variables   string
	Description string
}

// Manager 多语言短信模板管理器
type Manager struct {
	mu        sync.RWMutex
	templates map[string]map[string]*Template
}

func NewManager() *Manager {
	return &Manager{
		templates: make(map[string]map[string]*Template),
	}
}

func (m *Manager) Get(name, lang string) *Template {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if langTemplates, ok := m.templates[lang]; ok {
		if tpl, ok := langTemplates[name]; ok {
			return tpl
		}
	}
	if lang != "zh-CN" {
		if langTemplates, ok := m.templates["zh-CN"]; ok {
			if tpl, ok := langTemplates[name]; ok {
				return tpl
			}
		}
	}
	if lang != "en-US" {
		if langTemplates, ok := m.templates["en-US"]; ok {
			if tpl, ok := langTemplates[name]; ok {
				return tpl
			}
		}
	}
	return nil
}

func (m *Manager) Render(name, lang, code, expire string) (content string, signName string) {
	tpl := m.Get(name, lang)
	if tpl == nil {
		return fmt.Sprintf("您的验证码是 %s，有效期 %s 分钟", code, expire), ""
	}

	content = tpl.Content
	content = replaceVar(content, "code", code)
	content = replaceVar(content, "expire", expire)
	return content, tpl.SignName
}

func replaceVar(text, name, value string) string {
	text = replaceAll(text, "{"+name+"}", value)
	text = replaceAll(text, "{{"+name+"}}", value)
	text = replaceAll(text, "$"+name, value)
	return text
}

func replaceAll(s, old, new string) string {
	for {
		idx := -1
		for i := 0; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx == -1 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
	return s
}

func (m *Manager) Add(tpl *Template) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.templates[tpl.Lang]; !ok {
		m.templates[tpl.Lang] = make(map[string]*Template)
	}
	m.templates[tpl.Lang][tpl.Name] = tpl
}

func (m *Manager) InitDefaultTemplates() {
	m.Add(&Template{
		Name:        "register_code",
		Lang:        "zh-CN",
		SignName:    "",
		Content:     "【{app_name}】您的注册验证码是 {code}，{expire}分钟内有效。如非本人操作，请忽略此短信。",
		Variables:   "code, expire, app_name",
		Description: "用户注册验证码",
	})
	m.Add(&Template{
		Name:        "register_code",
		Lang:        "en-US",
		SignName:    "",
		Content:     "[{app_name}] Your verification code is {code}, valid for {expire} minutes. If not requested by you, please ignore.",
		Variables:   "code, expire, app_name",
		Description: "User registration verification code",
	})
	m.Add(&Template{
		Name:        "login_code",
		Lang:        "zh-CN",
		SignName:    "",
		Content:     "【{app_name}】您正在登录，验证码是 {code}，{expire}分钟内有效。",
		Variables:   "code, expire, app_name",
		Description: "用户登录验证码",
	})
	m.Add(&Template{
		Name:        "login_code",
		Lang:        "en-US",
		SignName:    "",
		Content:     "[{app_name}] Your login code is {code}, valid for {expire} minutes.",
		Variables:   "code, expire, app_name",
		Description: "User login verification code",
	})
	m.Add(&Template{
		Name:        "reset_password",
		Lang:        "zh-CN",
		SignName:    "",
		Content:     "【{app_name}】您正在重置密码，验证码是 {code}，{expire}分钟内有效。如非本人操作，请忽略。",
		Variables:   "code, expire, app_name",
		Description: "密码重置验证码",
	})
	m.Add(&Template{
		Name:        "reset_password",
		Lang:        "en-US",
		SignName:    "",
		Content:     "[{app_name}] Your password reset code is {code}, valid for {expire} minutes. Ignore if not requested.",
		Variables:   "code, expire, app_name",
		Description: "Password reset verification code",
	})
	m.Add(&Template{
		Name:        "bind_phone",
		Lang:        "zh-CN",
		SignName:    "",
		Content:     "【{app_name}】您正在绑定手机，验证码是 {code}，{expire}分钟内有效。",
		Variables:   "code, expire, app_name",
		Description: "手机绑定验证码",
	})
	m.Add(&Template{
		Name:        "bind_phone",
		Lang:        "en-US",
		SignName:    "",
		Content:     "[{app_name}] Your phone binding code is {code}, valid for {expire} minutes.",
		Variables:   "code, expire, app_name",
		Description: "Phone binding verification code",
	})
}
