package user

import (
	"med-platform/internal/common/db"
	"net/http"
	"os"            // 👈 新增：删文件用
	"path/filepath" // 👈 新增：处理扩展名
	"strconv"
	"strings"       // 👈 新增：处理路径字符串
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"     // 👈 新增：生成文件名
	"golang.org/x/crypto/bcrypt" // 👈 新增：密码加密
)

// =======================
// 👮 管理员基础管理 (列表/角色/封禁)
// =======================

// ListUsers 获取用户列表 (带持仓概览)
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")

	var users []User
	var total int64
	offset := (page - 1) * pageSize

	query := db.DB.Model(&User{})
	if keyword != "" {
		query = query.Where("username LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)
	
	if err := query.
		Order("id asc").
		Limit(pageSize).Offset(offset).
		Preload("UserProducts", "expire_at > ?", time.Now()). 
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users, "total": total})
}

// UpdateRole 修改用户角色
func (h *Handler) UpdateRole(c *gin.Context) {
	var req struct {
		UserID  uint   `json:"user_id" binding:"required"`
		NewRole string `json:"new_role" binding:"required,oneof=admin agent user"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}

	currentUserID := c.MustGet("userID").(uint)
	if currentUserID == req.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能修改自己的角色"}); return
	}

	db.DB.Model(&User{}).Where("id = ?", req.UserID).Update("role", req.NewRole)
	c.JSON(http.StatusOK, gin.H{"message": "角色已更新"})
}

// BanUser 封禁用户
func (h *Handler) BanUser(c *gin.Context) {
	var req struct {
		UserID   uint `json:"user_id" binding:"required"`
		Duration int  `json:"duration" binding:"required"` 
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}

	currentUserID := c.MustGet("userID").(uint)
	if currentUserID == req.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能封禁自己"}); return
	}
	
	updates := map[string]interface{}{"status": 2}
	if req.Duration == -1 {
		updates["ban_until"] = time.Now().AddDate(100, 0, 0)
	} else {
		updates["ban_until"] = time.Now().Add(time.Duration(req.Duration) * time.Hour)
	}

	db.DB.Model(&User{}).Where("id = ?", req.UserID).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"message": "用户已封禁"})
}

// UnbanUser 解封用户
func (h *Handler) UnbanUser(c *gin.Context) {
	var req struct { UserID uint `json:"user_id" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}

	db.DB.Model(&User{}).Where("id = ?", req.UserID).Updates(map[string]interface{}{
		"status": 1, "ban_until": nil,
	})
	c.JSON(http.StatusOK, gin.H{"message": "用户已解封"})
}

// ==========================================
// 🔥🔥🔥 管理员上帝视角操作 (增删改查) 🔥🔥🔥
// ==========================================

// AdminGetUserDetail 获取详情
func (h *Handler) AdminGetUserDetail(c *gin.Context) {
	id := c.Param("id")
	var user User
	// Preload 关联数据，方便管理员查看
	if err := db.DB.Preload("UserProducts.Product").First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(200, gin.H{"data": user})
}

// AdminUpdateUserInfo 强制修改资料
func (h *Handler) AdminUpdateUserInfo(c *gin.Context) {
	id := c.Param("id")
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

	if err := db.DB.Model(&User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "更新失败"}); return
	}
	c.JSON(200, gin.H{"message": "用户资料已强制更新"})
}

// AdminResetPassword 强制重置密码
func (h *Handler) AdminResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "密码最少6位"}); return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)

	if err := db.DB.Model(&User{}).Where("id = ?", id).Update("password", string(hashedPwd)).Error; err != nil {
		c.JSON(500, gin.H{"error": "重置失败"}); return
	}
	c.JSON(200, gin.H{"message": "密码已重置"})
}

// AdminUploadAvatar 强制修改头像 (带旧文件清理)
func (h *Handler) AdminUploadAvatar(c *gin.Context) {
	// 1. 获取目标用户ID (从URL参数获取，而非Token)
	targetID := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "请选择图片文件"}); return
	}

	// 2. 🔍 查询旧头像用于清理
	var user User
	if err := db.DB.Select("avatar").First(&user, targetID).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"}); return
	}

	// 3. 🧹 清理垃圾文件
	// 确保只删除本地 uploads 目录下的文件，不误删其他
	if user.Avatar != "" && strings.HasPrefix(user.Avatar, "/uploads/") {
		oldFilePath := "." + user.Avatar
		_ = os.Remove(oldFilePath) // 忽略错误，继续上传
	}

	// 4. 保存新头像
	ext := filepath.Ext(file.Filename)
	fileName := uuid.New().String() + ext
	savePath := "./uploads/" + fileName
	
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": "保存图片失败"}); return
	}

	// 5. 更新数据库
	accessUrl := "/uploads/" + fileName
	db.DB.Model(&User{}).Where("id = ?", targetID).Update("avatar", accessUrl)

	c.JSON(200, gin.H{"message": "头像已强制修改", "url": accessUrl})
}