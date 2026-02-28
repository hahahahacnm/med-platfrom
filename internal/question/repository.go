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
// 💡 注意：前端题库做题已切换至 Skeleton + 单题 GetDetail 模式。
// 当前 List 接口主要供“后台管理面板”或“全局关键词搜索”使用。
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
		// 模式 A：后台无条件随便看看 (只看大题，防止子题刷屏)
		query = query.Where("parent_id IS NULL OR parent_id = 0")
	} else {
		// 模式 B：按章节或关键词搜索
		if keyword != "" {
			// 搜索模式：保持宽泛匹配
			likeStr := "%" + keyword + "%"
			query = query.Where("stem LIKE ? OR analysis LIKE ?", likeStr, likeStr)
		} else {
			// 按章节浏览模式 (排除纯父题壳子，防止计数虚高)
			query = query.Where("category_path LIKE ?", category+"%")
			groupTypes := []string{"A3", "A4", "B1"}
			query = query.Where("NOT ((parent_id IS NULL OR parent_id = 0) AND type IN ?)", groupTypes)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询 (按 ID 倒序)
	err := query.Order("id asc").Offset(offset).Limit(pageSize).Find(&questions).Error
	return questions, total, err
}

// 2. 基础详情 (单题加载模式的核心支撑)
// 🔥 强化点：通过 Preload("Children") 确保 A3/B1 等组合题被完整拉出
// 🔥 强化点：通过 Preload("Parent") 确保直接请求子题时能向上追溯公共题干
func (r *Repository) GetDetail(id uint) (*Question, error) {
	var q Question
	// 使用 Unscoped 以支持后台回收站预览
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
	ID           uint            `json:"id"`
	Name         string          `json:"name"`
	Full         string          `json:"full"`
	SortOrder    int             `json:"sort_order"`
	Level        int             `json:"level"`
	IsLeaf       bool            `json:"is_leaf"`
	TotalCount   int64           `json:"total_count"`   // 本分类下总题数
	DoneCount    int64           `json:"done_count"`    // 当前用户已做题数 (去重后的题量)
	CorrectCount int64           `json:"correct_count"` // 🔥 新增：当前用户已答对的题数
	Children     []*CategoryNode `json:"children"`
}

// GetTree 获取目录树 (进度统计强化版)
func (r *Repository) GetTree(parentID *int, source string, userID uint) ([]*CategoryNode, error) {
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
	query = query.Where("level <= ?", MaxLevel)

	if err := query.Find(&cats).Error; err != nil {
		return nil, err
	}

var nodes []*CategoryNode
	for _, c := range cats {
		pathPattern := c.FullPath + "%"

		// 1. 统计总题数 (口径：所有子题 + 独立单题)
		var total int64
		db.DB.Table("questions").
			Where("source = ? AND category_path LIKE ? AND deleted_at IS NULL", source, pathPattern).
			Where("(parent_id > 0 OR (type NOT LIKE 'A3%' AND type NOT LIKE 'A4%' AND type NOT LIKE 'B1%'))").
			Count(&total)

		// 2. 统计已做题数
		var done int64
		if userID > 0 && total > 0 {
			db.DB.Table("answer_records").
				Joins("JOIN questions ON answer_records.question_id = questions.id").
				Where("answer_records.user_id = ?", userID).
				Where("questions.source = ?", source).
				Where("questions.category_path LIKE ?", pathPattern).
				Where("questions.deleted_at IS NULL").
				Where("(questions.parent_id > 0 OR (questions.type NOT LIKE 'A3%' AND questions.type NOT LIKE 'A4%' AND questions.type NOT LIKE 'B1%'))").
				Select("COUNT(DISTINCT answer_records.question_id)").
				Scan(&done)
		}

		// 3. 🔥 新增：统计已答对题数
		var correct int64
		if userID > 0 && done > 0 {
			db.DB.Table("answer_records").
				Joins("JOIN questions ON answer_records.question_id = questions.id").
				Where("answer_records.user_id = ? AND answer_records.is_correct = ?", userID, true).
				Where("questions.source = ?", source).
				Where("questions.category_path LIKE ?", pathPattern).
				Where("questions.deleted_at IS NULL").
				Where("(questions.parent_id > 0 OR (questions.type NOT LIKE 'A3%' AND questions.type NOT LIKE 'A4%' AND questions.type NOT LIKE 'B1%'))").
				Select("COUNT(DISTINCT answer_records.question_id)").
				Scan(&correct)
		}

		isLeaf := false
		if c.Level >= MaxLevel {
			isLeaf = true
		} else {
			var subCount int64
			db.DB.Model(&Category{}).Where("parent_id = ? AND level <= ?", c.ID, MaxLevel).Count(&subCount)
			isLeaf = (subCount == 0)
		}
		
		nodes = append(nodes, &CategoryNode{
			ID:           c.ID,
			Name:         c.Name,
			Full:         c.FullPath,
			SortOrder:    c.SortOrder,
			Level:        c.Level,
			IsLeaf:       isLeaf,
			TotalCount:   total,
			DoneCount:    done,
			CorrectCount: correct, // 🔥 填入正确数
		})
	}
	return nodes, nil
}

// SyncCategories 同步并修复目录结构
func (r *Repository) SyncCategories() error {
	type PathInfo struct {
		CategoryPath string
		Source       string
	}
	var pathInfos []PathInfo
	// 过滤掉包含非法字符的路径
	db.DB.Model(&Question{}).
		Select("DISTINCT category_path, source").
		Where("category_path != '' AND category_path NOT LIKE '%【%'").
		Scan(&pathInfos)

	for _, info := range pathInfos {
		parts := strings.Split(info.CategoryPath, " > ")
		var parentID *uint

		for i, part := range parts {
			partName := strings.TrimSpace(part)
			if partName == "" {
				continue
			}

			var cat Category
			var err error
			
			query := db.DB.Where("name = ? AND source = ?", partName, info.Source)
			if parentID == nil {
				query = query.Where("parent_id IS NULL")
			} else {
				query = query.Where("parent_id = ?", *parentID)
			}
			
			err = query.First(&cat).Error

			if err != nil { // 没找到则创建
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
					FullPath:  "", 
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

	pathMap := make(map[uint]string)

	for _, cat := range allCats {
		correctPath := ""

		if cat.ParentID == nil {
			correctPath = cat.Name
		} else {
			if parentPath, ok := pathMap[*cat.ParentID]; ok {
				correctPath = parentPath + " > " + cat.Name
			} else {
				var parentCat Category
				db.DB.First(&parentCat, *cat.ParentID)
				correctPath = parentCat.FullPath + " > " + cat.Name
			}
		}

		pathMap[cat.ID] = correctPath

		if cat.FullPath != correctPath {
			db.DB.Model(&cat).Updates(map[string]interface{}{
				"full_path": correctPath,
			})
			fmt.Printf("🔧 修复目录: ID=%d, Name=%s, Path=%s\n", cat.ID, cat.Name, correctPath)
		}
	}

	return nil
}

type UpdateCategoryReq struct {
	Name      string `json:"name"`
	SortOrder *int   `json:"sort_order"`
	IsDirty   *bool  `json:"is_dirty"`
}

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

// 5. 排序操作
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