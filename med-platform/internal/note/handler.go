package note

import (
	"med-platform/internal/common/db"
	"med-platform/internal/common/service"
	"med-platform/internal/common/uploader"
	"med-platform/internal/question"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
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
// 0. 图片上传 (存入临时池)
// ==========================================
func (h *Handler) UploadImage(c *gin.Context) {
	url, err := uploader.SaveImageWithHash(c, "file", uploader.MaxNoteImageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// ==========================================
// 1. 保存或更新笔记 (修复了修改频率和查重的误杀)
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

	// 🔥 核心修复点：防刷屏和防重复代码，只在“新建”时生效 (req.ID == 0)
	if req.ID == 0 {
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
	}

	finalImages := uploader.ConfirmImages(req.Images, "notes")

	var n Note
	if req.ID > 0 {
		// ================= 执行更新逻辑 =================
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
		// ================= 执行新建逻辑 =================
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

		if !req.IsPublic {
			db.DB.Model(&n).UpdateColumn("is_public", false)
		}

		if req.ParentID != nil && *req.ParentID > 0 {
			go func() {
				var parentNote Note
				if err := db.DB.Select("user_id").First(&parentNote, *req.ParentID).Error; err == nil {
					var q question.Question
					db.DB.Select("stem").First(&q, req.QuestionID)

					title := "题目讨论"
					stemClean := q.Stem
					if len([]rune(stemClean)) > 15 {
						title = "题目：" + string([]rune(stemClean)[:15]) + "..."
					} else if stemClean != "" {
						title = "题目：" + stemClean
					}

					service.SendNotification(parentNote.UserID, userID, "question", req.QuestionID, req.Content, title)
				}
			}()
		}
	}

	db.DB.Preload("User").Preload("Parent.User").First(&n, n.ID)
	c.JSON(http.StatusOK, gin.H{"message": "操作成功", "data": n})
}

// ==========================================
// 2. 获取某题的评论区 (保持不变)
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
		Preload("User").Preload("Parent").Preload("Parent.User").
		Where("question_id = ?", qID).
		Where("is_public = ? OR user_id = ?", true, userID)

	query.Count(&total)

	orderClause := "CASE WHEN user_id = " + strconv.Itoa(int(userID)) + " THEN 1 ELSE 0 END DESC, "
	if sortMode == "time" {
		orderClause += "created_at DESC"
	} else {
		orderClause += "like_count DESC, created_at DESC"
	}

	if err := query.Order(orderClause).Limit(pageSize).Offset(offset).Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	h.attachDynamicStatus(userID, notes)

	c.JSON(http.StatusOK, gin.H{
		"data":      notes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"has_more":  total > int64(page*pageSize),
	})
}

// ==========================================
// 3. 🔥🔥🔥 大厂级重构：获取我的笔记本 Feed 流
// ==========================================
func (h *Handler) GetMyNotes(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	// 核心参数：分辨是“我发布的(published)”还是“我收藏的(collected)”
	tab := c.DefaultQuery("tab", "published")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	categoryIDStr := c.Query("category_id")
	source := c.Query("source")

	var total int64
	var notes []Note

	// 1. 构建主查询 (以 Note 为主体)
	query := db.DB.Model(&Note{})

	if tab == "collected" {
		query = query.Joins("JOIN note_collects ON note_collects.note_id = notes.id").
			Where("note_collects.user_id = ?", userID)
	} else {
		query = query.Where("notes.user_id = ?", userID)
	}

	// 2. 根据目录和题库过滤 (连表查询)
	if source != "" || (categoryIDStr != "" && categoryIDStr != "0") {
		query = query.Joins("JOIN questions ON notes.question_id = questions.id")
		if source != "" {
			query = query.Where("questions.source = ?", source)
		}
		if categoryIDStr != "" && categoryIDStr != "0" {
			var cat question.Category
			if err := db.DB.First(&cat, categoryIDStr).Error; err == nil {
				query = query.Where("questions.category_path LIKE ?", cat.FullPath+"%")
			}
		}
	}

	query.Count(&total)

	// 3. 预加载关联数据（只取必要字段，极大减轻带宽）
	err := query.Order("notes.created_at desc").
		Limit(pageSize).Offset(offset).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, nickname, avatar, role")
		}).
		Preload("Question", func(db *gorm.DB) *gorm.DB {
			// 🔥 轻量级快照：前端只需要题目ID、类型和题干前缀，不需要选项和长篇大论的解析
			return db.Select("id, type, stem, parent_id")
		}).
		Preload("Parent").
		Preload("Parent.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, nickname")
		}).
		Find(&notes).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取笔记失败"})
		return
	}

	// 4. 挂载点赞与收藏状态
	h.attachDynamicStatus(userID, notes)

	c.JSON(http.StatusOK, gin.H{
		"data":  notes,
		"total": total,
		"page":  page,
	})
}

