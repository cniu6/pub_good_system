package utils

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"fst/backend/pkg/config"
	"log"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var errSMTPStartTLSUnsupported = errors.New("SMTP server does not support STARTTLS")

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

// SendEmail 发送邮件 (支持 SSL / STARTTLS)
func SendEmail(msg EmailMessage) error {
	cfg := config.GlobalConfig
	if cfg == nil {
		return fmt.Errorf("email config not initialized")
	}
	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	// 使用 SYSTEM_EMAIL_ADDRESS 作为发信人邮箱地址，如果为空则使用 SMTPUser
	fromEmail := cfg.SystemEmail
	if fromEmail == "" {
		fromEmail = cfg.SMTPUser
	}
	if fromEmail == "" {
		return fmt.Errorf("sender email not configured")
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

	var auth smtp.Auth
	if cfg.SMTPUser != "" || cfg.SMTPPass != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	}

	// 重试机制：最多尝试3次，处理 EOF 等瞬时错误
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		err = sendEmailWithBestTLS(cfg, fromEmail, msg.To, message, auth)
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

// sendEmailWithBestTLS 按配置选择加密方式，并对常见端口误配自动回退。
// 465 = 隐式 SSL；587/25 = 明文握手后再 STARTTLS。
// 用户常把「SSL」开在 587 上，会导致 tls handshake 失败，这里自动回退。
func sendEmailWithBestTLS(cfg *config.Config, from, to, message string, auth smtp.Auth) error {
	useSSL := cfg.SMTPSSL
	port := strings.TrimSpace(cfg.SMTPPort)

	// 端口明确时优先按惯例（仍尊重用户开关，失败后再回退）
	if port == "465" && !useSSL {
		log.Printf("[Email] 端口 465 通常需隐式 SSL，将优先尝试 SSL")
		useSSL = true
	}

	if useSSL {
		err := sendEmailSSL(cfg.SMTPHost, cfg.SMTPPort, from, to, message, auth, cfg.SMTPUser, cfg.SMTPPass)
		if err == nil {
			return nil
		}
		// 587 等端口开了 SSL：对端先回明文 SMTP，表现为 handshake 不像 TLS
		if isImplicitTLSMismatch(err) {
			log.Printf("[Email] 隐式 TLS 失败（多见于 587 误开 SSL），回退 STARTTLS: %v", err)
			err2 := sendEmailStartTLS(cfg.SMTPHost, cfg.SMTPPort, from, to, message, auth, cfg.SMTPUser, cfg.SMTPPass)
			if err2 == nil {
				return nil
			}
			if errors.Is(err2, errSMTPStartTLSUnsupported) {
				log.Printf("[Email] 服务器不支持 STARTTLS，回退到普通 SMTP: %v", err2)
				return sendEmailPlain(cfg.SMTPHost, cfg.SMTPPort, from, to, message, auth, cfg.SMTPUser, cfg.SMTPPass)
			}
			return err2
		}
		return err
	}

	err := sendEmailStartTLS(cfg.SMTPHost, cfg.SMTPPort, from, to, message, auth, cfg.SMTPUser, cfg.SMTPPass)
	if err == nil {
		return nil
	}
	if errors.Is(err, errSMTPStartTLSUnsupported) {
		log.Printf("[Email] 服务器不支持 STARTTLS，回退到普通 SMTP: %v", err)
		return sendEmailPlain(cfg.SMTPHost, cfg.SMTPPort, from, to, message, auth, cfg.SMTPUser, cfg.SMTPPass)
	}
	// 465 却关了 SSL：对端期望 TLS，STARTTLS/明文会失败，尝试隐式 SSL
	if port == "465" {
		log.Printf("[Email] STARTTLS 失败且端口为 465，尝试隐式 SSL: %v", err)
		return sendEmailSSL(cfg.SMTPHost, cfg.SMTPPort, from, to, message, auth, cfg.SMTPUser, cfg.SMTPPass)
	}
	return err
}

func isImplicitTLSMismatch(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "does not look like a TLS handshake") ||
		strings.Contains(msg, "first record does not look like a TLS handshake")
}

