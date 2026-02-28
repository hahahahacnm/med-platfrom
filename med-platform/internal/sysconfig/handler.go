package sysconfig

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"time"

	"med-platform/internal/common/db"
	"med-platform/internal/common/service" // 🔥 允许导入 service

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	// 🔥 核心修复：在这里注入配置获取函数，打破编译时的循环依赖
	service.ConfigProvider = GetConfig 
	return &Handler{}
}

// ================== 原有基础配置接口 ==================

func (h *Handler) ListConfigs(c *gin.Context) {
	var configs []SysConfig
	db.DB.Find(&configs)
	c.JSON(http.StatusOK, gin.H{"data": configs})
}

func (h *Handler) SaveConfig(c *gin.Context) {
	var req struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	if req.Key == KeyAgentRateDirect || req.Key == KeyAgentRateCard {
		rate, err := strconv.ParseFloat(req.Value, 64)
		if err != nil || rate < 0 || rate > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "分润比例必须是0到1之间的小数"})
			return
		}
	}

	var config SysConfig
	if err := db.DB.Where("key = ?", req.Key).First(&config).Error; err != nil {
		config = SysConfig{Key: req.Key, Value: req.Value, Description: req.Description}
		db.DB.Create(&config)
	} else {
		db.DB.Model(&config).Updates(map[string]interface{}{
			"value":       req.Value,
			"description": req.Description,
		})
	}
	InitConfig() // 刷新内存缓存
	c.JSON(http.StatusOK, gin.H{"message": "配置更新成功"})
}

func (h *Handler) SendTestEmail(c *gin.Context) {
	var req struct {
		TargetEmail string `json:"target_email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供有效的测试邮箱地址"})
		return
	}

	host := GetConfig("SMTP_HOST")
	port := GetConfig("SMTP_PORT")
	user := GetConfig("SMTP_USER")
	pass := GetConfig("SMTP_PASS")
	senderName := GetConfig("SMTP_SENDER_NAME")
	if senderName == "" { senderName = "平台系统测试" }

	if host == "" || user == "" || pass == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "系统 SMTP 尚未配置完成"})
		return
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, pass, host)

	subjectText := "🚀 平台配置中心测试邮件"
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subjectText)))
	encodedSender := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(senderName)))

	headerStr := fmt.Sprintf("To: %s\r\nFrom: %s <%s>\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		req.TargetEmail, encodedSender, user, encodedSubject)

	body := "<h3>这是一封测试邮件。</h3><p>如果您收到此邮件，说明 SMTP 服务已配置成功！</p>"
	msg := []byte(headerStr + body)

	var err error
	if port == "465" {
		tlsconfig := &tls.Config{InsecureSkipVerify: true, ServerName: host}
		conn, errConn := tls.Dial("tcp", addr, tlsconfig)
		if errConn != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "安全连接失败: " + errConn.Error()})
			return
		}
		defer conn.Close()

		client, errClient := smtp.NewClient(conn, host)
		if errClient != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建SMTP客户端失败: " + errClient.Error()})
			return
		}
		defer client.Quit()
		if err = client.Auth(auth); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "账号授权认证失败: " + err.Error()})
			return
		}
		if err = client.Mail(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发件人错误: " + err.Error()})
			return
		}
		if err = client.Rcpt(req.TargetEmail); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "收件人错误: " + err.Error()})
			return
		}
		w, err := client.Data()
		if err != nil { return }
		if _, err = w.Write(msg); err != nil { return }
		err = w.Close()
	} else {
		err = smtp.SendMail(addr, auth, user, []string{req.TargetEmail}, msg)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "测试邮件已发送"})
}

// ================== 🔥 新增：邮件营销/群发后台 ==================

// UserEmailInfo 用户轻量结构，用于查询
type UserEmailInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}
func (UserEmailInfo) TableName() string { return "users" }

// ListEmailUsers 搜索带邮箱的用户
func (h *Handler) ListEmailUsers(c *gin.Context) {
	var users []UserEmailInfo
	q := c.Query("q")
	query := db.DB.Model(&UserEmailInfo{}).Where("email IS NOT NULL AND email != ''")
	if q != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", "%"+q+"%", "%"+q+"%", "%"+q+"%")
	}
	query.Find(&users)
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// SendCustomMail 执行异步群发任务
func (h *Handler) SendCustomMail(c *gin.Context) {
	var req struct {
		TargetType string `json:"target_type" binding:"required"`
		UserIDs    []uint `json:"user_ids"`
		Subject    string `json:"subject" binding:"required"`
		Content    string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var targets []UserEmailInfo
	query := db.DB.Model(&UserEmailInfo{}).Where("email IS NOT NULL AND email != ''")
	if req.TargetType == "specific" {
		if len(req.UserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择收件人"})
			return
		}
		query = query.Where("id IN ?", req.UserIDs)
	}
	query.Find(&targets)

	if len(targets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到有效收件人"})
		return
	}

	// 异步发送任务
	go func() {
		for i, t := range targets {
			name := t.Nickname
			if name == "" { name = t.Username }
			_ = service.SendCustomEmail(t.Email, name, req.Subject, req.Content)
			
			// 频率限制：每 5 封休息 2 秒
			if i > 0 && i%5 == 0 { time.Sleep(2 * time.Second) }
		}
		log.Printf("📬 任务完成：已向 %d 位用户发送邮件", len(targets))
	}()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("发信任务已启动，预计发送 %d 人", len(targets))})
}