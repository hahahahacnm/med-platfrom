package product

import (
	"fmt"
	"med-platform/internal/common/db"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	repo *Repository
}

func NewHandler() *Handler {
	return &Handler{repo: NewRepository()}
}

// --- 1. 商品管理 (Product SKU) ---

// CreateProduct 创建商品壳子
func (h *Handler) CreateProduct(c *gin.Context) {
	var req struct { Name string `json:"name"`; Description string `json:"description"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }
	p := Product{Name: req.Name, Description: req.Description}
	if err := db.DB.Create(&p).Error; err != nil { c.JSON(500, gin.H{"error": "创建失败"}); return }
	c.JSON(200, gin.H{"message": "商品创建成功", "data": p})
}

// ListProducts 查看所有商品
func (h *Handler) ListProducts(c *gin.Context) {
	var list []Product
	db.DB.Find(&list)
	c.JSON(200, gin.H{"data": list})
}

// DeleteProduct 删除商品 (级联删除所有绑定和用户持有)
func (h *Handler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	// 简单的 string 转 uint
	idInt, _ := strconv.Atoi(idStr)
	id := uint(idInt)

	// 🔥 调用 Repo 的混合删除逻辑 (Content硬删、Product硬删、UserProduct硬删)
	if err := h.repo.DeleteProduct(id); err != nil {
		c.JSON(500, gin.H{"error": "删除失败：" + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "商品已下架：内容绑定已彻底清除，用户记录已归档"})
}

// --- 2. 内容绑定管理 (Binding) ---

// BindContent 往商品里装题库科目
func (h *Handler) BindContent(c *gin.Context) {
	var req struct { ProductID uint `json:"product_id"`; Source string `json:"source"`; Category string `json:"category"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }
	
	// 查重
	var count int64
	db.DB.Model(&ProductContent{}).Where("product_id = ? AND source = ? AND category = ?", req.ProductID, req.Source, req.Category).Count(&count)
	if count > 0 { c.JSON(200, gin.H{"message": "已存在，无需重复添加"}); return }

	pc := ProductContent{ProductID: req.ProductID, Source: req.Source, Category: req.Category}
	db.DB.Create(&pc)
	c.JSON(200, gin.H{"message": "绑定成功"})
}

