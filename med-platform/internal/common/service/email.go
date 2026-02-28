package service

import (
	"bytes"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/smtp"
	// ❌ 删除了对 sysconfig 的导入，彻底打破循环
)

//go:embed templates/*.html
var emailTemplates embed.FS

// 🔥 核心修复：定义一个全局函数变量，用于由外部注入配置获取逻辑
var ConfigProvider func(string) string

// EmailData 动态数据结构
type EmailData struct {
	Username    string
	AdminAvatar string
	SenderName  string
	MagicLink   string
	Subject     string
	Body        template.HTML
}

// SendVerificationEmail 发送验证邮件
func SendVerificationEmail(toEmail, username, token, emailType string) error {
	host, port, user, pass, frontendURL, senderName := getEmailConfig()
	if host == "" { return fmt.Errorf("系统发信配置缺失") }

	adminQQ := "2219911811"
	data := EmailData{
		Username:    username,
		AdminAvatar: fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=640", adminQQ),
		SenderName:  senderName,
	}

	var subject, templateName string
	if emailType == "register" {
		subject = fmt.Sprintf("【%s】欢迎注册！请激活您的账号", senderName)
		data.Subject = "欢迎加入 🚀"
		data.MagicLink = fmt.Sprintf("%s/verify-email?token=%s&type=register", frontendURL, token)
		templateName = "templates/register.html"
	} else if emailType == "change_email" {
		subject = fmt.Sprintf("【%s】请确认您的新邮箱绑定", senderName)
		data.Subject = "邮箱换绑确认 🔐"
		data.MagicLink = fmt.Sprintf("%s/verify-email?token=%s&type=change_email", frontendURL, token)
		templateName = "templates/change_email.html"
	}

	t, err := template.ParseFS(emailTemplates, templateName)
	if err != nil { return err }
	var bodyBuffer bytes.Buffer
	if err := t.Execute(&bodyBuffer, data); err != nil { return err }

	return sendSMTP(toEmail, subject, bodyBuffer.String(), host, port, user, pass, senderName)
}

// SendCustomEmail 发送管理员自定义模版邮件
func SendCustomEmail(toEmail, username, subject, htmlContent string) error {
	host, port, user, pass, _, senderName := getEmailConfig()
	if host == "" { return fmt.Errorf("系统发信配置缺失") }

	adminQQ := "2219911811"
	data := EmailData{
		Username:    username,
		AdminAvatar: fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=640", adminQQ),
		SenderName:  senderName,
		Subject:     subject,
		Body:        template.HTML(htmlContent),
	}

	t, err := template.ParseFS(emailTemplates, "templates/custom_notice.html")
	if err != nil { return err }
	var bodyBuffer bytes.Buffer
	if err := t.Execute(&bodyBuffer, data); err != nil { return err }

	return sendSMTP(toEmail, subject, bodyBuffer.String(), host, port, user, pass, senderName)
}

// getEmailConfig 通过 Provider 获取配置
func getEmailConfig() (string, string, string, string, string, string) {
	if ConfigProvider == nil {
		return "", "", "", "", "", ""
	}
	host := ConfigProvider("SMTP_HOST")
	port := ConfigProvider("SMTP_PORT")
	user := ConfigProvider("SMTP_USER")
	pass := ConfigProvider("SMTP_PASS")
	fURL := ConfigProvider("FRONTEND_URL")
	sName := ConfigProvider("SMTP_SENDER_NAME")

	if fURL == "" { fURL = "http://localhost:5173" }
	if sName == "" { sName = "题酷官方团队" }
	return host, port, user, pass, fURL, sName
}

// sendSMTP 底层发送逻辑 (带 465 SSL 兼容)
func sendSMTP(to, subject, body, host, port, user, pass, sender string) error {
	headers := make(map[string]string)
	encodedSender := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(sender)))
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	headers["From"] = fmt.Sprintf("%s <%s>", encodedSender, user)
	headers["To"] = to
	headers["Subject"] = encodedSubject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	headerStr := ""
	for k, v := range headers {
		headerStr += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg := []byte(headerStr + "\r\n" + body)

	auth := smtp.PlainAuth("", user, pass, host)
	addr := fmt.Sprintf("%s:%s", host, port)

	if port == "465" {
		tlsconfig := &tls.Config{InsecureSkipVerify: true, ServerName: host}
		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil { return err }
		defer conn.Close()
		c, err := smtp.NewClient(conn, host)
		if err != nil { return err }
		defer c.Quit()
		if err = c.Auth(auth); err != nil { return err }
		if err = c.Mail(user); err != nil { return err }
		if err = c.Rcpt(to); err != nil { return err }
		w, err := c.Data()
		if err != nil { return err }
		_, err = w.Write(msg)
		if err != nil { return err }
		return w.Close()
	}
	return smtp.SendMail(addr, auth, user, []string{to}, msg)
}