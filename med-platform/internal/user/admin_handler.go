package user

import (
	"fmt"
	"med-platform/internal/common/db"
	"med-platform/internal/common/uploader"
	"med-platform/internal/product" 
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// =======================
// 📊 控制台 / 仪表盘 (Dashboard)
// =======================

// GetDashboardStats 获取控制台统计数据
func (h *Handler) GetDashboardStats(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	role := c.MustGet("role").(string)

	stats := gin.H{}

	if role == "agent" {
		// === 👮 代理视角 ===
		var agent User
		db.DB.First(&agent, userID)

		// 1. 计算可提现余额
		var availableBalance float64
		db.DB.Model(&product.SalesRecord{}).
			Where("agent_id = ? AND withdraw_status = 0", userID).
			Select("COALESCE(SUM(agent_profit), 0)").
			Scan(&availableBalance)

		// 2. 计算累计总收益
		var totalProfit float64
		db.DB.Model(&product.SalesRecord{}).
			Where("agent_id = ?", userID).
			Select("COALESCE(SUM(agent_profit), 0)").
			Scan(&totalProfit)

		// 3. 计算累计邀请人数
		var inviteCount int64
		db.DB.Model(&User{}).Where("invited_by = ?", userID).Count(&inviteCount)

		// 4. 获取最近 10 条收益记录
		var recentSales []product.SalesRecord
		db.DB.Where("agent_id = ?", userID).Order("id desc").Limit(10).Find(&recentSales)

		// 5. 检查是否有正在审核的提现申请
		var pendingWithdraw product.WithdrawRequest
		hasPending := false
		if err := db.DB.Where("agent_id = ? AND status = 'PENDING'", userID).First(&pendingWithdraw).Error; err == nil {
			hasPending = true
		}

		stats = gin.H{
			"role":                 "agent",
			"invitation_code":      agent.InvitationCode,
			"payment_image":        agent.PaymentImage, // 🔥 返回已保存的收款码
			"available_balance":    availableBalance,
			"total_profit":         totalProfit,
			"invite_count":         inviteCount,
			"recent_sales":         recentSales,
			"has_pending_withdraw": hasPending,
		}

	} else if role == "admin" {
		// ... (管理员逻辑保持不变) ...
		var totalUsers int64
		db.DB.Model(&User{}).Count(&totalUsers)

		var totalRevenue float64
		db.DB.Model(&product.SalesRecord{}).
			Select("COALESCE(SUM(final_amount), 0)").
			Scan(&totalRevenue)

		var pendingCount int64
		db.DB.Model(&product.WithdrawRequest{}).Where("status = 'PENDING'").Count(&pendingCount)

		var withdrawList []product.WithdrawRequest
		db.DB.Order("created_at desc").Limit(50).Find(&withdrawList)

		type OrderDTO struct {
			OrderNo     string    `json:"order_no"`
			Amount      float64   `json:"amount"`
			Status      string    `json:"status"`
			CreatedAt   time.Time `json:"created_at"`
			Username    string    `json:"username"`
			ProductName string    `json:"product_name"` 
		}
		var recentOrders []OrderDTO
		
		db.DB.Table("orders").
			Select("orders.order_no, orders.amount, orders.status, orders.created_at, users.username, products.name as product_name").
			Joins("left join users on users.id = orders.user_id").
			Joins("left join products on products.id = orders.product_id").
			Order("orders.created_at desc").
			Limit(10).
			Scan(&recentOrders)

		stats = gin.H{
			"role":           "admin",
			"total_users":    totalUsers,
			"total_revenue":  totalRevenue,
			"pending_count":  pendingCount,
			"withdraw_list":  withdrawList,
			"recent_orders":  recentOrders,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// ApplyWithdraw 代理申请提现 (🔥 智能识别收款码 🔥)
func (h *Handler) ApplyWithdraw(c *gin.Context) {
	agentID := c.MustGet("userID").(uint)

	// 1. 检查是否有未完成的申请
	var count int64
	db.DB.Model(&product.WithdrawRequest{}).
		Where("agent_id = ? AND status = 'PENDING'", agentID).
		Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "您有一笔提现正在审核中，请勿重复申请"})
		return
	}

	// payment_image 变为可选参数
	var req struct {
		PaymentImage string `json:"payment_image"` 
	}
	_ = c.ShouldBindJSON(&req) 

	// 2. 确定最终使用的收款码
	var agent User
	db.DB.First(&agent, agentID)

	finalImage := ""
	
	if req.PaymentImage != "" {
		// A. 用户本次提交了新图 -> 使用新图 + 更新到 Profile
		finalImage = req.PaymentImage
		db.DB.Model(&User{}).Where("id = ?", agentID).Update("payment_image", finalImage)
	} else if agent.PaymentImage != "" {
		// B. 用户没传，但 Profile 里有 -> 使用存图
		finalImage = agent.PaymentImage
	} else {
		// C. 都没有 -> 报错
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传收款码或在个人中心设置"})
		return
	}

	// 3. 事务提现流程
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var amount float64
		tx.Model(&product.SalesRecord{}).
			Where("agent_id = ? AND withdraw_status = 0", agentID).
			Select("COALESCE(SUM(agent_profit), 0)").
			Scan(&amount)

		if amount < 1 {
			return fmt.Errorf("可提现余额不足 1 元")
		}

		withdraw := product.WithdrawRequest{
			AgentID:      agentID,
			AgentName:    agent.Nickname,
			Amount:       amount,
			PaymentImage: finalImage, // 使用最终确定的图
			Status:       "PENDING",
		}
		if err := tx.Create(&withdraw).Error; err != nil {
			return err
		}

		if err := tx.Model(&product.SalesRecord{}).
			Where("agent_id = ? AND withdraw_status = 0", agentID).
			Update("withdraw_status", 1).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "提现申请已提交，请等待管理员审核"})
}

