package note

import (
	"med-platform/internal/answer"
	"med-platform/internal/common/db"
	"med-platform/internal/common/logger"
	"med-platform/internal/common/service" // 🔥 1. 引入通用通知服务
	"med-platform/internal/common/uploader"
	"med-platform/internal/question"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// 辅助函数：获取当天的 0 点时间
func getStartOfDay() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// ==========================================
// 🔥 0. 图片上传 (存入临时池)
// ==========================================
func (h *Handler) UploadImage(c *gin.Context) {
	// 存入 temp 目录，文件名哈希化
	url, err := uploader.SaveImageWithHash(c, "file", uploader.MaxNoteImageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// ==========================================
// 1. 保存或更新笔记 (升级版：固化图片 + 防御 + 🔥全局通知)
// ==========================================
func (h *Handler) SaveNote(c *gin.Context) {
	var req struct {
		ID         uint     `json:"id"`
		QuestionID uint     `json:"question_id"`
		Content    string   `json:"content"`
		IsPublic   bool     `json:"is_public"`
		ParentID   *uint    `json:"parent_id"`
		Images     []string `json:"images"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 强力校验
	if utf8.RuneCountInString(req.Content) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "字数不能超过200字"})
		return
	}
	if len(req.Images) > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "最多只能上传5张图片"})
		return
	}
	if req.Content == "" && len(req.Images) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "内容不能为空"})
		return
	}

	userID := c.MustGet("userID").(uint)

	// 防御代码
	var lastNote Note
	result := db.DB.Where("user_id = ?", userID).Order("created_at desc").First(&lastNote)

	if result.Error == nil {
		timePassed := time.Since(lastNote.CreatedAt)
		if timePassed < 10*time.Second {
			timeLeft := 10 - int(timePassed.Seconds())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "说话太快啦，请休息 " + strconv.Itoa(timeLeft) + " 秒后再发",
			})
			return
		}
		if req.Content != "" && req.Content == lastNote.Content {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请勿重复发送相同内容"})
			return
		}
	}

	// 图片固化
	finalImages := uploader.ConfirmImages(req.Images, "notes")

	var n Note
	if req.ID > 0 {
		// 更新逻辑
		if err := db.DB.First(&n, req.ID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到原笔记"})
			return
		}
		if n.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权修改"})
			return
		}
		n.Content = req.Content
		n.IsPublic = req.IsPublic
		n.Images = finalImages
		db.DB.Save(&n)
	} else {
		// 新增逻辑
		n = Note{
			UserID:     userID,
			QuestionID: req.QuestionID,
			Content:    req.Content,
			IsPublic:   req.IsPublic,
			ParentID:   req.ParentID,
			Images:     finalImages,
		}
		if err := db.DB.Create(&n).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
			return
		}

		// 🔥🔥🔥 核心升级：发送通知 (如果是回复他人) 🔥🔥🔥
		if req.ParentID != nil && *req.ParentID > 0 {
			// 异步发送，不阻塞主流程
			go func() {
				var parentNote Note
				// 1. 查出父笔记作者
				if err := db.DB.Select("user_id").First(&parentNote, *req.ParentID).Error; err == nil {
					// 2. 查出题目信息（用于通知标题）
					var q question.Question
					db.DB.Select("stem").First(&q, req.QuestionID)

					// 3. 构建标题（截取部分题干）
					title := "题目讨论"
					// 简单的去HTML标签处理（虽然SendNotification里也会截断，这里做个源头处理更好）
					stemClean := q.Stem // 这里假设Stem可能含HTML，简单展示即可
					if len([]rune(stemClean)) > 15 {
						title = "题目：" + string([]rune(stemClean)[:15]) + "..."
					} else if stemClean != "" {
						title = "题目：" + stemClean
					}

					// 4. 发送通知
					// SourceType: "question" -> 前端跳转到做题页
					// SourceID: req.QuestionID -> 题目ID
					service.SendNotification(
						parentNote.UserID,
						userID,
						"question",
						req.QuestionID,
						req.Content,
						title,
					)
				}
			}()
		}
	}

	db.DB.Preload("User").Preload("Parent.User").First(&n, n.ID)
	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "data": n})
}

// ==========================================
// 2. 获取某题的评论区 (分页 + 排序)
// ==========================================
func (h *Handler) ListNotes(c *gin.Context) {
	qID := c.Query("question_id")
	if qID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少题目ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "5"))
	if pageSize > 20 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	sortMode := c.DefaultQuery("sort", "hot")

	userID := c.MustGet("userID").(uint)
	var notes []Note
	var total int64

	query := db.DB.Model(&Note{}).
		Preload("User").
		Preload("Parent").
		Preload("Parent.User").
		Where("question_id = ?", qID).
		Where("is_public = ? OR user_id = ?", true, userID)

	query.Count(&total)

	orderClause := "CASE WHEN user_id = " + strconv.Itoa(int(userID)) + " THEN 1 ELSE 0 END DESC, "
	if sortMode == "time" {
		orderClause += "created_at DESC"
	} else {
		orderClause += "like_count DESC, created_at DESC"
	}

	if err := query.Order(orderClause).
		Limit(pageSize).Offset(offset).
		Find(&notes).Error; err != nil {
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
		for _, id := range likedNoteIDs {
			likedMap[id] = true
		}

		var collectedNoteIDs []uint
		db.DB.Model(&NoteCollect{}).Where("user_id = ? AND note_id IN ?", userID, noteIDs).Pluck("note_id", &collectedNoteIDs)
		collectedMap := make(map[uint]bool)
		for _, id := range collectedNoteIDs {
			collectedMap[id] = true
		}

		for i := range notes {
			notes[i].IsLiked = likedMap[notes[i].ID]
			notes[i].IsCollected = collectedMap[notes[i].ID]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      notes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"has_more":  total > int64(page*pageSize),
	})
}

// ==========================================
// 3. 获取我的笔记本列表
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

	var myNoteQIDs []uint
	db.DB.Model(&Note{}).Where("user_id = ?", userID).Distinct("question_id").Pluck("question_id", &myNoteQIDs)

	var collectedNoteQIDs []uint
	db.DB.Table("note_collects").
		Joins("JOIN notes ON note_collects.note_id = notes.id").
		Where("note_collects.user_id = ?", userID).
		Distinct("notes.question_id").
		Pluck("notes.question_id", &collectedNoteQIDs)

	idMap := make(map[uint]bool)
	for _, id := range myNoteQIDs {
		idMap[id] = true
	}
	for _, id := range collectedNoteQIDs {
		idMap[id] = true
	}

	var rawQIDs []uint
	for id := range idMap {
		rawQIDs = append(rawQIDs, id)
	}

	if len(rawQIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0, "page": page})
		return
	}

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

	err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&rawQuestions).Error
	if err != nil {
		logger.Log.Error("查询笔记本题目失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

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

	var finalQuestions []question.Question
	db.DB.Preload("Children", func(db *gorm.DB) *gorm.DB { return db.Order("id asc") }).
		Where("id IN ?", finalQIDs).
		Find(&finalQuestions)

	qMap := make(map[uint]question.Question)
	for _, q := range finalQuestions {
		qMap[q.ID] = q
	}

	var responseList []question.Question
	addedMap := make(map[uint]bool)

	var allRelatedQIDs []uint
	for _, q := range finalQuestions {
		allRelatedQIDs = append(allRelatedQIDs, q.ID)
		for _, child := range q.Children {
			allRelatedQIDs = append(allRelatedQIDs, child.ID)
		}
	}

	var records []answer.AnswerRecord
	db.DB.Where("user_id = ? AND question_id IN ?", userID, allRelatedQIDs).Find(&records)
	recordMap := make(map[uint]answer.AnswerRecord)
	for _, r := range records {
		recordMap[r.QuestionID] = r
	}

	type CountResult struct {
		QuestionID uint
		Total      int64
	}
	var counts []CountResult
	db.DB.Table("notes").Select("question_id, count(1) as total").Where("question_id IN ?", allRelatedQIDs).Group("question_id").Scan(&counts)
	countMap := make(map[uint]int64)
	for _, c := range counts {
		countMap[c.QuestionID] = c.Total
	}

	for _, rawQ := range rawQuestions {
		targetID := rawQ.ID
		if pid, ok := parentIDMap[rawQ.ID]; ok {
			targetID = pid
		}

		if addedMap[targetID] {
			continue
		}

		if fullQ, exists := qMap[targetID]; exists {
			if rec, ok := recordMap[fullQ.ID]; ok {
				fullQ.UserRecord = rec
			}
			fullQ.NoteCount = countMap[fullQ.ID]

			for i := range fullQ.Children {
				child := &fullQ.Children[i]
				if rec, ok := recordMap[child.ID]; ok {
					child.UserRecord = rec
				}
				child.NoteCount = countMap[child.ID]
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
// 4. 获取笔记目录树
// ==========================================
func (h *Handler) GetNoteTree(c *gin.Context) {
	userID, _ := c.Get("userID")
	parentIDStr := c.Query("parent_id")
	source := c.Query("source")
	const MaxLevel = 5

	query := db.DB.Model(&question.Category{})

	if parentIDStr == "" || parentIDStr == "0" {
		query = query.Where("parent_id IS NULL")
		if source != "" {
			query = query.Where("source = ?", source)
		}
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

		if noteCount > 0 {
			hasMyNote = true
		}

		if !hasMyNote {
			noteCount = 0
			db.DB.Table("note_collects").
				Joins("JOIN notes ON note_collects.note_id = notes.id").
				Joins("JOIN questions ON notes.question_id = questions.id").
				Where("note_collects.user_id = ?", userID).
				Where("questions.category_path LIKE ?", cat.FullPath+"%").
				Where("questions.deleted_at IS NULL").
				Select("count(1)").Limit(1).Find(&noteCount)
			if noteCount > 0 {
				hasMyNote = true
			}
		}

		if !hasMyNote {
			continue
		}

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

				if subNoteCount > 0 {
					isLeaf = false
					break
				}

				db.DB.Table("note_collects").
					Joins("JOIN notes ON note_collects.note_id = notes.id").
					Joins("JOIN questions ON notes.question_id = questions.id").
					Where("note_collects.user_id = ?", userID).
					Where("questions.category_path LIKE ?", sub.FullPath+"%").
					Where("questions.deleted_at IS NULL").
					Select("count(1)").Limit(1).Find(&subNoteCount)

				if subNoteCount > 0 {
					isLeaf = false
					break
				}
			}
		}

		cat.IsLeaf = isLeaf
		finalCats = append(finalCats, cat)
	}

	c.JSON(http.StatusOK, gin.H{"data": finalCats})
}

// DeleteNote 删除笔记
func (h *Handler) DeleteNote(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	noteID := c.Param("id")

	var n Note
	if err := db.DB.Unscoped().First(&n, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "笔记不存在"})
		return
	}
	if n.UserID != userID.(uint) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除"})
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("note_id = ?", n.ID).Delete(&NoteReport{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&n).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ToggleLike 点赞/取消点赞
func (h *Handler) ToggleLike(c *gin.Context) {
	noteID := c.Param("id")
	userID := c.MustGet("userID").(uint)
	var note Note
	if err := db.DB.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "笔记不存在"})
		return
	}
	startOfDay := getStartOfDay()
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var like NoteLike
		result := tx.Where("user_id = ? AND note_id = ? AND created_at >= ?", userID, noteID, startOfDay).First(&like)
		if result.RowsAffected > 0 {
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			if err := tx.Model(&note).UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
				return err
			}
			note.IsLiked = false
			note.LikeCount--
		} else {
			newLike := NoteLike{UserID: userID, NoteID: note.ID}
			if err := tx.Create(&newLike).Error; err != nil {
				return err
			}
			if err := tx.Model(&note).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
				return err
			}
			note.IsLiked = true
			note.LikeCount++
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "操作成功", "is_liked": note.IsLiked, "like_count": note.LikeCount})
}

// ToggleCollect 收藏/取消收藏
func (h *Handler) ToggleCollect(c *gin.Context) {
	noteIDStr := c.Param("id")
	userID := c.MustGet("userID").(uint)
	noteID, _ := strconv.Atoi(noteIDStr)
	if noteID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效ID"})
		return
	}
	var collect NoteCollect
	result := db.DB.Where("user_id = ? AND note_id = ?", userID, noteID).First(&collect)
	isCollected := false
	if result.RowsAffected > 0 {
		db.DB.Delete(&collect)
		isCollected = false
	} else {
		newCollect := NoteCollect{UserID: userID, NoteID: uint(noteID)}
		db.DB.Create(&newCollect)
		isCollected = true
	}
	c.JSON(http.StatusOK, gin.H{"message": "操作成功", "is_collected": isCollected})
}

// ==========================================
// 🔥 5. 举报系统
// ==========================================

// ReportNote 用户举报笔记
func (h *Handler) ReportNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, _ := strconv.Atoi(noteIDStr)
	userID := c.MustGet("userID").(uint)

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择或填写举报理由"})
		return
	}

	var existingReport NoteReport
	err := db.DB.Where("user_id = ? AND note_id = ?", userID, noteID).First(&existingReport).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "您已经举报过该内容，请勿重复提交"})
		return
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		newReport := NoteReport{
			UserID: userID,
			NoteID: uint(noteID),
			Reason: req.Reason,
		}
		if err := tx.Create(&newReport).Error; err != nil {
			return err
		}
		if err := tx.Model(&Note{}).Where("id = ?", noteID).
			Updates(map[string]interface{}{
				"is_reported":  true,
				"report_count": gorm.Expr("report_count + 1"),
			}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "举报失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "举报成功，感谢您的监督"})
}

// AdminListNotes 管理员获取笔记列表
func (h *Handler) AdminListNotes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	keyword := c.Query("keyword")
	userID := c.Query("user_id")
	questionID := c.Query("question_id")
	onlyReported := c.Query("reported")

	query := db.DB.Model(&Note{}).
		Preload("User").
		Preload("Question").
		Preload("Reports").
		Order("created_at desc")

	if onlyReported == "true" {
		query = query.Where("is_reported = ?", true).Order("report_count desc")
	}

	if keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if questionID != "" {
		query = query.Where("question_id = ?", questionID)
	}

	var total int64
	query.Count(&total)

	var notes []Note
	if err := query.Offset(offset).Limit(pageSize).Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  notes,
		"total": total,
		"page":  page,
	})
}

// AdminDismissReport 管理员忽略举报
func (h *Handler) AdminDismissReport(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, _ := strconv.Atoi(noteIDStr)

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("note_id = ?", noteID).Delete(&NoteReport{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&Note{}).Where("id = ?", noteID).
			Updates(map[string]interface{}{
				"is_reported":  false,
				"report_count": 0,
			}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已忽略该举报，相关记录已清空"})
}