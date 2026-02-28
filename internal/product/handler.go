package product

import (
	"errors"
	"fmt"
	"med-platform/internal/common/db"
	"strconv"
	"time"
	"med-platform/internal/common/uploader"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strings"
	"gorm.io/gorm/clause"
)

type Handler struct {
	repo *Repository
}

func NewHandler() *Handler {
	return &Handler{repo: NewRepository()}
}

// ==========================================
// 🛒 1. 商品基础管理 (增改查)
// ==========================================

// CreateProduct 创建商品 (🔥 增加图片转正逻辑)
func (h *Handler) CreateProduct(c *gin.Context) {
	type SkuReq struct {
		Name         string `json:"name" binding:"required"`
		Points       int    `json:"points"`
		DurationDays int    `json:"duration_days" binding:"required"`
	}
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		CoverImg    string   `json:"cover_img"` // 这里接收的是 /uploads/temp/...
		Category    string   `json:"category"`
		Tags        string   `json:"tags"`
		Detail      string   `json:"detail"`
		Skus        []SkuReq `json:"skus"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 🔥🔥🔥 核心逻辑：图片转正
	// 如果路径中包含 temp，说明是刚上传的，将其移动到 products 目录
	if req.CoverImg != "" && strings.Contains(req.CoverImg, "/uploads/temp/") {
		finalPaths := uploader.ConfirmImages([]string{req.CoverImg}, "products")
		if len(finalPaths) > 0 {
			req.CoverImg = finalPaths[0] // 更新为永久路径：/uploads/products/...
		}
	}

	p := Product{
		Name:        req.Name,
		Description: req.Description,
		CoverImg:    req.CoverImg,
		Category:    req.Category,
		Tags:        req.Tags,
		Detail:      req.Detail,
		IsOnShelf:   true,
	}

	for _, s := range req.Skus {
		p.Skus = append(p.Skus, ProductSku{Name: s.Name, Points: s.Points, DurationDays: s.DurationDays})
	}

	if err := db.DB.Create(&p).Error; err != nil {
		c.JSON(500, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(200, gin.H{"message": "商品创建成功", "data": p})
}

// UpdateProduct 更新商品 (🔥 同样增加图片转正逻辑)
func (h *Handler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		CoverImg    string `json:"cover_img"`
		Category    string `json:"category"`
		Tags        string `json:"tags"`
		Detail      string `json:"detail"`
		IsOnShelf   bool   `json:"is_on_shelf"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 🔥🔥🔥 核心逻辑：图片转正
	if req.CoverImg != "" && strings.Contains(req.CoverImg, "/uploads/temp/") {
		finalPaths := uploader.ConfirmImages([]string{req.CoverImg}, "products")
		if len(finalPaths) > 0 {
			req.CoverImg = finalPaths[0]
		}
	}

	if err := db.DB.Model(&Product{}).Where("id = ?", id).Updates(req).Error; err != nil {
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(200, gin.H{"message": "更新成功"})
}

// ListProducts 获取商品列表 (🔥 核心优化：分页、分类筛选、排除大文本 Detail 提升性能)
func (h *Handler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	adminView := c.Query("admin") == "1" // 管理员视角看全部

	query := db.DB.Model(&Product{}).Preload("Skus").Order("id desc")
	
	// 只有非管理员视角才只看上架的商品
	if !adminView {
		query = query.Where("is_on_shelf = ?", true)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var total int64
	query.Count(&total)

	var list []Product
	// ⚠️ Omit("Detail")：在列表页坚决不查详情文本，极大降低网络带宽占用
	query.Omit("Detail").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	c.JSON(200, gin.H{"data": list, "total": total, "page": page})
}

// ListMarketProducts 供前端调用的市场列表
func (h *Handler) ListMarketProducts(c *gin.Context) {
	c.Request.URL.RawQuery = c.Request.URL.RawQuery + "&admin=0" // 强制非管理员视角
	h.ListProducts(c)
}

// GetProductDetail 获取商品详情 (🔥 新增接口，前端点进商品页时调用)
func (h *Handler) GetProductDetail(c *gin.Context) {
	id := c.Param("id")
	var p Product
	if err := db.DB.Preload("Skus").First(&p, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "未找到该商品"})
		return
	}
	c.JSON(200, gin.H{"data": p})
}

// DeleteProduct 删除商品
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.repo.DeleteProduct(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "删除成功"})
}


// ==========================================
// 🛡️ 2. 核心兑换逻辑 (高并发安全)
// ==========================================