// HandleWithdraw 管理员审核提现
func (h *Handler) HandleWithdraw(c *gin.Context) {
	var req struct {
		RequestID uint   `json:"request_id" binding:"required"`
		Action    string `json:"action" binding:"required,oneof=APPROVED REJECTED"`
		Comment   string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var withdraw product.WithdrawRequest
		if err := tx.First(&withdraw, req.RequestID).Error; err != nil {
			return fmt.Errorf("申请单不存在")
		}

		if withdraw.Status != "PENDING" {
			return fmt.Errorf("该申请单已被处理")
		}

		withdraw.Status = req.Action
		withdraw.AdminComment = req.Comment
		if err := tx.Save(&withdraw).Error; err != nil {
			return err
		}

		if req.Action == "APPROVED" {
			tx.Model(&product.SalesRecord{}).
				Where("agent_id = ? AND withdraw_status = 1", withdraw.AgentID).
				Update("withdraw_status", 2)
		} else {
			tx.Model(&product.SalesRecord{}).
				Where("agent_id = ? AND withdraw_status = 1", withdraw.AgentID).
				Update("withdraw_status", 3) // 驳回=状态3
		}

		return nil
	})

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "操作成功"})
}

func (h *Handler) DeleteWithdraw(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Unscoped().Delete(&product.WithdrawRequest{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(200, gin.H{"message": "记录已删除"})
}

func (h *Handler) ClearHandledWithdraws(c *gin.Context) {
	result := db.DB.Unscoped().Where("status IN ?", []string{"APPROVED", "REJECTED"}).Delete(&product.WithdrawRequest{})
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "清理失败"})
		return
	}
	c.JSON(200, gin.H{
		"message": fmt.Sprintf("清理成功，共释放 %d 条记录", result.RowsAffected),
	})
}

// =======================
// 👮 基础管理代码
// =======================

// ListUsers 获取用户列表
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	role := c.Query("role") 

	var users []User
	var total int64
	offset := (page - 1) * pageSize

	query := db.DB.Model(&User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.MustGet("userID").(uint)
	if currentUserID == req.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能修改自己的角色"})
		return
	}

	updates := map[string]interface{}{
		"role": req.NewRole,
	}

	if req.NewRole == "agent" {
		var user User
		if err := db.DB.First(&user, req.UserID).Error; err == nil {
			if user.InvitationCode == "" {
				rawUUID := strings.ReplaceAll(uuid.New().String(), "-", "")
				randomSuffix := strings.ToUpper(rawUUID)[:4]
				code := fmt.Sprintf("AG%d%s", req.UserID, randomSuffix)
				updates["invitation_code"] = code
			}
		}
	}

	if err := db.DB.Model(&User{}).Where("id = ?", req.UserID).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色已更新"})
}

// BanUser 封禁用户
func (h *Handler) BanUser(c *gin.Context) {
	var req struct {
		UserID   uint `json:"user_id" binding:"required"`
		Duration int  `json:"duration" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.MustGet("userID").(uint)
	if currentUserID == req.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能封禁自己"})
		return
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
	var req struct{ UserID uint `json:"user_id" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.DB.Model(&User{}).Where("id = ?", req.UserID).Updates(map[string]interface{}{
		"status": 1, "ban_until": nil,
	})
	c.JSON(http.StatusOK, gin.H{"message": "用户已解封"})
}

// AdminGetUserDetail 获取详情
func (h *Handler) AdminGetUserDetail(c *gin.Context) {
	id := c.Param("id")
	var user User
	// Preload 关联数据
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
		c.JSON(400, gin.H{"error": err.Error()})
		return
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
		c.JSON(500, gin.H{"error": "更新失败"})
		return
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
		c.JSON(400, gin.H{"error": "密码最少6位"})
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)

	if err := db.DB.Model(&User{}).Where("id = ?", id).Update("password", string(hashedPwd)).Error; err != nil {
		c.JSON(500, gin.H{"error": "重置失败"})
		return
	}
	c.JSON(200, gin.H{"message": "密码已重置"})
}

// AdminUploadAvatar 强制修改头像
func (h *Handler) AdminUploadAvatar(c *gin.Context) {
	targetID := c.Param("id")

	accessUrl, err := uploader.SaveImage(c, "file", "avatars", uploader.MaxAvatarSize)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := db.DB.Select("avatar").First(&user, targetID).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	if user.Avatar != "" && strings.HasPrefix(user.Avatar, "/uploads/") {
		_ = os.Remove("." + user.Avatar)
	}

	db.DB.Model(&User{}).Where("id = ?", targetID).Update("avatar", accessUrl)

	c.JSON(200, gin.H{"message": "头像已强制修改", "url": accessUrl})
}