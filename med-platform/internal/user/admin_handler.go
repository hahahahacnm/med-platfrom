package user

import (
	"fmt"
	"med-platform/internal/common/db"
	"med-platform/internal/common/uploader"
	"med-platform/internal/payment" // 🔥 引用 payment 包
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// =======================
// 🟢 常量定义
// =======================
const (
	RoleAdmin = "admin"
	RoleAgent = "agent"
	RoleUser  = "user"

	WithdrawStatusPending  = "PENDING"
	WithdrawStatusApproved = "APPROVED"
	WithdrawStatusRejected = "REJECTED"

	SalesStatusUnwithdrawn = 0 
	SalesStatusFrozen      = 1 
	SalesStatusSettled     = 2 
	SalesStatusRefunded    = 3 

	UserStatusActive = 1
	UserStatusBanned = 2
)

// GetDashboardStats 获取控制台统计数据
func (h *Handler) GetDashboardStats(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	role := c.MustGet("role").(string)

	var responseData gin.H

	if role == RoleAgent {
		responseData = h.getAgentStats(userID)
	} else if role == RoleAdmin {
		responseData = h.getAdminStats()
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问控制台"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": responseData})
}

func (h *Handler) getAgentStats(agentID uint) gin.H {
	var agent User
	db.DB.First(&agent, agentID)

	var availableBalance float64
	db.DB.Model(&payment.CommissionLog{}).
		Where("agent_id = ? AND withdraw_status = ?", agentID, SalesStatusUnwithdrawn).
		Select("COALESCE(SUM(profit), 0)").
		Scan(&availableBalance)

	var totalProfit float64
	db.DB.Model(&payment.CommissionLog{}).
		Where("agent_id = ?", agentID).
		Select("COALESCE(SUM(profit), 0)").
		Scan(&totalProfit)

	var inviteCount int64
	db.DB.Model(&User{}).Where("invited_by = ?", agentID).Count(&inviteCount)

	var recentSales []payment.CommissionLog
	db.DB.Where("agent_id = ?", agentID).Order("id desc").Limit(10).Find(&recentSales)

	var pendingCount int64
	// 🔥 修正：使用 payment.WithdrawRequest
	db.DB.Model(&payment.WithdrawRequest{}).
		Where("agent_id = ? AND status = ?", agentID, WithdrawStatusPending).
		Count(&pendingCount)

	return gin.H{
		"role":                 RoleAgent,
		"invitation_code":      agent.InvitationCode,
		"payment_image":        agent.PaymentImage,
		"available_balance":    availableBalance,
		"total_profit":         totalProfit,
		"invite_count":         inviteCount,
		"recent_sales":         recentSales,
		"has_pending_withdraw": pendingCount > 0,
	}
}

func (h *Handler) getAdminStats() gin.H {
	var totalUsers int64
	db.DB.Model(&User{}).Count(&totalUsers)

	var totalRevenue float64
	db.DB.Model(&payment.Order{}).
		Where("status = 'PAID'").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRevenue)

	var pendingCount int64
	// 🔥 修正：使用 payment.WithdrawRequest
	db.DB.Model(&payment.WithdrawRequest{}).Where("status = ?", WithdrawStatusPending).Count(&pendingCount)

	var withdrawList []payment.WithdrawRequest
	db.DB.Order("created_at desc").Limit(50).Find(&withdrawList)

	var recentOrders []payment.Order
	db.DB.Order("created_at desc").Limit(10).Find(&recentOrders)

	return gin.H{
		"role":          RoleAdmin,
		"total_users":   totalUsers,
		"total_revenue": totalRevenue,
		"pending_count": pendingCount,
		"withdraw_list": withdrawList,
		"recent_orders": recentOrders,
	}
}

// ApplyWithdraw 代理申请提现
func (h *Handler) ApplyWithdraw(c *gin.Context) {
	agentID := c.MustGet("userID").(uint)

	var count int64
	// 🔥 修正
	db.DB.Model(&payment.WithdrawRequest{}).
		Where("agent_id = ? AND status = ?", agentID, WithdrawStatusPending).
		Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "您有一笔提现正在审核中"})
		return
	}

	var req struct {
		PaymentImage string `json:"payment_image"`
	}
	_ = c.ShouldBindJSON(&req)

	var agent User
	if err := db.DB.First(&agent, agentID).Error; err != nil {
		c.JSON(500, gin.H{"error": "账户异常"})
		return
	}

	finalImage := agent.PaymentImage
	if req.PaymentImage != "" {
		finalImage = req.PaymentImage
		db.DB.Model(&User{}).Where("id = ?", agentID).Update("payment_image", finalImage)
	}

	if finalImage == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传收款码"})
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var records []payment.CommissionLog
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("agent_id = ? AND withdraw_status = ?", agentID, SalesStatusUnwithdrawn).
			Find(&records).Error; err != nil {
			return err
		}

		var amount float64
		for _, r := range records {
			amount += r.Profit
		}

		if amount < 1.0 {
			return fmt.Errorf("可提现余额不足 1 元")
		}

		// 🔥 修正
		withdraw := payment.WithdrawRequest{
			AgentID:      agentID,
			AgentName:    agent.Nickname,
			Amount:       amount,
			PaymentImage: finalImage,
			Status:       WithdrawStatusPending,
		}
		if err := tx.Create(&withdraw).Error; err != nil {
			return err
		}

		if err := tx.Model(&payment.CommissionLog{}).
			Where("agent_id = ? AND withdraw_status = ?", agentID, SalesStatusUnwithdrawn).
			Update("withdraw_status", SalesStatusFrozen).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "申请已提交"})
}

