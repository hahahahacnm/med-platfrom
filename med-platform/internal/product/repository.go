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
	
	// 逻辑：
	// 1. 找 user_products (必须是存在的记录)
	// 2. 关联 product_contents (商品内容必须存在)
	// 3. 校验有效期
	
	err := db.DB.Table("user_products").
		Joins("JOIN product_contents ON user_products.product_id = product_contents.product_id").
		Where("user_products.user_id = ?", userID).
		Where("user_products.expire_at > ?", time.Now()). 
		
		// 虽然我们现在主要用硬删除，但保留这两行 deleted_at 检查是良好的防御性编程习惯。
		// 万一将来某个地方误用了软删除，这里依然能守住底线，防止已删除的凭证被使用。
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
// 🔥🔥🔥 [最终版] 删除商品专用逻辑 🔥🔥🔥
// ==========================================
func (r *Repository) DeleteProduct(productID uint) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 【硬删除】商品内容绑定 (ProductContent)
		// 配置数据，不要了就彻底删掉
		if err := tx.Unscoped().Where("product_id = ?", productID).Delete(&ProductContent{}).Error; err != nil {
			return err
		}

		// 🔥🔥🔥 2. [新增] 【硬删除】商品规格 (ProductSku) 🔥🔥🔥
		// 商品都没了，在这个商品下定义的“月卡”、“年卡”规格也必须删掉
		if err := tx.Unscoped().Where("product_id = ?", productID).Delete(&ProductSku{}).Error; err != nil {
			return err
		}

		// 3. 【硬删除】用户持有记录 (UserProduct)
		// 之前版本：软删除 (为了留证)。
		// 当前版本：硬删除 (Unscoped)。
		// 原因：因为我们已经有了 ProductAuthLog 审计表，所有的历史记录、被删记录都在那里查。
		// UserProduct 表只保留“当前有效”的记录，保持数据库轻量洁净。
		if err := tx.Unscoped().Where("product_id = ?", productID).Delete(&UserProduct{}).Error; err != nil {
			return err
		}

		// 4. 【硬删除】商品本身 (Product)
		// 商品定义彻底删除
		if err := tx.Unscoped().Delete(&Product{}, productID).Error; err != nil {
			return err
		}

		return nil
	})
}