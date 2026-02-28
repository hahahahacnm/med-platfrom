package user

import (
	"net/http"
	"regexp"
	"time"

	"med-platform/internal/common/captcha"
	"med-platform/internal/common/db"
	"med-platform/internal/common/jwt"
	"med-platform/internal/common/service" // 🔥 引入邮件服务
	"med-platform/internal/common/uploader"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type RegisterRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	Nickname       string `json:"nickname" binding:"required"`
	Email          string `json:"email" binding:"required"`
	InvitationCode string `json:"invitation_code"` 
	CaptchaId      string `json:"captcha_id"`
	CaptchaValue   string `json:"captcha_value"`
}

// =======================
// 🚪 基础认证 (注册、魔法链接验证、重发)
// =======================

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数填写不完整"})
		return
	}

	if !captcha.Verify(req.CaptchaId, req.CaptchaValue) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已失效"})
		return
	}

	if matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]{3,19}$`, req.Username); !matched {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名需以字母开头，仅含字母数字下划线，4-20位"})
		return
	}

	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, req.Email); !matched {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
		return
	}

	var agentID uint
	if req.InvitationCode != "" {
		var agent User
		if err := db.DB.Where("invitation_code = ? AND role = 'agent'", req.InvitationCode).First(&agent).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请码"})
			return
		}
		agentID = agent.ID
	}

	// 核心逻辑：检查是否已存在记录，及“懒惰覆盖”机制
	var existingUser User
	err := db.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error

	if err == nil {
		if existingUser.Status == 1 || existingUser.Status == 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该用户名或邮箱已被占用"})
			return
		}

		var lastToken VerificationToken
		if errToken := db.DB.Where("user_id = ? AND type = 'register'", existingUser.ID).Order("created_at desc").First(&lastToken).Error; errToken == nil {
			if time.Since(lastToken.CreatedAt) < 60*time.Second {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "邮件发送太频繁，请 1 分钟后再试"})
				return
			}
		}

		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		
		db.DB.Transaction(func(tx *gorm.DB) error {
			tx.Model(&existingUser).Updates(map[string]interface{}{
				"username":   req.Username,
				"password":   string(hashedPwd),
				"nickname":   req.Nickname,
				"email":      req.Email,
				"invited_by": agentID,
			})
			
			tx.Where("user_id = ? AND type = 'register'", existingUser.ID).Delete(&VerificationToken{})
			tokenStr := uuid.New().String()
			tx.Create(&VerificationToken{
				UserID:    existingUser.ID,
				Email:     req.Email,
				Token:     tokenStr,
				Type:      "register",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			})
			// 🔥 修复处 1：增加 req.Username
			go service.SendVerificationEmail(req.Email, req.Username, tokenStr, "register")
			return nil
		})

		c.JSON(http.StatusOK, gin.H{"message": "我们已向您的邮箱发送了验证链接，请前往点击激活即可完成注册（24小时内有效）"})
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	newUser := User{
		Username:  req.Username,
		Password:  string(hashedPwd),
		Nickname:  req.Nickname,
		Email:     req.Email,
		Role:      "user",
		Status:    0,
		InvitedBy: agentID,
	}

	db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}
		tokenStr := uuid.New().String()
		tx.Create(&VerificationToken{
			UserID:    newUser.ID,
			Email:     req.Email,
			Token:     tokenStr,
			Type:      "register",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
		// 🔥 修复处 2：增加 req.Username
		go service.SendVerificationEmail(req.Email, req.Username, tokenStr, "register")
		return nil
	})

	c.JSON(http.StatusOK, gin.H{"message": "我们已向您的邮箱发送了验证链接，请前往点击激活即可完成注册（24小时内有效）"})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	tokenType := c.Query("type")

	if token == "" || tokenType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的验证链接，缺少参数"})
		return
	}

	var vt VerificationToken
	if err := db.DB.Where("token = ? AND type = ?", token, tokenType).First(&vt).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该验证链接不存在或已被使用"})
		return
	}

	if time.Now().After(vt.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该验证链接已过期，请重新获取"})
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if tokenType == "register" {
			if err := tx.Model(&User{}).Where("id = ?", vt.UserID).Update("status", 1).Error; err != nil {
				return err
			}
		} else if tokenType == "change_email" {
			if err := tx.Model(&User{}).Where("id = ?", vt.UserID).Update("email", vt.Email).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&vt).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统内部错误，处理失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "验证成功"})
}

func (h *Handler) ResendEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱不能为空"})
		return
	}

	var u User
	if err := db.DB.Where("email = ? AND status = 0", req.Email).First(&u).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到待激活的账号或账号已激活"})
		return
	}

	var lastToken VerificationToken
	if err := db.DB.Where("user_id = ? AND type = 'register'", u.ID).Order("created_at desc").First(&lastToken).Error; err == nil {
		if time.Since(lastToken.CreatedAt) < 60*time.Second {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "邮件发送太频繁，请 1 分钟后再试"})
			return
		}
	}

	db.DB.Where("user_id = ? AND type = 'register'", u.ID).Delete(&VerificationToken{})
	tokenStr := uuid.New().String()
	db.DB.Create(&VerificationToken{
		UserID:    u.ID,
		Email:     u.Email,
		Token:     tokenStr,
		Type:      "register",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	// 🔥 修复处 3：自动获取用户昵称用于发信
	name := u.Nickname
	if name == "" { name = u.Username }
	go service.SendVerificationEmail(u.Email, name, tokenStr, "register")
	
	c.JSON(http.StatusOK, gin.H{"message": "激活邮件已重新发送，请注意查收"})
}

// =======================
// 🔑 登录逻辑
// =======================
type LoginRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	CaptchaId    string `json:"captcha_id"`
	CaptchaValue string `json:"captcha_value"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !captcha.Verify(req.CaptchaId, req.CaptchaValue) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已失效"})
		return
	}

	var user User
	if err := db.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}

	if user.Status == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "账号尚未激活，请前往邮箱点击验证链接",
			"email": user.Email,
		})
		return
	}
	if user.Status == 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "账号已被封禁"})
		return
	}

	token, _ := jwt.GenerateToken(user.ID, user.Username)

	c.JSON(http.StatusOK, gin.H{
		"token":           token,
		"id":              user.ID,
		"username":        user.Username,
		"nickname":        user.Nickname,
		"role":            user.Role,
		"avatar":          user.Avatar,
		"school":          user.School,
		"major":           user.Major,
		"grade":           user.Grade,
		"invitation_code": user.InvitationCode,
	})
}


