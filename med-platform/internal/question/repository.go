package question

import (
	"fmt"
	"strings"

	"med-platform/internal/common/db"

	"gorm.io/gorm"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// 1. 题目查询 (List)
func (r *Repository) List(page, pageSize int, category, keyword, source string) ([]Question, int64, error) {
	var questions []Question
	var total int64
	offset := (page - 1) * pageSize
	
	// 使用 Preload 加载子题
	query := db.DB.Model(&Question{}).Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Order("id asc") })

	// 软删除过滤
	query = query.Where("deleted_at IS NULL")

	if source != "" {
		query = query.Where("source = ?", source)
	}

	// 判断是否处于“具体内容搜索”模式
	isSearchingSpecifics := keyword != "" || category != ""

	if !isSearchingSpecifics {
		// 模式 A：首页随便看看 (只看父题，防止子题刷屏)
		query = query.Where("parent_id IS NULL")
	} else {
		// 模式 B：浏览具体章节 或 搜索关键词
		if keyword != "" {
			// 如果是搜关键词，保持宽泛，只要匹配就显示 (哪怕多显示几个父题也没关系，主要是为了搜到)
			likeStr := "%" + keyword + "%"
			query = query.Where("stem LIKE ? OR analysis LIKE ?", likeStr, likeStr)
		} else {
			// 🔥🔥🔥 核心修复：浏览章节时的计数修正 🔥🔥🔥
			// 现象：Total = 129 (实际 115)。原因：把“A3/A4/B1 的父题壳子”也算进去了。
			// 修复：我们只查“能做的题” (即：单题 + 子题)。
			// 逻辑：排除掉 (没有父亲 且 是组合题型) 的记录。
			// 这样 Handler 依然能通过子题找到父题，但 Total 计数只会统计子题数量。
			
			query = query.Where("category_path LIKE ?", category+"%")
			
			// 排除“纯父题壳子”
			// 只有当 parent_id IS NULL (是父题) 且 Type 是组合题代码时，才排除
			groupTypes := []string{"A3", "A4", "B1", "案例", "案例分析"}
			query = query.Where("NOT (parent_id IS NULL AND type IN ?)", groupTypes)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询 (按 ID 倒序)
	err := query.Order("id asc").Offset(offset).Limit(pageSize).Find(&questions).Error
	return questions, total, err
}
// 2. 基础详情
func (r *Repository) GetDetail(id uint) (*Question, error) {
	var q Question
	// 使用 Unscoped 以支持回收站预览
	err := db.DB.Unscoped().
		Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Unscoped().Order("id asc") }).
		Preload("Parent").
		First(&q, id).Error
	return &q, err
}

func (r *Repository) GetAllPaths() ([]string, error) {
	var paths []string
	err := db.DB.Model(&Question{}).Where("category_path != ''").Distinct("category_path").Pluck("category_path", &paths).Error
	return paths, err
}

func (r *Repository) GetSources() ([]string, error) {
	var sources []string
	err := db.DB.Model(&Question{}).Where("source != ''").Distinct("source").Pluck("source", &sources).Error
	return sources, err
}

// ---------------------------------------------------------
// 3. 目录树逻辑 (Category)
// ---------------------------------------------------------

type CategoryNode struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name"`
	Full      string          `json:"full"`
	SortOrder int             `json:"sort_order"`
	Level     int             `json:"level"`
	IsLeaf    bool            `json:"is_leaf"`
	Children  []*CategoryNode `json:"children"`
}

