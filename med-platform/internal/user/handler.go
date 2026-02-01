package user

import (
	"med-platform/internal/common/db"
	"med-platform/internal/common/jwt"
	"med-platform/internal/common/uploader" 
	"net/http"
	"os"     
	"regexp"
	"strings" 

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// =======================
// 🚪 基础认证 (Auth)
// =======================

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"` 
	Email    string `json:"email"`
	InvitationCode string `json:"invitation_code"` 
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 注册
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var count int64
	db.DB.Model(&User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	// 校验邀请码
	var agentID uint
	if req.InvitationCode != "" {
		var agent User
		if err := db.DB.Where("invitation_code = ? AND role = 'agent'", req.InvitationCode).First(&agent).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请码"})
			return
		}
		agentID = agent.ID
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	finalNickname := req.Nickname
	if finalNickname == "" {
		finalNickname = req.Username
	}

	user := User{
		Username: req.Username,
		Password: string(hashedPwd),
		Nickname: finalNickname,
		Email:    req.Email,
		Role:     "user", 
		Status:   1,
		InvitedBy: agentID, 
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}

	if user.Status == 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "账号已被封禁"})
		return
	}

	token, _ := jwt.GenerateToken(user.ID, user.Username)

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"role":     user.Role,
		"avatar":   user.Avatar,
		"invitation_code": user.InvitationCode, 
	})
}

// =======================
// 👤 个人中心 (Profile)
// =======================

// GetProfile 获取详细资料
func (h *Handler) GetProfile(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	var user User
	// Preload 关联数据
	if err := db.DB.Preload("UserProducts.Product").First(&user, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(200, gin.H{"data": user})
}

// UpdateProfile 更新资料 (支持代理折扣、收款码)
func (h *Handler) UpdateProfile(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	
	// 先查出当前用户角色
	var currentUser User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	var req struct {
		Nickname string `json:"nickname"`
		School   string `json:"school"`
		Major    string `json:"major"`
		Grade    string `json:"grade"`
		QQ       string `json:"qq"`
		WeChat   string `json:"wechat"`
		Gender   int    `json:"gender"`
		Email    string `json:"email"`
		
		// 代理专属字段
		AgentDiscountRate *int `json:"agent_discount_rate"`
		// 🔥🔥🔥 新增：允许通过资料更新保存收款码
		PaymentImage      string `json:"payment_image"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if req.Nickname != "" { updates["nickname"] = req.Nickname }
	if req.School != "" { updates["school"] = req.School }
	if req.Major != "" { updates["major"] = req.Major }
	if req.Grade != "" { updates["grade"] = req.Grade }
	if req.QQ != "" { updates["qq"] = req.QQ }
	if req.WeChat != "" { updates["wechat"] = req.WeChat }
	if req.Gender != 0 { updates["gender"] = req.Gender }
	if req.PaymentImage != "" { updates["payment_image"] = req.PaymentImage } // 🔥 保存收款码
	
	if req.Email != "" {
		pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		if matched, _ := regexp.MatchString(pattern, req.Email); !matched {
			c.JSON(400, gin.H{"error": "邮箱格式不正确"})
			return
		}
		updates["email"] = req.Email
	}

	// 代理设置让利比例逻辑
	if currentUser.Role == "agent" && req.AgentDiscountRate != nil {
		rate := *req.AgentDiscountRate
		if rate < 0 || rate > 20 {
			c.JSON(400, gin.H{"error": "让利比例必须在 0% 到 20% 之间"})
			return
		}
		updates["agent_discount_rate"] = rate
	}

	if err := db.DB.Model(&User{}).Where("id = ?", uid).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(200, gin.H{"message": "资料已更新"})
}

// UploadAvatar 用户上传头像
func (h *Handler) UploadAvatar(c *gin.Context) {
	uid := c.MustGet("userID").(uint)

	accessUrl, err := uploader.SaveImage(c, "file", "avatars", uploader.MaxAvatarSize)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := db.DB.Select("avatar").First(&user, uid).Error; err != nil {
		c.JSON(500, gin.H{"error": "查询用户失败"})
		return
	}

	if user.Avatar != "" && strings.HasPrefix(user.Avatar, "/uploads/") {
		_ = os.Remove("." + user.Avatar)
	}

	db.DB.Model(&User{}).Where("id = ?", uid).Update("avatar", accessUrl)

	c.JSON(200, gin.H{"message": "上传成功", "url": accessUrl})
}

// UploadPaymentCode 专用：上传收款码 (不修改用户头像)
func (h *Handler) UploadPaymentCode(c *gin.Context) {
	// 1. 调用通用上传工具，存入 "payments" 文件夹
	accessUrl, err := uploader.SaveImage(c, "file", "payments", 5*1024*1024) 
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 2. 只返回 URL，不更新数据库 (由前端 UpdateProfile 更新，或 ApplyWithdraw 携带)
	c.JSON(200, gin.H{
		"message": "上传成功", 
		"url": accessUrl,
	})
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user User
	db.DB.First(&user, uid)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(400, gin.H{"error": "旧密码错误"})
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	db.DB.Model(&user).Update("password", string(hashedPwd))

	c.JSON(200, gin.H{"message": "密码修改成功"})
}