// =======================
// 👤 个人中心 (Profile)
// =======================
func (h *Handler) GetProfile(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	var user User
	if err := db.DB.Preload("UserProducts.Product").First(&user, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(200, gin.H{"data": user})
}

// BindNewEmail 换绑新邮箱专用接口
func (h *Handler) BindNewEmail(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式错误"})
		return
	}

	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, req.Email); !matched {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
		return
	}

	var count int64
	db.DB.Model(&User{}).Where("email = ? AND status != 0 AND id != ?", req.Email, uid).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该邮箱已被其他账号绑定"})
		return
	}

	var lastToken VerificationToken
	if err := db.DB.Where("user_id = ? AND type = 'change_email'", uid).Order("created_at desc").First(&lastToken).Error; err == nil {
		if time.Since(lastToken.CreatedAt) < 60*time.Second {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "获取验证链接太频繁，请 1 分钟后再试"})
			return
		}
	}

	db.DB.Where("user_id = ? AND type = 'change_email'", uid).Delete(&VerificationToken{})
	tokenStr := uuid.New().String()
	db.DB.Create(&VerificationToken{
		UserID:    uid,
		Email:     req.Email,
		Token:     tokenStr,
		Type:      "change_email",
		ExpiresAt: time.Now().Add(30 * time.Minute),
	})

	// 🔥 修复处 4：极速查询当前操作者的用户名
	var currentUser User
	db.DB.Select("username", "nickname").First(&currentUser, uid)
	name := currentUser.Nickname
	if name == "" { name = currentUser.Username }

	go service.SendVerificationEmail(req.Email, name, tokenStr, "change_email")
	
	c.JSON(http.StatusOK, gin.H{"message": "确认链接已发送至新邮箱，请在 30 分钟内点击确认"})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	uid := c.MustGet("userID").(uint)

	var currentUser User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	var req struct {
		Nickname          string `json:"nickname"`
		School            string `json:"school"`
		Major             string `json:"major"`
		Grade             string `json:"grade"`
		QQ                string `json:"qq"`
		WeChat            string `json:"wechat"`
		Gender            int    `json:"gender"`
		AgentDiscountRate *int   `json:"agent_discount_rate"`
		PaymentImage      string `json:"payment_image"`
	}
	_ = c.ShouldBindJSON(&req)

	updates := map[string]interface{}{}
	if req.Nickname != "" { updates["nickname"] = req.Nickname }
	if req.School != "" { updates["school"] = req.School }
	if req.Major != "" { updates["major"] = req.Major }
	if req.Grade != "" { updates["grade"] = req.Grade }
	if req.QQ != "" { updates["qq"] = req.QQ }
	if req.WeChat != "" { updates["wechat"] = req.WeChat }
	if req.Gender != 0 { updates["gender"] = req.Gender }
	if req.PaymentImage != "" { updates["payment_image"] = req.PaymentImage }

	if currentUser.Role == "agent" && req.AgentDiscountRate != nil {
		rate := *req.AgentDiscountRate
		if rate >= 0 && rate <= 20 {
			updates["agent_discount_rate"] = rate
		}
	}

	if err := db.DB.Model(&User{}).Where("id = ?", uid).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(200, gin.H{"message": "资料已更新"})
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	accessUrl, err := uploader.SaveImage(c, "file", "avatars", uploader.MaxAvatarSize)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&User{}).Where("id = ?", uid).Update("avatar", accessUrl)
	c.JSON(200, gin.H{"message": "上传成功", "url": accessUrl})
}

func (h *Handler) UploadPaymentCode(c *gin.Context) {
	accessUrl, err := uploader.SaveImage(c, "file", "payments", 5*1024*1024)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "上传成功", "url": accessUrl})
}

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

func (h *Handler) GetCaptcha(c *gin.Context) {
	key, thumb, master, err := captcha.Generate()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "验证码生成失败"})
		return
	}
	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"captcha_id": key,
			"block":      thumb,
			"background": master,
		},
	})
}