// ExchangeProduct 积分兑换商品 (🔥🔥🔥 封堵各种羊毛漏洞)
func (h *Handler) ExchangeProduct(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var req struct {
		SkuID uint `json:"sku_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var sku ProductSku
		if err := tx.First(&sku, req.SkuID).Error; err != nil {
			return errors.New("商品规格不存在")
		}

		// 🚨 安全拦截 1：防负数越权
		if sku.Points < 0 {
			return errors.New("非法规格：积分异常")
		}

		var prod Product
		if err := tx.First(&prod, sku.ProductID).Error; err != nil {
			return errors.New("商品数据异常")
		}
		if !prod.IsOnShelf {
			return errors.New("该商品已下架")
		}

		// 🚨 安全拦截 2：防 0 元购的无限刷单
		if sku.Points == 0 {
			var count int64
			tx.Model(&ExchangeRecord{}).Where("user_id = ? AND sku_id = ?", userID, sku.ID).Count(&count)
			if count > 0 {
				return errors.New("限时免费商品每人仅限兑换一次哦！")
			}
		}

		// 查询用户现有凭证
		var existingUserProd UserProduct
		err := tx.Where("user_id = ? AND product_id = ?", userID, prod.ID).
			Order("expire_at desc").
			First(&existingUserProd).Error
		hasExisting := err == nil

		// 🚨 安全拦截 3：防终身会员的重复购买叠加
		if hasExisting && existingUserProd.ExpireAt.Year() >= 2099 {
			return errors.New("您已永久解锁该商品，无需重复兑换！")
		}

		// 锁住用户积分（防高并发双花）
		type Buyer struct {
			ID     uint
			Points int
		}
		var u Buyer
		if err := tx.Table("users").Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, userID).Error; err != nil {
			return err
		}

		if u.Points < sku.Points {
			return fmt.Errorf("积分不足，需要 %d，当前仅有 %d", sku.Points, u.Points)
		}

		// 扣除积分
		if err := tx.Table("users").Where("id = ?", userID).Update("points", u.Points-sku.Points).Error; err != nil {
			return err
		}

		// 发放权益
		if !hasExisting {
			// 首次购买
			newExpire := time.Now().AddDate(0, 0, sku.DurationDays)
			if sku.DurationDays == -1 {
				newExpire = time.Date(2099, 12, 31, 23, 59, 59, 0, time.Local)
			}
			newUp := UserProduct{
				UserID:      userID,
				ProductID:   prod.ID,
				ProductName: prod.Name,
				ExpireAt:    newExpire,
			}
			if err := tx.Create(&newUp).Error; err != nil {
				return err
			}
		} else {
			// 续费逻辑
			var newExpireAt time.Time
			if sku.DurationDays == -1 {
				newExpireAt = time.Date(2099, 12, 31, 23, 59, 59, 0, time.Local)
			} else {
				if existingUserProd.ExpireAt.After(time.Now()) {
					// 还没过期，在原来的基础上加
					newExpireAt = existingUserProd.ExpireAt.AddDate(0, 0, sku.DurationDays)
				} else {
					// 已经过期，从今天开始重新算
					newExpireAt = time.Now().AddDate(0, 0, sku.DurationDays)
				}
			}
			if err := tx.Model(&existingUserProd).Update("expire_at", newExpireAt).Error; err != nil {
				return err
			}
		}

		// 记录兑换流水
		exchangeLog := ExchangeRecord{
			UserID:      userID,
			ProductID:   prod.ID,
			SkuID:       sku.ID,
			ProductName: prod.Name,
			SkuName:     sku.Name,
			PointsPaid:  sku.Points,
		}
		return tx.Create(&exchangeLog).Error
	})

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "兑换成功，快去学习吧！"})
}


// ==========================================
// 🔗 3. 内容绑定与授权分配 (后台功能)
// ==========================================

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
	db.DB.Model(&Product{}).Where("id = ?", req.ProductID).Count(&count)
	if count == 0 {
		c.JSON(404, gin.H{"error": "商品不存在"})
		return
	}
	err := db.DB.Create(&ProductContent{ProductID: req.ProductID, Source: req.Source, Category: req.Category}).Error
	if err != nil {
		c.JSON(500, gin.H{"error": "绑定失败"})
		return
	}
	c.JSON(200, gin.H{"message": "绑定成功"})
}

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

func (h *Handler) GetProductContents(c *gin.Context) {
	id := c.Param("id")
	var list []ProductContent
	if err := db.DB.Where("product_id = ?", id).Find(&list).Error; err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(200, gin.H{"data": list})
}

func (h *Handler) GrantProductToUser(c *gin.Context) {
	var req struct {
		UserID       uint   `json:"user_id"`
		ProductID    uint   `json:"product_id"`
		DurationDays int    `json:"duration_days"`
		Reason       string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	opID := c.MustGet("userID").(uint)
	opName := c.MustGet("username").(string)

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var prod Product
		if err := tx.First(&prod, req.ProductID).Error; err != nil {
			return fmt.Errorf("商品不存在")
		}

		var targetUserName string
		if err := tx.Table("users").Select("username").Where("id = ?", req.UserID).Scan(&targetUserName).Error; err != nil {
			return fmt.Errorf("用户不存在")
		}

		var newExpire time.Time
		if req.DurationDays == -1 {
			newExpire = time.Date(2099, 12, 31, 23, 59, 59, 0, time.Local)
		} else {
			var exist UserProduct
			err := tx.Where("user_id = ? AND product_id = ?", req.UserID, req.ProductID).
				Order("expire_at desc").First(&exist).Error
			
			if err == nil && exist.ExpireAt.After(time.Now()) {
				newExpire = exist.ExpireAt.AddDate(0, 0, req.DurationDays)
			} else {
				newExpire = time.Now().AddDate(0, 0, req.DurationDays)
			}
		}

		var up UserProduct
		if err := tx.Where("user_id = ? AND product_id = ?", req.UserID, req.ProductID).First(&up).Error; err == nil {
			up.ExpireAt = newExpire
			up.ProductName = prod.Name 
			if err := tx.Save(&up).Error; err != nil {
				return err
			}
		} else {
			up = UserProduct{UserID: req.UserID, ProductID: req.ProductID, ProductName: prod.Name, ExpireAt: newExpire}
			if err := tx.Create(&up).Error; err != nil {
				return err
			}
		}

		log := ProductAuthLog{
			OperatorID: opID, OperatorName: opName, TargetUserID: req.UserID, TargetUserName: targetUserName,
			Action: "GRANT", ProductID: req.ProductID, ProductName: prod.Name, DurationDays: req.DurationDays, ExpireAt: newExpire, Memo: req.Reason,
		}
		return tx.Create(&log).Error
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "授权失败: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "后台授权成功"})
}

func (h *Handler) RevokeUserProduct(c *gin.Context) {
	var req struct {
		UserID    uint   `json:"user_id"`
		ProductID uint   `json:"product_id"`
		Reason    string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	opID := c.MustGet("userID").(uint)
	opName := c.MustGet("username").(string)

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var up UserProduct
		if err := tx.Where("user_id = ? AND product_id = ?", req.UserID, req.ProductID).First(&up).Error; err != nil {
			return fmt.Errorf("用户未持有该商品")
		}
		var targetUserName string
		tx.Table("users").Select("username").Where("id = ?", req.UserID).Scan(&targetUserName)

		if err := tx.Unscoped().Delete(&up).Error; err != nil {
			return err
		}
		log := ProductAuthLog{
			OperatorID: opID, OperatorName: opName, TargetUserID: req.UserID, TargetUserName: targetUserName,
			Action: "REVOKE", ProductID: req.ProductID, ProductName: up.ProductName, DurationDays: 0, ExpireAt: time.Now(), Memo: req.Reason,
		}
		return tx.Create(&log).Error
	})
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "已成功收回用户权限"})
}

func (h *Handler) GetUserProducts(c *gin.Context) {
	uid := c.Param("id")
	var list []UserProduct
	db.DB.Preload("Product").Where("user_id = ? AND expire_at > ?", uid, time.Now()).Find(&list)
	c.JSON(200, gin.H{"data": list})
}

func (h *Handler) GetAuthLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	targetUserID := c.Query("user_id")
	var logs []ProductAuthLog
	var total int64
	query := db.DB.Model(&ProductAuthLog{})
	if targetUserID != "" { query = query.Where("target_user_id = ?", targetUserID) }
	query.Count(&total)
	query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&logs)
	c.JSON(200, gin.H{"data": logs, "total": total})
}

// UploadCover 处理商品封面上传 (初始保存在 temp)
func (h *Handler) UploadCover(c *gin.Context) {
	// 默认 SaveImageWithHash 会存入 /uploads/temp
	url, err := uploader.SaveImageWithHash(c, "file", 5*1024*1024)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"url": url})
}