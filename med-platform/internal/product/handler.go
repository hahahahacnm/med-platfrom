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

// --- 1. 商品管理 (Product + SKU) ---

// CreateProduct 创建商品壳子及规格
func (h *Handler) CreateProduct(c *gin.Context) {
	// 定义请求结构体
	type SkuReq struct {
		Name         string  `json:"name"`          // 规格名
		Price        float64 `json:"price"`         // 价格
		DurationDays int     `json:"duration_days"` // 时长
	}
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Skus        []SkuReq `json:"skus"` // 允许同时传 SKU 列表
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 构建模型
	p := Product{
		Name:        req.Name,
		Description: req.Description,
		IsOnShelf:   true,
		Skus:        []ProductSku{}, // 初始化
	}

	// 填充 SKUs
	for _, s := range req.Skus {
		p.Skus = append(p.Skus, ProductSku{
			Name:         s.Name,
			Price:        s.Price,
			DurationDays: s.DurationDays,
		})
	}

	// 事务创建
	if err := db.DB.Create(&p).Error; err != nil {
		c.JSON(500, gin.H{"error": "创建失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "商品创建成功", "data": p})
}

// UpdateProduct 更新商品 (支持修改 SKU 规格)
func (h *Handler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	// 定义请求结构
	type SkuReq struct {
		ID           uint    `json:"id"`            // 如果有ID，说明是更新；没有则是新增
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		DurationDays int     `json:"duration_days"`
	}
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		IsOnShelf   *bool    `json:"is_on_shelf"`
		Skus        []SkuReq `json:"skus"` // 接收 SKU 列表
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 开启事务处理 (保证原子性)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var p Product
		if err := tx.First(&p, id).Error; err != nil {
			return err
		}

		// 1. 更新商品基础信息
		if req.Name != "" {
			p.Name = req.Name
		}
		if req.Description != "" {
			p.Description = req.Description
		}
		// 上下架控制
		if req.IsOnShelf != nil {
			p.IsOnShelf = *req.IsOnShelf
		}
		if err := tx.Save(&p).Error; err != nil {
			return err
		}

		// 2. 🔥🔥🔥 核心：处理 SKU 规格的增删改 🔥🔥🔥

		// 步骤 A: 找出前端这次提交的所有 SKU ID (用于判断哪些要保留)
		keepIds := []uint{}
		for _, s := range req.Skus {
			if s.ID > 0 {
				keepIds = append(keepIds, s.ID)
			}
		}

		// 步骤 B: 删除那些“数据库里有，但前端没传”的 SKU (说明用户删掉了)
		if len(keepIds) > 0 {
			if err := tx.Where("product_id = ? AND id NOT IN ?", p.ID, keepIds).Delete(&ProductSku{}).Error; err != nil {
				return err
			}
		} else {
			// 如果前端一个旧ID都没传（keepIds为空），且 skus 不为空，说明全是新增；如果 skus 为空，说明全删
			// 这里简单处理：如果 req.Skus 为空，则删除所有；如果不为空，未传 ID 的是新增，已传的是保留。
			// 上面的 keepIds 逻辑已经涵盖了保留的部分。如果 keepIds 为空，确实应该删除所有旧的。
			if err := tx.Where("product_id = ?", p.ID).Delete(&ProductSku{}).Error; err != nil {
				return err
			}
		}

		// 步骤 C: 循环处理 新增 或 更新
		for _, s := range req.Skus {
			if s.ID > 0 {
				// === 更新 (Update) ===
				if err := tx.Model(&ProductSku{}).Where("id = ? AND product_id = ?", s.ID, p.ID).
					Updates(map[string]interface{}{
						"name":          s.Name,
						"price":         s.Price,
						"duration_days": s.DurationDays,
					}).Error; err != nil {
					return err
				}
			} else {
				// === 新增 (Create) ===
				newSku := ProductSku{
					ProductID:    p.ID,
					Name:         s.Name,
					Price:        s.Price,
					DurationDays: s.DurationDays,
				}
				if err := tx.Create(&newSku).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "更新失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "更新成功"})
}

// 管理员：ListProducts 查看所有商品 (包含 SKU 信息)
func (h *Handler) ListProducts(c *gin.Context) {
	var list []Product
	db.DB.Preload("Skus").Find(&list)
	c.JSON(200, gin.H{"data": list})
}

// ListMarketProducts 前台商城专用列表 (只返回上架商品)
func (h *Handler) ListMarketProducts(c *gin.Context) {
	var list []Product
	result := db.DB.Preload("Skus").Where("is_on_shelf = ?", true).Find(&list)
	
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "获取商品列表失败"})
		return
	}

	c.JSON(200, gin.H{"data": list})
}

// DeleteProduct 删除商品
func (h *Handler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	idInt, _ := strconv.Atoi(idStr)
	id := uint(idInt)

	if err := h.repo.DeleteProduct(id); err != nil {
		c.JSON(500, gin.H{"error": "删除失败：" + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "商品已下架：规格已清空，内容绑定已清除，用户记录已归档"})
}

// --- 2. 内容绑定管理 (Binding) ---

// BindContent 往商品里装题库科目
func (h *Handler) BindContent(c *gin.Context) {
	var req struct {
		ProductID uint   `json:"product_id"`
		Source    string `json:"source"`
		Category  string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var count int64
	db.DB.Model(&ProductContent{}).Where("product_id = ? AND source = ? AND category = ?", req.ProductID, req.Source, req.Category).Count(&count)
	if count > 0 {
		c.JSON(200, gin.H{"message": "已存在，无需重复添加"})
		return
	}

	pc := ProductContent{ProductID: req.ProductID, Source: req.Source, Category: req.Category}
	db.DB.Create(&pc)
	c.JSON(200, gin.H{"message": "绑定成功"})
}

// UnbindContent 解绑
func (h *Handler) UnbindContent(c *gin.Context) {
	var req struct {
		ProductID uint   `json:"product_id"`
		Source    string `json:"source"`
		Category  string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

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
	var user struct{ Username string }
	if err := db.DB.Table("users").Select("username").Where("id = ?", uid).First(&user).Error; err != nil {
		return "未知用户"
	}
	return user.Username
}

// GrantProductToUser 给用户发证 (核心接口 - 带审计)
func (h *Handler) GrantProductToUser(c *gin.Context) {
	var req struct {
		UserID       uint `json:"user_id"`
		ProductID    uint `json:"product_id"`
		DurationDays int  `json:"duration_days"` // 授权几天 (支持 -1 永久)
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 1. 获取操作员信息 (从中间件 AuthJWT 注入)
	opID := c.GetUint("userID")
	opName := c.GetString("username")
	if opName == "" {
		opName = "System/Unknown"
	}

	// 2. 查商品 (快照用)
	var product Product
	if err := db.DB.First(&product, req.ProductID).Error; err != nil {
		c.JSON(404, gin.H{"error": "商品不存在"})
		return
	}

	// 3. 查目标客户用户名 (快照用)
	targetUserName := getUserName(req.UserID)

	// 4. 开启事务
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var up UserProduct
		// 查找现有记录
		res := tx.Where("user_id = ? AND product_id = ?", req.UserID, req.ProductID).Order("expire_at desc").First(&up)

		now := time.Now()
		var newExpire time.Time

		// 处理永久授权 (-1)
		if req.DurationDays == -1 {
			newExpire = time.Date(2099, 12, 31, 23, 59, 59, 0, time.Local)
		} else {
			newExpire = now.AddDate(0, 0, req.DurationDays)
		}

		// A. 执行授权逻辑
		if res.Error == nil {
			// 续期逻辑
			if req.DurationDays == -1 {
				up.ExpireAt = newExpire
			} else {
				if up.ExpireAt.After(now) {
					// 还没过期：顺延
					up.ExpireAt = up.ExpireAt.AddDate(0, 0, req.DurationDays)
				} else {
					// 已过期：重新计算
					up.ExpireAt = newExpire
				}
			}

			up.ProductName = product.Name // 更新快照
			if err := tx.Save(&up).Error; err != nil {
				return err
			}
		} else {
			// 新增
			up = UserProduct{
				UserID:      req.UserID,
				ProductID:   req.ProductID,
				ExpireAt:    newExpire,
				ProductName: product.Name, 
			}
			if err := tx.Create(&up).Error; err != nil {
				return err
			}
		}

		// B. 写入审计日志 (GRANT)
		log := ProductAuthLog{
			OperatorID:     opID,
			OperatorName:   opName,
			TargetUserID:   req.UserID,
			TargetUserName: targetUserName,
			Action:         "GRANT",
			ProductID:      req.ProductID,
			ProductName:    product.Name,
			DurationDays:   req.DurationDays,
			ExpireAt:       up.ExpireAt, 
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "授权失败：" + err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": fmt.Sprintf("已授权商品：%s", product.Name)})
}

// RevokeUserProduct 收回凭证
func (h *Handler) RevokeUserProduct(c *gin.Context) {
	var req struct {
		UserID    uint `json:"user_id"`
		ProductID uint `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	opID := c.GetUint("userID")
	opName := c.GetString("username")
	if opName == "" {
		opName = "System/Unknown"
	}

	var up UserProduct
	if err := db.DB.Where("user_id = ? AND product_id = ?", req.UserID, req.ProductID).First(&up).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户未持有该商品或已失效"})
		return
	}
	targetUserName := getUserName(req.UserID)

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// 硬删除凭证
		if err := tx.Unscoped().Delete(&up).Error; err != nil {
			return err
		}

		// 写入日志
		log := ProductAuthLog{
			OperatorID:     opID,
			OperatorName:   opName,
			TargetUserID:   req.UserID,
			TargetUserName: targetUserName,
			Action:         "REVOKE",
			ProductID:      req.ProductID,
			ProductName:    up.ProductName,
			DurationDays:   0,
			ExpireAt:       up.ExpireAt,
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
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
	// Preload Product 只是为了兜底
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
	if operatorId != "" {
		query = query.Where("operator_id = ?", operatorId)
	}
	if targetId != "" {
		query = query.Where("target_user_id = ?", targetId)
	}

	query.Count(&total)
	query.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

	c.JSON(200, gin.H{"data": logs, "total": total})
}