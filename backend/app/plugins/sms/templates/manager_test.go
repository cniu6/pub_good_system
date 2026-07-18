package templates_test

import (
	"testing"

	"fst/backend/app/plugins/sms/templates"
)

func TestTemplateManager(t *testing.T) {
	m := templates.NewManager()
	m.Add(&templates.Template{
		Name: "register", Lang: "zh-CN", SignName: "测试", Content: "验证码{code}，{expire}分钟",
	})

	tpl := m.Get("register", "zh-CN")
	if tpl == nil || tpl.SignName != "测试" {
		t.Fatalf("Get zh-CN 失败: %+v", tpl)
	}

	tpl2 := m.Get("register", "fr-FR")
	if tpl2 == nil {
		t.Fatal("应回退到 zh-CN")
	}

	content, sign := m.Render("register", "zh-CN", "123456", "5")
	if content == "" || sign != "测试" {
		t.Fatalf("Render: content=%q sign=%q", content, sign)
	}
	if content != "验证码123456，5分钟" {
		t.Fatalf("Render 内容不对: %q", content)
	}

	content2, _ := m.Render("missing", "zh-CN", "111111", "10")
	if content2 == "" {
		t.Fatal("缺模板时应有默认文案")
	}
}
