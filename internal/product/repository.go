package product

import (
	"med-platform/internal/common/db"
	"time"

	"gorm.io/gorm"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// ==========================================
// 🔐 核心鉴权逻辑 (CheckPermission)
// ==========================================
func (r *Repository) CheckPermission(userID uint, source string, category string) bool {
	var count int64
	
	err := db.DB.Table("user_products").
		Joins("JOIN product_contents ON user_products.product_id = product_contents.product_id").
		Where("user_products.user_id = ?", userID).
		Where("user_products.expire_at > ?", time.Now()). 
		Where("user_products.deleted_at IS NULL").      
		Where("product_contents.deleted_at IS NULL").   
		Where("product_contents.source = ?", source).
		Where("product_contents.category = ?", category).
		Count(&count).Error

	return err == nil && count > 0
}

// ==========================================
// 🧹 级联清理逻辑 (Source/Category 删除时)
// ==========================================
func (r *Repository) CleanUpBySource(source string) error {
	return db.DB.Unscoped().Where("source = ?", source).Delete(&ProductContent{}).Error
}

func (r *Repository) CleanUpByCategory(source string, category string) error {
	return db.DB.Unscoped().Where("source = ? AND category = ?", source, category).Delete(&ProductContent{}).Error
}

// ==========================================
// 🛡️ [最终版] 安全删除商品逻辑
// ==========================================
func (r *Repository) DeleteProduct(productID uint) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 强制下架商品
		if err := tx.Model(&Product{}).Where("id = ?", productID).Update("is_on_shelf", false).Error; err != nil {
			return err
		}

		// 2. 【软删除】商品本身 (Product)
		// 为什么是软删除？因为用户的 UserProduct 表里还存着 product_id，
		// 如果硬删除，前端在查“我的商品”时，关联查询 Preload("Product") 就会找不到数据而报错。
		if err := tx.Delete(&Product{}, productID).Error; err != nil {
			return err
		}

		// 3. 【软删除】商品规格 (ProductSku)
		// 阻止任何人通过旧的 sku_id 再次尝试发起兑换
		if err := tx.Where("product_id = ?", productID).Delete(&ProductSku{}).Error; err != nil {
			return err
		}

		// 🔥🔥🔥 核心修正区：以下两项绝对不能删！🔥🔥🔥
		
		// 🚫 不要删除 ProductContent：
		// 就算商品不卖了，但以前买过的人还需要靠这层映射关系去解锁题库 (CheckPermission 依赖它)。
		
		// 🚫 不要删除 UserProduct：
		// 用户的资产神圣不可侵犯，只要没到 ExpireAt 过期时间，这笔资产就必须躺在用户的背包里。

		return nil
	})
}