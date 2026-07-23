package models_test

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
)

// TestSeedEmailTemplates_DoesNotOverwriteAdminEdit 回归真实 bug：旧实现里 SeedEmailTemplates
// 对已存在的模板会无条件把 content 覆盖回硬编码默认值，导致管理员在后台改过的邮件模板在服务
// 重启（每次 RunAutoMigrate 都会调一遍）后被冲掉。修复后应该是「仅缺失时插入，已存在则不动」。
func TestSeedEmailTemplates_DoesNotOverwriteAdminEdit(t *testing.T) {
	cleanup := testutil.SetupSQLite(t) // SetupSQLite 内部会跑一次迁移，已经种过默认模板
	defer cleanup()

	tpl, err := models.GetEmailTemplate("register_code", "zh-CN")
	if err != nil {
		t.Fatalf("初始种子模板应存在: %v", err)
	}

	// 模拟管理员在后台把内容改成自定义文案
	customContent := "管理员自定义的注册验证码内容 {code}"
	if err := models.UpdateEmailTemplate(tpl.ID, tpl.Subject, customContent, tpl.Description, tpl.Status); err != nil {
		t.Fatalf("UpdateEmailTemplate 失败: %v", err)
	}

	// 再跑一次种子逻辑（等价于服务重启触发的 RunAutoMigrate）
	models.SeedEmailTemplates()

	after, err := models.GetEmailTemplateByID(tpl.ID)
	if err != nil {
		t.Fatalf("GetEmailTemplateByID 失败: %v", err)
	}
	if after.Content != customContent {
		t.Fatalf("管理员编辑的内容被种子逻辑覆盖了：got %q want %q", after.Content, customContent)
	}
}

// TestResetEmailTemplateToDefault_RestoresSeedContent 验证 Reset 恢复的内容与 Init 种子完全一致
// （回归：旧实现里 Controller.Reset 内嵌的默认文案和 models.SeedEmailTemplates 的种子不是同一份，
// 两处「默认模板」对不上）。
func TestResetEmailTemplateToDefault_RestoresSeedContent(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	tpl, err := models.GetEmailTemplate("reset_password", "zh-CN")
	if err != nil {
		t.Fatalf("初始种子模板应存在: %v", err)
	}

	seed, ok := models.GetDefaultEmailTemplateByNameLang("reset_password", "zh-CN")
	if !ok {
		t.Fatalf("应存在 reset_password/zh-CN 的默认种子")
	}

	// 先改掉，再 Reset，验证恢复的内容跟种子严格一致
	if err := models.UpdateEmailTemplate(tpl.ID, "改过的标题", "改过的内容", "改过的描述", tpl.Status); err != nil {
		t.Fatalf("UpdateEmailTemplate 失败: %v", err)
	}
	if err := models.ResetEmailTemplateToDefault(tpl.ID); err != nil {
		t.Fatalf("ResetEmailTemplateToDefault 失败: %v", err)
	}

	after, err := models.GetEmailTemplateByID(tpl.ID)
	if err != nil {
		t.Fatalf("GetEmailTemplateByID 失败: %v", err)
	}
	if after.Content != seed.Content {
		t.Fatalf("Reset 后内容与种子不一致：got %q want %q", after.Content, seed.Content)
	}
	if after.Subject != seed.Subject {
		t.Fatalf("Reset 后主题与种子不一致：got %q want %q", after.Subject, seed.Subject)
	}
}

// TestResetEmailTemplateToDefault_NoDefault 无默认模板的场景应返回明确错误，而不是 panic/静默成功。
func TestResetEmailTemplateToDefault_NoDefault(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	if err := models.CreateEmailTemplate(&models.EmailTemplate{
		Name: "custom_no_default", Lang: "zh-CN",
		Title: "自定义模板", Subject: "s", Content: "c", Status: 1,
	}); err != nil {
		t.Fatalf("CreateEmailTemplate 失败: %v", err)
	}
	tpl, err := models.GetEmailTemplate("custom_no_default", "zh-CN")
	if err != nil {
		t.Fatalf("查询刚创建的模板失败: %v", err)
	}

	err = models.ResetEmailTemplateToDefault(tpl.ID)
	if err == nil {
		t.Fatalf("无默认模板时应返回错误")
	}
}