// UnbindContent 把科目从商品里拿出来
func (h *Handler) UnbindContent(c *gin.Context) {
	var req struct { ProductID uint `json:"product_id"`; Source string `json:"source"`; Category string `json:"category"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }
	
	// 🔥 必须使用 Unscoped 进行硬删除，不留垃圾
	db.DB.Unscoped().Where("product_id = ? AND source = ? AND category = ?", req.ProductID, req.Source, req.Category).Delete(&ProductContent{})
	
	c.JSON(200, gin.H{"message": "解绑成功"})
}

// GetProductContents 查看某个商品里装了啥
func (h *Handler) GetProductContents(c *gin.Context) {
	pid := c.Param("id")
	var list []ProductContent
	db.DB.Where("product_id = ?", pid).Find(&list)
	c.JSON(200, gin.H{"data": list})
}

// --- 3. 用户授权管理 (Granting) + 审计日志 ---

// 辅助：获取用户名
func getUserName(uid uint) string {
	var user struct { Username string }
	// 假设您的用户表叫 users
	if err := db.DB.Table("users").Select("username").Where("id = ?", uid).First(&user).Error; err != nil {
		return "未知用户"
	}
	return user.Username
}

// GrantProductToUser 给用户发证 (核心接口 - 带审计)
func (h *Handler) GrantProductToUser(c *gin.Context) {
	var req struct { 
		UserID uint `json:"user_id"`; ProductID uint `json:"product_id"`
		DurationDays int `json:"duration_days"` // 授权几天
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }

	// 1. 获取操作员信息
	opID := c.GetUint("userID") 
	opName := c.GetString("username") 
	if opName == "" { opName = "System/Unknown" }

	// 2. 查商品 (快照用)
	var product Product
	if err := db.DB.First(&product, req.ProductID).Error; err != nil {
		c.JSON(404, gin.H{"error": "商品不存在"}); return
	}

	// 3. 查目标客户用户名 (快照用)
	targetUserName := getUserName(req.UserID)

	// 4. 开启事务
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var up UserProduct
		res := tx.Where("user_id = ? AND product_id = ?", req.UserID, req.ProductID).First(&up)
		
		now := time.Now()
		newExpire := now.AddDate(0, 0, req.DurationDays)

		// A. 执行授权逻辑
		if res.Error == nil {
			// 续期
			if up.ExpireAt.After(now) {
				newExpire = up.ExpireAt.AddDate(0, 0, req.DurationDays)
			}
			up.ExpireAt = newExpire
			up.ProductName = product.Name // 更新快照
			if err := tx.Save(&up).Error; err != nil { return err }
		} else {
			// 新增
			up = UserProduct{
				UserID:      req.UserID,
				ProductID:   req.ProductID,
				ExpireAt:    newExpire,
				ProductName: product.Name, // 📸 写入快照
			}
			if err := tx.Create(&up).Error; err != nil { return err }
		}

		// B. 🔥 写入审计日志 (GRANT)
		log := ProductAuthLog{
			OperatorID:     opID,
			OperatorName:   opName,
			TargetUserID:   req.UserID,
			TargetUserName: targetUserName, // 🔥 写入客户名
			Action:         "GRANT",
			ProductID:      req.ProductID,
			ProductName:    product.Name,
			DurationDays:   req.DurationDays,
			ExpireAt:       newExpire,
		}
		if err := tx.Create(&log).Error; err != nil { return err }

		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "授权失败：" + err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": fmt.Sprintf("已授权商品：%s", product.Name)})
}

// RevokeUserProduct 收回凭证 (带审计 + 硬删除)
func (h *Handler) RevokeUserProduct(c *gin.Context) {
	var req struct { UserID uint `json:"user_id"`; ProductID uint `json:"product_id"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }

	// 1. 获取操作员
	opID := c.GetUint("userID")
	opName := c.GetString("username")
	if opName == "" { opName = "System/Unknown" }

	// 2. 先查记录 (为了拿 ProductName 写日志)
	var up UserProduct
	if err := db.DB.Where("user_id = ? AND product_id = ?", req.UserID, req.ProductID).First(&up).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户未持有该商品或已失效"}); return
	}

	// 3. 查目标客户用户名
	targetUserName := getUserName(req.UserID)

	// 4. 事务
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// A. 🔥 硬删除凭证 (不留软删除尸体)
		if err := tx.Unscoped().Delete(&up).Error; err != nil { return err }

		// B. 🔥 写入审计日志 (REVOKE)
		log := ProductAuthLog{
			OperatorID:     opID,
			OperatorName:   opName,
			TargetUserID:   req.UserID,
			TargetUserName: targetUserName, 
			Action:         "REVOKE",
			ProductID:      req.ProductID,
			ProductName:    up.ProductName, // 用 UserProduct 里的快照名
			DurationDays:   0,
			ExpireAt:       up.ExpireAt, // 记录一下当时原本是啥时候过期的
		}
		if err := tx.Create(&log).Error; err != nil { return err }

		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "收回失败：" + err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "已收回权限"})
}

// GetUserProducts 查看用户有哪些证
func (h *Handler) GetUserProducts(c *gin.Context) {
    uid := c.Param("id") 

    var list []UserProduct
    // Preload Product 只是为了兜底，优先展示 list[i].ProductName
    db.DB.Preload("Product").Where("user_id = ? AND expire_at > ?", uid, time.Now()).Find(&list)
    c.JSON(200, gin.H{"data": list})
}

// GetAuthLogs 查询审计日志
func (h *Handler) GetAuthLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	operatorId := c.Query("operator_id")
	targetId := c.Query("target_id")

	var logs []ProductAuthLog
	var total int64
	
	query := db.DB.Model(&ProductAuthLog{})
	if operatorId != "" { query = query.Where("operator_id = ?", operatorId) }
	if targetId != "" { query = query.Where("target_user_id = ?", targetId) }

	query.Count(&total)
	query.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

	c.JSON(200, gin.H{"data": logs, "total": total})
}