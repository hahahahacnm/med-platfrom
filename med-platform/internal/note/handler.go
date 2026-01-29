package note

import (
	"net/http"
	"strconv"
	"time"

	"med-platform/internal/answer"
	"med-platform/internal/common/db"
	"med-platform/internal/common/logger"
	"med-platform/internal/question"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct{}

// ==========================================
// 1. 保存或更新笔记 (升级版：支持多条 + 回复)
// ==========================================
func (h *Handler) SaveNote(c *gin.Context) {
	var req struct {
		ID         uint   `json:"id"`
		QuestionID uint   `json:"question_id" binding:"required"`
		Content    string `json:"content" binding:"required"`
		IsPublic   bool   `json:"is_public"`
		ParentID   *uint  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)
	var n Note

	if req.ID > 0 {
		if err := db.DB.First(&n, req.ID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到原笔记"})
			return
		}
		if n.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权修改他人笔记"})
			return
		}
		n.Content = req.Content
		n.IsPublic = req.IsPublic
		db.DB.Save(&n)
	} else {
		n = Note{
			UserID:     userID,
			QuestionID: req.QuestionID,
			Content:    req.Content,
			IsPublic:   req.IsPublic,
			ParentID:   req.ParentID,
		}
		db.DB.Create(&n)
	}

	db.DB.Preload("User").Preload("Parent.User").First(&n, n.ID)
	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "data": n})
}