// GetNoteSkeleton 获取笔记题目骨架
func (h *Handler) GetNoteSkeleton(c *gin.Context) {
	userId, _ := c.Get("userID")
	userID := userId.(uint)
	source := c.Query("source")
	category := c.Query("category")

	groupExpr := "CASE WHEN questions.parent_id IS NOT NULL AND questions.parent_id > 0 THEN questions.parent_id ELSE questions.id END"

	baseQuery := db.DB.Table("notes").
		Select(groupExpr+" as id, MAX(questions.type) as type").
		Joins("JOIN questions ON notes.question_id = questions.id").
		Where("notes.user_id = ?", userID).
		Where("questions.deleted_at IS NULL").
		Group(groupExpr)

	if source != "" {
		baseQuery = baseQuery.Where("questions.source = ?", source)
	}
	if category != "" {
		baseQuery = baseQuery.Where("questions.category_path LIKE ?", category+"%")
	}

	type SkeletonItem struct {
		ID   uint   `json:"id"`
		Type string `json:"type"`
	}
	var items []SkeletonItem

	if err := baseQuery.Order("MAX(notes.updated_at) desc").Scan(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取笔记骨架失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(items),
		"data":  items,
	})
}

// ==========================================
// 4. 获取笔记目录树 (修复版：精准兼容父子题与分类)
// ==========================================
func (h *Handler) GetNoteTree(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	parentIDStr := c.Query("parent_id")
	if parentIDStr == "" {
		parentIDStr = c.Query("parent_key")
	}
	source := c.Query("source")
	tab := c.DefaultQuery("tab", "published")
	const MaxLevel = 5

	// 1. 获取当前层级的目录
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
		c.JSON(http.StatusOK, gin.H{"data": make([]interface{}, 0)}) // 防止返回 null
		return
	}

	// 2. 初始化结果数组 (必须用 make，避免前端接收到 null)
	result := make([]map[string]interface{}, 0)

	// 3. 循环当前层级分类，使用 category_path 进行精准匹配统计
	for _, cat := range currentCats {
		var count int64

		// 核心统计：使用 category_path LIKE 完美兼容父子题，且通过 DISTINCT 去重
		countQuery := db.DB.Table("notes").
			Select("COUNT(DISTINCT CASE WHEN questions.parent_id IS NOT NULL AND questions.parent_id > 0 THEN questions.parent_id ELSE questions.id END)").
			Joins("JOIN questions ON notes.question_id = questions.id").
			Where("questions.category_path LIKE ?", cat.FullPath+"%").
			Where("questions.deleted_at IS NULL")

		if tab == "collected" {
			countQuery = countQuery.Joins("JOIN note_collects ON note_collects.note_id = notes.id").
				Where("note_collects.user_id = ?", userID)
		} else {
			countQuery = countQuery.Where("notes.user_id = ?", userID)
		}

		countQuery.Scan(&count)

		if count == 0 {
			continue
		}

		isLeaf := false
		if cat.Level >= MaxLevel {
			isLeaf = true
		} else {
			// 简单判断是否有下级子分类
			var subCount int64
			db.DB.Model(&question.Category{}).Where("parent_id = ? AND level <= ?", cat.ID, MaxLevel).Count(&subCount)
			isLeaf = (subCount == 0)
		}

		result = append(result, map[string]interface{}{
			"id":     cat.ID,
			"label":  cat.Name + " (" + strconv.Itoa(int(count)) + ")",
			"name":   cat.Name,
			"full":   cat.FullPath,
			"isLeaf": isLeaf,
			"count":  count,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// 附加状态：批量查询点赞与收藏状态
func (h *Handler) attachDynamicStatus(userID uint, notes []Note) {
	if len(notes) == 0 {
		return
	}
	var noteIDs []uint
	for _, n := range notes {
		noteIDs = append(noteIDs, n.ID)
	}

	startOfDay := getStartOfDay()
	var likedNoteIDs []uint
	db.DB.Model(&NoteLike{}).Where("user_id = ? AND note_id IN ? AND created_at >= ?", userID, noteIDs, startOfDay).Pluck("note_id", &likedNoteIDs)
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

// ==========================================
// 5. 以下为通用操作 (增删改查、点赞等，保持不变)
// ==========================================
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
		newReport := NoteReport{UserID: userID, NoteID: uint(noteID), Reason: req.Reason}
		if err := tx.Create(&newReport).Error; err != nil {
			return err
		}
		if err := tx.Model(&Note{}).Where("id = ?", noteID).
			Updates(map[string]interface{}{"is_reported": true, "report_count": gorm.Expr("report_count + 1")}).Error; err != nil {
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

func (h *Handler) AdminListNotes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	keyword := c.Query("keyword")
	userID := c.Query("user_id")
	questionID := c.Query("question_id")
	onlyReported := c.Query("reported")

	query := db.DB.Model(&Note{}).Preload("User").Preload("Question").Preload("Reports").Order("created_at desc")

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

	c.JSON(http.StatusOK, gin.H{"data": notes, "total": total, "page": page})
}

func (h *Handler) AdminDismissReport(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, _ := strconv.Atoi(noteIDStr)

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("note_id = ?", noteID).Delete(&NoteReport{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&Note{}).Where("id = ?", noteID).Updates(map[string]interface{}{"is_reported": false, "report_count": 0}).Error; err != nil {
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
