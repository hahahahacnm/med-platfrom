package feedback

import (
	"encoding/json"
	"med-platform/internal/common/db"
	"med-platform/internal/common/uploader" // 🔥 引入 uploader
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// =======================
// 👤 用户端接口
// =======================

// Create 提交平台反馈
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Type    string   `json:"type" binding:"required"`
		Content string   `json:"content" binding:"required"`
		Images  []string `json:"images"` // 前端先上传图片拿到URL，再传这个数组
		Contact string   `json:"contact"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.MustGet("userID").(uint)

	// 🔥🔥🔥 核心修复：调用通用工具，将图片固化到 "feedback" 目录 🔥🔥🔥
	// 这会自动将 /uploads/temp/xxx.jpg 移动到 /uploads/feedback/xxx.jpg
	finalImages := uploader.ConfirmImages(req.Images, "feedback")

	// 将字符串数组转为 JSON存储
	var imgJSON datatypes.JSON
	if len(finalImages) > 0 {
		bytes, _ := json.Marshal(finalImages)
		imgJSON = datatypes.JSON(bytes)
	}

	fb := PlatformFeedback{
		UserID:  uid,
		Type:    req.Type,
		Content: req.Content,
		Images:  imgJSON,
		Contact: req.Contact,
		Status:  0, // 默认待处理
	}

	if err := db.DB.Create(&fb).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "反馈提交成功，我们会尽快处理！"})
}

// GetMyList 用户查询自己的反馈进度
func (h *Handler) GetMyList(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	var list []PlatformFeedback
	var total int64

	db.DB.Model(&PlatformFeedback{}).Where("user_id = ?", uid).Count(&total)

	// 按时间倒序，最新的在前面
	if err := db.DB.Where("user_id = ?", uid).
		Order("created_at desc").
		Limit(pageSize).Offset(offset).
		Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page})
}

// =======================
// 👮 管理员接口
// =======================

// AdminList 获取所有反馈
func (h *Handler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	statusStr := c.Query("status") // 筛选状态
	typeStr := c.Query("type")     // 筛选类型
	offset := (page - 1) * pageSize

	query := db.DB.Model(&PlatformFeedback{}).Preload("User")

	if statusStr != "" {
		query = query.Where("status = ?", statusStr)
	}
	if typeStr != "" {
		query = query.Where("type = ?", typeStr)
	}

	var total int64
	query.Count(&total)

	var list []PlatformFeedback
	if err := query.Order("created_at desc").Limit(pageSize).Offset(offset).Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

// AdminReply 管理员回复/处理
func (h *Handler) AdminReply(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status     int    `json:"status"` // 1:处理中 2:已解决 3:驳回
		AdminReply string `json:"admin_reply"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"status":      req.Status,
		"admin_reply": req.AdminReply,
	}

	if err := db.DB.Model(&PlatformFeedback{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "处理成功"})
}