// 辅助函数
func getStartOfDay() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// ==========================================
// 2. 获取某题的评论区
// ==========================================
func (h *Handler) ListNotes(c *gin.Context) {
	qID := c.Query("question_id")
	if qID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少题目ID"})
		return
	}

	userID := c.MustGet("userID").(uint)
	var notes []Note

	err := db.DB.Preload("User").
		Preload("Parent").
		Preload("Parent.User").
		Where("question_id = ? AND (is_public = true OR user_id = ?)", qID, userID).
		Order("like_count desc, created_at desc").
		Find(&notes).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	if len(notes) > 0 {
		var noteIDs []uint
		for _, n := range notes {
			noteIDs = append(noteIDs, n.ID)
		}

		startOfDay := getStartOfDay()
		var likedNoteIDs []uint
		db.DB.Model(&NoteLike{}).
			Where("user_id = ? AND note_id IN ? AND created_at >= ?", userID, noteIDs, startOfDay).
			Pluck("note_id", &likedNoteIDs)
		
		likedMap := make(map[uint]bool)
		for _, id := range likedNoteIDs { likedMap[id] = true }

		var collectedNoteIDs []uint
		db.DB.Model(&NoteCollect{}).Where("user_id = ? AND note_id IN ?", userID, noteIDs).Pluck("note_id", &collectedNoteIDs)
		collectedMap := make(map[uint]bool)
		for _, id := range collectedNoteIDs { collectedMap[id] = true }

		for i := range notes {
			notes[i].IsLiked = likedMap[notes[i].ID]
			notes[i].IsCollected = collectedMap[notes[i].ID]
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": notes})
}

// ==========================================
// 3. 获取我的笔记本列表 (🔥 核心修复：寻根逻辑 🔥)
// ==========================================
func (h *Handler) GetMyNotes(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize
	
	categoryIDStr := c.Query("category_id")
	source := c.Query("source")

	// 1. 找到所有相关题目 ID (我的笔记 + 我收藏的笔记)
	var myNoteQIDs []uint
	db.DB.Model(&Note{}).Where("user_id = ?", userID).Distinct("question_id").Pluck("question_id", &myNoteQIDs)
	
	var collectedNoteQIDs []uint
	db.DB.Table("note_collects").
		Joins("JOIN notes ON note_collects.note_id = notes.id").
		Where("note_collects.user_id = ?", userID).
		Distinct("notes.question_id").
		Pluck("notes.question_id", &collectedNoteQIDs)

	idMap := make(map[uint]bool)
	for _, id := range myNoteQIDs { idMap[id] = true }
	for _, id := range collectedNoteQIDs { idMap[id] = true }
	
	var rawQIDs []uint
	for id := range idMap { rawQIDs = append(rawQIDs, id) }

	if len(rawQIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0, "page": page})
		return
	}

	// 2. 查询原始题目基础信息 (为了分页)
	var rawQuestions []question.Question
	var total int64

	query := db.DB.Model(&question.Question{}).Where("id IN ?", rawQIDs)

	if source != "" {
		query = query.Where("source = ?", source)
	}

	if categoryIDStr != "" && categoryIDStr != "0" {
		var cat question.Category
		if err := db.DB.First(&cat, categoryIDStr).Error; err == nil {
			query = query.Where("category_path LIKE ?", cat.FullPath+"%")
		}
	}

	query.Count(&total)

	// 分页查出当前页的题目 (可能是子题，也可能是单题)
	err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&rawQuestions).Error
	if err != nil {
		logger.Log.Error("查询笔记本题目失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	// 3. 🧠 寻根逻辑：批量还原大题
	var finalQIDs []uint
	parentIDMap := make(map[uint]uint)

	for _, q := range rawQuestions {
		if q.ParentID != nil && *q.ParentID > 0 {
			finalQIDs = append(finalQIDs, *q.ParentID)
			parentIDMap[q.ID] = *q.ParentID
		} else {
			finalQIDs = append(finalQIDs, q.ID)
		}
	}

	// 4. 加载全家桶
	var finalQuestions []question.Question
	db.DB.Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Order("id asc") }).
		Where("id IN ?", finalQIDs).
		Find(&finalQuestions)

	qMap := make(map[uint]question.Question)
	for _, q := range finalQuestions {
		qMap[q.ID] = q
	}

	// 5. 挂载附加信息 & 组装结果
	var responseList []question.Question
	addedMap := make(map[uint]bool) // 去重

	// 辅助统计信息
	var allRelatedQIDs []uint
	for _, q := range finalQuestions {
		allRelatedQIDs = append(allRelatedQIDs, q.ID)
		for _, child := range q.Children { allRelatedQIDs = append(allRelatedQIDs, child.ID) }
	}

	var records []answer.AnswerRecord
	db.DB.Where("user_id = ? AND question_id IN ?", userID, allRelatedQIDs).Find(&records)
	recordMap := make(map[uint]answer.AnswerRecord)
	for _, r := range records { recordMap[r.QuestionID] = r }

	type CountResult struct { QuestionID uint; Total int64 }
	var counts []CountResult
	db.DB.Table("notes").Select("question_id, count(1) as total").Where("question_id IN ?", allRelatedQIDs).Group("question_id").Scan(&counts)
	countMap := make(map[uint]int64)
	for _, c := range counts { countMap[c.QuestionID] = c.Total }

	// 按分页顺序组装
	for _, rawQ := range rawQuestions {
		targetID := rawQ.ID
		if pid, ok := parentIDMap[rawQ.ID]; ok {
			targetID = pid
		}

		if addedMap[targetID] { continue }

		if fullQ, exists := qMap[targetID]; exists {
			// 填充大题信息
			if rec, ok := recordMap[fullQ.ID]; ok { fullQ.UserRecord = rec }
			fullQ.NoteCount = countMap[fullQ.ID]

			// 填充子题信息
			for i := range fullQ.Children {
				child := &fullQ.Children[i]
				if rec, ok := recordMap[child.ID]; ok { child.UserRecord = rec }
				child.NoteCount = countMap[child.ID]
				// 累加子题笔记数到父题展示
				fullQ.NoteCount += child.NoteCount
			}

			responseList = append(responseList, fullQ)
			addedMap[targetID] = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responseList,
		"total": total,
		"page":  page,
	})
}