// GetTree 获取目录树 (5级限制)
func (r *Repository) GetTree(parentID *int, source string) ([]*CategoryNode, error) {
	// 🔥 核心配置：最大显示层级
	const MaxLevel = 5

	var cats []Category
	query := db.DB.Order("sort_order asc").Order("id asc")

	if source != "" {
		query = query.Where("source = ?", source)
	}

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	// 🔥 物理过滤：只查 5 级及以内的目录
	query = query.Where("level <= ?", MaxLevel)

	if err := query.Find(&cats).Error; err != nil {
		return nil, err
	}

	var nodes []*CategoryNode
	for _, c := range cats {
		isLeaf := false

		// 🔥 智能判断叶子节点
		if c.Level >= MaxLevel {
			// 情况 A: 已经到了第 5 级 -> 强制设为叶子
			isLeaf = true
		} else {
			// 情况 B: 不到第 5 级 -> 检查是否有子节点
			var count int64
			subQuery := db.DB.Model(&Category{}).
				Where("parent_id = ?", c.ID).
				Where("level <= ?", MaxLevel)

			if source != "" {
				subQuery = subQuery.Where("source = ?", source)
			}
			subQuery.Count(&count)
			isLeaf = (count == 0)
		}

		node := &CategoryNode{
			ID:        c.ID,
			Name:      c.Name,
			Full:      c.FullPath,
			SortOrder: c.SortOrder,
			Level:     c.Level,
			IsLeaf:    isLeaf,
			Children:  nil, // 懒加载
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// SyncCategories 同步并修复目录结构 (强力修复版)
func (r *Repository) SyncCategories() error {
	// 1. 从题目表中提取所有路径，创建缺失节点
	type PathInfo struct {
		CategoryPath string
		Source       string
	}
	var pathInfos []PathInfo
	// 过滤掉包含非法字符的路径
	db.DB.Model(&Question{}).
		Select("DISTINCT category_path, source").
		Where("category_path != '' AND category_path NOT LIKE '%【%'"). // 简单过滤脏数据
		Scan(&pathInfos)

	for _, info := range pathInfos {
		parts := strings.Split(info.CategoryPath, " > ")
		var parentID *uint

		for i, part := range parts {
			partName := strings.TrimSpace(part)
			if partName == "" {
				continue
			}

			// 查找或创建节点
			// 注意：这里不应该每次都 Create，必须先 Check
			var cat Category
			var err error
			
			// 修正查询逻辑：不仅看名字，还得看 source 和 parent_id
			query := db.DB.Where("name = ? AND source = ?", partName, info.Source)
			if parentID == nil {
				query = query.Where("parent_id IS NULL")
			} else {
				query = query.Where("parent_id = ?", *parentID)
			}
			
			err = query.First(&cat).Error

			if err != nil { // 没找到
				sortOrder := 999
				if strings.Contains(partName, "绪论") || strings.Contains(partName, "总论") {
					sortOrder = 1
				}
				newCat := Category{
					Name:      partName,
					ParentID:  parentID,
					Level:     i + 1,
					SortOrder: sortOrder,
					Source:    info.Source,
					FullPath:  "", // 暂时留空，下面会统一修复
				}
				db.DB.Create(&newCat)
				parentID = &newCat.ID
			} else {
				parentID = &cat.ID
			}
		}
	}

	// 2. 递归重算并强制更新 FullPath
	var allCats []Category
	if err := db.DB.Order("level asc").Find(&allCats).Error; err != nil {
		return err
	}

	// 建立 ID -> FullPath 映射
	pathMap := make(map[uint]string)

	for _, cat := range allCats {
		correctPath := ""

		if cat.ParentID == nil {
			correctPath = cat.Name
		} else {
			if parentPath, ok := pathMap[*cat.ParentID]; ok {
				correctPath = parentPath + " > " + cat.Name
			} else {
				// 兜底查库
				var parentCat Category
				db.DB.First(&parentCat, *cat.ParentID)
				correctPath = parentCat.FullPath + " > " + cat.Name
			}
		}

		pathMap[cat.ID] = correctPath

		// 强制更新 (Fix Dirty Data)
		if cat.FullPath != correctPath {
			db.DB.Model(&cat).Updates(map[string]interface{}{
				"full_path": correctPath,
			})
			fmt.Printf("🔧 修复目录: ID=%d, Name=%s, Path=%s\n", cat.ID, cat.Name, correctPath)
		}
	}

	return nil
}

// UpdateCategoryReq
type UpdateCategoryReq struct {
	Name      string `json:"name"`
	SortOrder *int   `json:"sort_order"`
	IsDirty   *bool  `json:"is_dirty"`
}

// UpdateCategory
func (r *Repository) UpdateCategory(id uint, req UpdateCategoryReq) error {
	var cat Category
	if err := db.DB.First(&cat, id).Error; err != nil {
		return err
	}
	if req.Name != "" {
		cat.Name = req.Name
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}
	if req.IsDirty != nil {
		cat.IsDirty = *req.IsDirty
	}
	return db.DB.Save(&cat).Error
}

// 4. Admin Ops
func (r *Repository) RenameSource(oldName, newName string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Question{}).Where("source = ?", oldName).Update("source", newName).Error; err != nil {
			return err
		}
		if err := tx.Model(&Category{}).Where("source = ?", oldName).Update("source", newName).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) DeleteSource(sourceName string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("source = ?", sourceName).Delete(&Question{}).Error; err != nil {
			return err
		}
		if err := tx.Where("source = ?", sourceName).Delete(&Category{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) TransferCategorySource(from, to, cat string) error {
	return db.DB.Model(&Question{}).Where("source = ? AND category = ?", from, cat).Update("source", to).Error
}

// 5. Sort
type ReorderItem struct {
	ID        uint `json:"id"`
	SortOrder int  `json:"sort_order"`
}

func (r *Repository) ReorderCategories(items []ReorderItem) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Model(&Category{}).Where("id = ?", item.ID).Update("sort_order", item.SortOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}