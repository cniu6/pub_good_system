//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

// 临时脚本：强制更新所有邮件模板为最新的 HTML 格式

func main() {
	// 加载 .env 配置
	if err := godotenv.Load(".env"); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "root")
	dbName := getEnv("DB_NAME", "fst_platform")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	fmt.Println("========================================")
	fmt.Println("开始强制更新邮件模板为新 HTML 格式...")
	fmt.Println("========================================")

	// 注册验证码模板 - 中文
	registerCodeZH := `<p style="margin:0 0 16px 0;">您好，感谢您的注册！请使用以下验证码完成验证：</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<div style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:32px;font-weight:700;letter-spacing:8px;padding:16px 40px;border-radius:12px;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ 验证码有效期为 <strong>{expire_minutes} 分钟</strong>，请尽快使用。</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">如果这不是您本人的操作，请忽略此邮件。请勿将验证码透露给任何人。</p>`

	// 注册验证码模板 - 英文
	registerCodeEN := `<p style="margin:0 0 16px 0;">Hello! Thank you for signing up. Please use the following code to verify your account:</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<div style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:32px;font-weight:700;letter-spacing:8px;padding:16px 40px;border-radius:12px;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ This code is valid for <strong>{expire_minutes} minutes</strong>.</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">If you did not request this, please ignore this email. Never share your code with anyone.</p>`

	// 密码重置模板 - 中文
	resetPasswordZH := `<p style="margin:0 0 16px 0;">您好，我们收到了您的密码重置请求。请点击下方按钮重置密码：</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<a href="{link}" style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;padding:14px 48px;border-radius:10px;">重置密码</a>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">如果按钮无法点击，您也可以使用以下验证码：</p>` +
		`<div style="text-align:center;margin:20px 0;">` +
		`<div style="display:inline-block;background:#f0f2f5;font-size:28px;font-weight:700;letter-spacing:6px;padding:14px 36px;border-radius:10px;color:#1a1a2e;border:2px dashed #667eea;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ 有效期为 <strong>15 分钟</strong>，请尽快操作。</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">如果这不是您本人的操作，请忽略此邮件，您的密码不会被更改。</p>`

	// 密码重置模板 - 英文
	resetPasswordEN := `<p style="margin:0 0 16px 0;">Hello, we received a request to reset your password. Click the button below to proceed:</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<a href="{link}" style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;padding:14px 48px;border-radius:10px;">Reset Password</a>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">If the button doesn't work, you can also use this verification code:</p>` +
		`<div style="text-align:center;margin:20px 0;">` +
		`<div style="display:inline-block;background:#f0f2f5;font-size:28px;font-weight:700;letter-spacing:6px;padding:14px 36px;border-radius:10px;color:#1a1a2e;border:2px dashed #667eea;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ Valid for <strong>15 minutes</strong>.</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">If you did not request a password reset, please ignore this email. Your password will remain unchanged.</p>`

	type tplUpdate struct {
		Name    string
		Lang    string
		Subject string
		Content string
		Vars    string
	}

	updates := []tplUpdate{
		{Name: "register_code", Lang: "zh-CN", Subject: "【{app_name}】注册验证码", Content: registerCodeZH, Vars: "code, app_name, expire_minutes"},
		{Name: "register_code", Lang: "en-US", Subject: "[{app_name}] Registration Code", Content: registerCodeEN, Vars: "code, app_name, expire_minutes"},
		{Name: "reset_password", Lang: "zh-CN", Subject: "【{app_name}】密码重置请求", Content: resetPasswordZH, Vars: "link, code, app_name"},
		{Name: "reset_password", Lang: "en-US", Subject: "[{app_name}] Password Reset Request", Content: resetPasswordEN, Vars: "link, code, app_name"},
	}

	query := `UPDATE email_templates SET subject = ?, content = ?, variables = ? WHERE name = ? AND lang = ?`

	for _, u := range updates {
		result, err := db.Exec(query, u.Subject, u.Content, u.Vars, u.Name, u.Lang)
		if err != nil {
			fmt.Printf("❌ 更新失败 [%s/%s]: %v\n", u.Name, u.Lang, err)
			continue
		}
		rows, _ := result.RowsAffected()
		if rows > 0 {
			fmt.Printf("✅ 已更新: %s (%s)\n", u.Name, u.Lang)
		} else {
			fmt.Printf("⚠️  未找到模板: %s (%s)，跳过\n", u.Name, u.Lang)
		}
	}

	fmt.Println("========================================")
	fmt.Println("邮件模板更新完成!")
	fmt.Println("========================================")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