// ==========================================
// 4. 获取笔记目录树 (保持原样)
// ==========================================
func (h *Handler) GetNoteTree(c *gin.Context) {
	userID, _ := c.Get("userID")
	parentIDStr := c.Query("parent_id")
	source := c.Query("source")
	const MaxLevel = 5 

	query := db.DB.Model(&question.Category{})

	if parentIDStr == "" || parentIDStr == "0" {
		query = query.Where("parent_id IS NULL")
		if source != "" { query = query.Where("source = ?", source) }
	} else {
		query = query.Where("parent_id = ?", parentIDStr)
	}

	query = query.Where("level <= ?", MaxLevel)

	var currentCats []question.Category
	if err := query.Order("sort_order asc").Order("id asc").Find(&currentCats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载目录失败"})
		return
	}

	var finalCats []question.Category

	for _, cat := range currentCats {
		var noteCount int64
		hasMyNote := false

		db.DB.Table("notes").
			Joins("JOIN questions ON notes.question_id = questions.id").
			Where("notes.user_id = ?", userID).
			Where("questions.category_path LIKE ?", cat.FullPath+"%").
			Where("questions.deleted_at IS NULL").
			Select("count(1)").Limit(1).Find(&noteCount)
		
		if noteCount > 0 { hasMyNote = true }

		if !hasMyNote {
			noteCount = 0
			db.DB.Table("note_collects").
				Joins("JOIN notes ON note_collects.note_id = notes.id").
				Joins("JOIN questions ON notes.question_id = questions.id").
				Where("note_collects.user_id = ?", userID).
				Where("questions.category_path LIKE ?", cat.FullPath+"%").
				Where("questions.deleted_at IS NULL").
				Select("count(1)").Limit(1).Find(&noteCount)
			if noteCount > 0 { hasMyNote = true }
		}

		if !hasMyNote { continue }

		isLeaf := false
		if cat.Level >= MaxLevel {
			isLeaf = true
		} else {
			isLeaf = true 
			var subCats []question.Category
			db.DB.Where("parent_id = ?", cat.ID).Where("level <= ?", MaxLevel).Find(&subCats)

			for _, sub := range subCats {
				var subNoteCount int64
				db.DB.Table("notes").
					Joins("JOIN questions ON notes.question_id = questions.id").
					Where("notes.user_id = ?", userID).
					Where("questions.category_path LIKE ?", sub.FullPath+"%").
					Where("questions.deleted_at IS NULL").
					Select("count(1)").Limit(1).Find(&subNoteCount)
				
				if subNoteCount > 0 { isLeaf = false; break }

				db.DB.Table("note_collects").
					Joins("JOIN notes ON note_collects.note_id = notes.id").
					Joins("JOIN questions ON notes.question_id = questions.id").
					Where("note_collects.user_id = ?", userID).
					Where("questions.category_path LIKE ?", sub.FullPath+"%").
					Where("questions.deleted_at IS NULL").
					Select("count(1)").Limit(1).Find(&subNoteCount)
				
				if subNoteCount > 0 { isLeaf = false; break }
			}
		}

		cat.IsLeaf = isLeaf
		finalCats = append(finalCats, cat)
	}

	c.JSON(http.StatusOK, gin.H{"data": finalCats})
}

// ... DeleteNote, ToggleLike, ToggleCollect 保持不变 ...
func (h *Handler) DeleteNote(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	noteID := c.Param("id")
	var n Note
	if err := db.DB.Unscoped().First(&n, noteID).Error; err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "笔记不存在"}); return }
	if n.UserID != userID.(uint) && role != "admin" { c.JSON(http.StatusForbidden, gin.H{"error": "无权删除"}); return }
	if err := db.DB.Unscoped().Delete(&n).Error; err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *Handler) ToggleLike(c *gin.Context) {
	noteID := c.Param("id"); userID := c.MustGet("userID").(uint)
	var note Note; if err := db.DB.First(&note, noteID).Error; err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "笔记不存在"}); return }
	startOfDay := getStartOfDay()
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var like NoteLike
		result := tx.Where("user_id = ? AND note_id = ? AND created_at >= ?", userID, noteID, startOfDay).First(&like)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&like).Error; err != nil { return err }
			if err := tx.Model(&note).UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error; err != nil { return err }
			note.IsLiked = false; note.LikeCount--
		} else {
			newLike := NoteLike{UserID: userID, NoteID: note.ID}
			if err := tx.Create(&newLike).Error; err != nil { return err }
			if err := tx.Model(&note).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; err != nil { return err }
			note.IsLiked = true; note.LikeCount++
		}
		return nil
	})
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "操作成功", "is_liked": note.IsLiked, "like_count": note.LikeCount})
}

func (h *Handler) ToggleCollect(c *gin.Context) {
	noteIDStr := c.Param("id"); userID := c.MustGet("userID").(uint)
	noteID, _ := strconv.Atoi(noteIDStr); if noteID == 0 { c.JSON(http.StatusBadRequest, gin.H{"error": "无效ID"}); return }
	var collect NoteCollect
	result := db.DB.Where("user_id = ? AND note_id = ?", userID, noteID).First(&collect)
	isCollected := false
	if result.RowsAffected > 0 { db.DB.Delete(&collect); isCollected = false } else { newCollect := NoteCollect{UserID: userID, NoteID: uint(noteID)}; db.DB.Create(&newCollect); isCollected = true }
	c.JSON(http.StatusOK, gin.H{"message": "操作成功", "is_collected": isCollected})
}