package utils

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"fst/backend/internal/config"
	"net"
	"net/smtp"
	"strings"
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

	// 调试：打印配置
	fmt.Printf("[DEBUG] SystemEmailName: '%s', SystemEmail: '%s', SMTPUser: '%s'\n",
		cfg.SystemEmailName, cfg.SystemEmail, cfg.SMTPUser)

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

	var err error
	if cfg.SMTPSSL {
		err = sendEmailSSL(cfg.SMTPHost, cfg.SMTPPort, fromEmail, msg.To, message, auth)
	} else {
		addr := net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort)
		err = smtp.SendMail(addr, auth, fromEmail, []string{msg.To}, []byte(message))
	}

	return err
}

func sendEmailSSL(host, port, from, to, message string, auth smtp.Auth) error {
	addr := net.JoinHostPort(host, port)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

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