// HandleWithdraw 管理员审核提现
func (h *Handler) HandleWithdraw(c *gin.Context) {
	var req struct {
		RequestID uint   `json:"request_id" binding:"required"`
		Action    string `json:"action" binding:"required,oneof=APPROVED REJECTED"`
		Comment   string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// 🔥 修正
		var withdraw payment.WithdrawRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&withdraw, req.RequestID).Error; err != nil {
			return fmt.Errorf("申请单不存在")
		}

		if withdraw.Status != WithdrawStatusPending {
			return fmt.Errorf("该申请单已被处理")
		}

		withdraw.Status = req.Action
		withdraw.AdminComment = req.Comment
		if err := tx.Save(&withdraw).Error; err != nil {
			return err
		}

		var newStatus int
		if req.Action == WithdrawStatusApproved {
			newStatus = SalesStatusSettled
		} else {
			newStatus = SalesStatusRefunded 
		}

		result := tx.Model(&payment.CommissionLog{}).
			Where("agent_id = ? AND withdraw_status = ?", withdraw.AgentID, SalesStatusFrozen).
			Update("withdraw_status", newStatus)

		if result.Error != nil {
			return result.Error
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
	// 🔥 修正
	if err := db.DB.Delete(&payment.WithdrawRequest{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(200, gin.H{"message": "记录已移除"})
}

func (h *Handler) ClearHandledWithdraws(c *gin.Context) {
	// 🔥 修正
	result := db.DB.Where("status IN ?", []string{WithdrawStatusApproved, WithdrawStatusRejected}).
		Delete(&payment.WithdrawRequest{})

	if result.Error != nil {
		c.JSON(500, gin.H{"error": "清理失败"})
		return
	}
	c.JSON(200, gin.H{"message": fmt.Sprintf("清理成功，共移除 %d 条记录", result.RowsAffected)})
}

// ListUsers (保持原样)
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	role := c.Query("role")

	var users []User
	var total int64
	query := db.DB.Model(&User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ?", keyword+"%", keyword+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	query.Count(&total)
	query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Preload("UserProducts").Find(&users)
	c.JSON(200, gin.H{"data": users, "total": total})
}

// UpdateRole (保持原样)
func (h *Handler) UpdateRole(c *gin.Context) {
	var req struct {
		UserID  uint   `json:"user_id" binding:"required"`
		NewRole string `json:"new_role" binding:"required,oneof=admin agent user"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	currentUserID := c.MustGet("userID").(uint)
	if currentUserID == req.UserID {
		c.JSON(403, gin.H{"error": "不能修改自己的角色"})
		return
	}
	var targetUser User
	db.DB.First(&targetUser, req.UserID)
	updates := map[string]interface{}{"role": req.NewRole}
	
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if req.NewRole == RoleAgent && targetUser.Role != RoleAgent {
			if targetUser.InvitationCode == nil || *targetUser.InvitationCode == "" {
				raw := strings.ReplaceAll(uuid.New().String(), "-", "")
				code := fmt.Sprintf("AG%d%s", req.UserID, strings.ToUpper(raw)[:4])
				updates["invitation_code"] = code
			}
		}
		if targetUser.Role == RoleAgent && req.NewRole == RoleUser {
			updates["invitation_code"] = nil
			updates["agent_discount_rate"] = 0
			tx.Model(&User{}).Where("invited_by = ?", targetUser.ID).Update("invited_by", 0)
		}
		return tx.Model(&User{}).Where("id = ?", req.UserID).Updates(updates).Error
	})

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "角色已更新"})
}

// BanUser, UnbanUser... 等其他函数保持不变，请确保从旧文件复制过来或保持原样
func (h *Handler) BanUser(c *gin.Context) {
	var req struct {
		UserID   uint `json:"user_id"`
		Duration int  `json:"duration"`
	}
	c.ShouldBindJSON(&req)
	updates := map[string]interface{}{"status": 2}
	if req.Duration == -1 {
		updates["ban_until"] = time.Now().AddDate(100, 0, 0)
	} else {
		updates["ban_until"] = time.Now().Add(time.Duration(req.Duration) * time.Hour)
	}
	db.DB.Model(&User{}).Where("id = ?", req.UserID).Updates(updates)
	c.JSON(200, gin.H{"message": "已封禁"})
}

func (h *Handler) UnbanUser(c *gin.Context) {
	var req struct{ UserID uint `json:"user_id"` }
	c.ShouldBindJSON(&req)
	db.DB.Model(&User{}).Where("id = ?", req.UserID).Updates(map[string]interface{}{"status": 1, "ban_until": nil})
	c.JSON(200, gin.H{"message": "已解封"})
}

func (h *Handler) AdminGetUserDetail(c *gin.Context) {
	id := c.Param("id")
	var user User
	db.DB.Preload("UserProducts.Product").First(&user, id)
	c.JSON(200, gin.H{"data": user})
}

func (h *Handler) AdminUpdateUserInfo(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	db.DB.Model(&User{}).Where("id = ?", id).Updates(req)
	c.JSON(200, gin.H{"message": "更新成功"})
}

func (h *Handler) AdminResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req struct{ NewPassword string `json:"new_password"` }
	c.ShouldBindJSON(&req)
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	db.DB.Model(&User{}).Where("id = ?", id).Update("password", string(hashed))
	c.JSON(200, gin.H{"message": "重置成功"})
}

func (h *Handler) AdminUploadAvatar(c *gin.Context) {
	id := c.Param("id")
	url, _ := uploader.SaveImage(c, "file", "avatars", 5*1024*1024)
	db.DB.Model(&User{}).Where("id = ?", id).Update("avatar", url)
	c.JSON(200, gin.H{"message": "上传成功", "url": url})
}