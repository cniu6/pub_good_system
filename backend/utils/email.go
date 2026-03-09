package utils

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"fst/backend/internal/config"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
	"unicode/utf8"
)

// EmailMessage 邮件内容
type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

func encodeRFC2047IfNeeded(s string) string {
	if s == "" {
		return s
	}
	if utf8.ValidString(s) {
		allASCII := true
		for i := 0; i < len(s); i++ {
			if s[i] >= 128 {
				allASCII = false
				break
			}
		}
		if allASCII {
			return s
		}
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(s))
	return fmt.Sprintf("=?UTF-8?B?%s?=", encoded)
}

// SendEmail 发送邮件 (支持 SSL)
func SendEmail(msg EmailMessage) error {
	cfg := config.GlobalConfig
	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	// 使用 SYSTEM_EMAIL_ADDRESS 作为发信人邮箱地址，如果为空则使用 SMTPUser
	fromEmail := cfg.SystemEmail
	if fromEmail == "" {
		fromEmail = cfg.SMTPUser
	}
	// 使用 SYSTEM_EMAIL_NAME 作为发信人名称，如果为空则使用 AppName
	emailName := cfg.SystemEmailName
	if emailName == "" {
		emailName = cfg.AppName
	}

	fromHeader := fmt.Sprintf("%s <%s>", encodeRFC2047IfNeeded(emailName), fromEmail)
	subjectHeader := encodeRFC2047IfNeeded(msg.Subject)

	message := ""
	message += fmt.Sprintf("From: %s\r\n", fromHeader)
	message += fmt.Sprintf("To: %s\r\n", msg.To)
	message += fmt.Sprintf("Subject: %s\r\n", subjectHeader)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n" + msg.Body

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)

	// 重试机制：最多尝试3次，处理 EOF 等瞬时错误
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		if cfg.SMTPSSL {
			err = sendEmailSSL(cfg.SMTPHost, cfg.SMTPPort, fromEmail, msg.To, message, auth)
		} else {
			addr := net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort)
			err = smtp.SendMail(addr, auth, fromEmail, []string{msg.To}, []byte(message))
		}
		if err == nil {
			return nil
		}
		log.Printf("[Email] 第 %d 次发送失败: %v", attempt, err)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
	}

	return fmt.Errorf("发送邮件失败（已重试3次）: %w", err)
}

// loginAuth 实现 LOGIN 认证方式（兼容 Yandex 等邮件服务商）
// Go 内置的 PlainAuth 在 tls.Dial 隐式 SSL 连接上会误判为非加密连接
type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unknown LOGIN challenge: %s", fromServer)
		}
	}
	return nil, nil
}

func sendEmailSSL(host, port, from, to, message string, auth smtp.Auth) error {
	addr := net.JoinHostPort(host, port)
	tlsconfig := &tls.Config{
		ServerName:         host,
	}

	log.Printf("[Email] 连接 %s (TLS)...", addr)
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("SMTP客户端创建失败: %w", err)
	}
	defer client.Close()

	log.Printf("[Email] EHLO 成功，开始认证...")
	if err = client.Auth(auth); err != nil {
		log.Printf("[Email] AUTH 失败: %v, 尝试 LOGIN 方式...", err)
		// PlainAuth 失败时回退到 LOGIN 认证
		loginA := LoginAuth(
			config.GlobalConfig.SMTPUser,
			config.GlobalConfig.SMTPPass,
		)
		if err = client.Auth(loginA); err != nil {
			return fmt.Errorf("认证失败: %w", err)
		}
	}
	log.Printf("[Email] 认证成功")

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM 失败: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO 失败: %w", err)
	}

	log.Printf("[Email] 开始写入邮件数据 (%d bytes)...", len(message))
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA命令失败: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("完成数据传输失败: %w", err)
	}

	log.Printf("[Email] 邮件发送成功")
	return client.Quit()
}

// ReplaceTemplateVars 替换模板变量
func ReplaceTemplateVars(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		placeholder := fmt.Sprintf("{%s}", k)
		result = strings.ReplaceAll(result, placeholder, v)
	}
	return result
}
