package user

import (
	"med-platform/internal/common/db"
	"med-platform/internal/common/jwt"
	"net/http"
	"path/filepath"
	"regexp"
	"os"      // 用于删除文件
	"strings" // 用于处理字符串路径

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	Nickname string `json:"nickname"` // 可选昵称
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 注册
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}

	var count int64
	db.DB.Model(&User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"}); return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	// 逻辑优化：如果没填昵称，默认等于用户名
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
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"}); return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}

	var user User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"}); return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"}); return
	}

	if user.Status == 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "账号已被封禁"}); return
	}

	token, _ := jwt.GenerateToken(user.ID, user.Username)

	// 🔥🔥🔥 关键修改：必须返回 ID，否则前端无法判定身份 🔥🔥🔥
	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"id":       user.ID,       // 👈 加上这一行！一切问题的解药！
		"username": user.Username,
		"nickname": user.Nickname,
		"role":     user.Role,
		"avatar":   user.Avatar,
	})
}

// =======================
// 👤 个人中心 (Profile)
// =======================

// GetProfile 获取详细资料
func (h *Handler) GetProfile(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	var user User
	if err := db.DB.Preload("UserProducts.Product").First(&user, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"}); return
	}
	c.JSON(200, gin.H{"data": user})
}

// UpdateProfile 更新资料
func (h *Handler) UpdateProfile(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	var req struct {
		Nickname string `json:"nickname"`
		School   string `json:"school"`
		Major    string `json:"major"`
		Grade    string `json:"grade"`
		QQ       string `json:"qq"`
		WeChat   string `json:"wechat"`
		Gender   int    `json:"gender"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()}); return
	}

	if req.Email != "" {
		pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		if matched, _ := regexp.MatchString(pattern, req.Email); !matched {
			c.JSON(400, gin.H{"error": "邮箱格式不正确"}); return
		}
	}

	updates := map[string]interface{}{
		"nickname": req.Nickname,
		"school":   req.School,
		"major":    req.Major,
		"grade":    req.Grade,
		"qq":       req.QQ,
		"wechat":   req.WeChat,
		"gender":   req.Gender,
		"email":    req.Email,
	}

	if err := db.DB.Model(&User{}).Where("id = ?", uid).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "更新失败"}); return
	}
	c.JSON(200, gin.H{"message": "资料已更新"})
}

// UploadAvatar 上传头像 (自动清理旧头像)
func (h *Handler) UploadAvatar(c *gin.Context) {
	uid := c.MustGet("userID").(uint)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "请选择图片文件"})
		return
	}

	// 🔍 查询旧头像
	var user User
	if err := db.DB.Select("avatar").First(&user, uid).Error; err != nil {
		c.JSON(500, gin.H{"error": "用户数据查询失败"})
		return
	}

	// 🧹 清理旧头像逻辑
	if user.Avatar != "" && strings.HasPrefix(user.Avatar, "/uploads/") {
		oldFilePath := "." + user.Avatar // 拼接相对路径
		_ = os.Remove(oldFilePath)       // 忽略错误，继续上传
	}

	// 保存新头像
	ext := filepath.Ext(file.Filename)
	fileName := uuid.New().String() + ext
	savePath := "./uploads/" + fileName
	
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": "保存图片失败"})
		return
	}

	// 更新数据库
	accessUrl := "/uploads/" + fileName
	db.DB.Model(&User{}).Where("id = ?", uid).Update("avatar", accessUrl)

	c.JSON(200, gin.H{"message": "上传成功", "url": accessUrl})
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()}); return
	}

	var user User
	db.DB.First(&user, uid)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(400, gin.H{"error": "旧密码错误"}); return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	db.DB.Model(&user).Update("password", string(hashedPwd))
	
	c.JSON(200, gin.H{"message": "密码修改成功"})
}