// smtpOutboundProxyConfig 从全局配置构造出站代理；开关关闭时 Enabled=false（直连）。
func smtpOutboundProxyConfig(cfg *config.Config) ProxyConfig {
	if cfg == nil {
		return ProxyConfig{Enabled: false, Timeout: 30 * time.Second}
	}
	port := 1080
	if p, err := strconv.Atoi(strings.TrimSpace(cfg.SMTPProxyPort)); err == nil && p > 0 {
		port = p
	}
	typ := strings.TrimSpace(cfg.SMTPProxyType)
	if typ == "" {
		typ = ProxyTypeSOCKS5
	}
	return ProxyConfig{
		Enabled:  cfg.SMTPProxyEnabled,
		Type:     typ,
		Host:     strings.TrimSpace(cfg.SMTPProxyHost),
		Port:     port,
		Username: cfg.SMTPProxyUser,
		Password: cfg.SMTPProxyPass,
		Timeout:  30 * time.Second,
	}
}

// dialSMTPConn 按代理开关拨号到 SMTP（plain / STARTTLS 底层 TCP）。
func dialSMTPConn(cfg *config.Config, addr string) (net.Conn, error) {
	proxyCfg := smtpOutboundProxyConfig(cfg)
	if proxyCfg.Enabled {
		log.Printf("[Email] 出站代理: %s", MaskProxyConfig(proxyCfg))
	}
	ctx, cancel := context.WithTimeout(context.Background(), proxyCfg.dialTimeout())
	defer cancel()
	return DialViaProxy(ctx, proxyCfg, "tcp", addr)
}

// dialSMTPTLSConn 隐式 SSL：经代理 TCP 后再做 TLS 握手。
func dialSMTPTLSConn(cfg *config.Config, addr string, tlsCfg *tls.Config) (*tls.Conn, error) {
	proxyCfg := smtpOutboundProxyConfig(cfg)
	if proxyCfg.Enabled {
		log.Printf("[Email] 出站代理(TLS): %s", MaskProxyConfig(proxyCfg))
	}
	ctx, cancel := context.WithTimeout(context.Background(), proxyCfg.dialTimeout())
	defer cancel()
	return DialTLSViaProxy(ctx, proxyCfg, "tcp", addr, tlsCfg)
}

func sendEmailPlain(host, port, from, to, message string, auth smtp.Auth, smtpUser, smtpPass string) error {
	addr := net.JoinHostPort(host, port)
	log.Printf("[Email] 连接 %s (plain SMTP)...", addr)
	conn, err := dialSMTPConn(config.GlobalConfig, addr)
	if err != nil {
		return fmt.Errorf("SMTP连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP客户端创建失败: %w", err)
	}
	defer client.Close()

	return sendSMTPMessage(client, from, to, message, auth, smtpUser, smtpPass)
}

func sendSMTPMessage(client *smtp.Client, from, to, message string, auth smtp.Auth, smtpUser, smtpPass string) error {
	if auth != nil {
		log.Printf("[Email] 开始认证...")
		if err := client.Auth(auth); err != nil {
			log.Printf("[Email] AUTH 失败: %v, 尝试 LOGIN 方式...", err)
			// PlainAuth 失败时回退到 LOGIN 认证
			loginA := LoginAuth(smtpUser, smtpPass)
			if err = client.Auth(loginA); err != nil {
				return fmt.Errorf("认证失败: %w", err)
			}
		}
		log.Printf("[Email] 认证成功")
	} else {
		log.Printf("[Email] 未配置 SMTP 认证信息，跳过认证")
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM 失败: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
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

func sendEmailStartTLS(host, port, from, to, message string, auth smtp.Auth, smtpUser, smtpPass string) error {
	addr := net.JoinHostPort(host, port)
	tlsconfig := &tls.Config{
		ServerName: host,
	}

	log.Printf("[Email] 连接 %s (STARTTLS)...", addr)
	conn, err := dialSMTPConn(config.GlobalConfig, addr)
	if err != nil {
		return fmt.Errorf("SMTP连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP客户端创建失败: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("%w", errSMTPStartTLSUnsupported)
	}

	if err = client.StartTLS(tlsconfig); err != nil {
		return fmt.Errorf("STARTTLS失败: %w", err)
	}

	return sendSMTPMessage(client, from, to, message, auth, smtpUser, smtpPass)
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

func sendEmailSSL(host, port, from, to, message string, auth smtp.Auth, smtpUser, smtpPass string) error {
	addr := net.JoinHostPort(host, port)
	tlsconfig := &tls.Config{
		ServerName: host,
	}

	log.Printf("[Email] 连接 %s (TLS)...", addr)
	conn, err := dialSMTPTLSConn(config.GlobalConfig, addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %w", err)
	}
	closeConn := true
	defer func() {
		if closeConn {
			_ = conn.Close()
		}
	}()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP客户端创建失败: %w", err)
	}
	closeConn = false
	defer client.Close()

	return sendSMTPMessage(client, from, to, message, auth, smtpUser, smtpPass)
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

