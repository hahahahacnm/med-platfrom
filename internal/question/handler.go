package question

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"med-platform/internal/common/db"
	"med-platform/internal/product"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Handler struct {
	repo *Repository
}

func NewHandler() *Handler {
	return &Handler{
		repo: NewRepository(),
	}
}

// =========================================================
// 🖼️ [新增] 外链图片转存辅助函数
// =========================================================

var (
	mdImgRegex     = regexp.MustCompile(`!\[(.*?)\]\((https?://[^\)]+)\)`)
	htmlImgRegex   = regexp.MustCompile(`src=["'](https?://[^"']+)["']`)
	customImgRegex = regexp.MustCompile(`\[图片:(https?://[^\]]+)\]`)
)

func processContentImages(content string) string {
	if content == "" {
		return ""
	}

	content = mdImgRegex.ReplaceAllStringFunc(content, func(s string) string {
		matches := mdImgRegex.FindStringSubmatch(s)
		if len(matches) < 3 {
			return s
		}
		localURL, err := downloadAndSaveImage(matches[2])
		if err != nil {
			return s
		}
		return fmt.Sprintf("![%s](%s)", matches[1], localURL)
	})

	content = htmlImgRegex.ReplaceAllStringFunc(content, func(s string) string {
		matches := htmlImgRegex.FindStringSubmatch(s)
		if len(matches) < 2 {
			return s
		}
		localURL, err := downloadAndSaveImage(matches[1])
		if err != nil {
			return s
		}
		return strings.Replace(s, matches[1], localURL, 1)
	})

	content = customImgRegex.ReplaceAllStringFunc(content, func(s string) string {
		matches := customImgRegex.FindStringSubmatch(s)
		if len(matches) < 2 {
			return s
		}
		localURL, err := downloadAndSaveImage(matches[1])
		if err != nil {
			return s
		}
		return fmt.Sprintf("![图片](%s)", localURL)
	})

	return content
}

func downloadAndSaveImage(remoteURL string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(remoteURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code %d", resp.StatusCode)
	}

	saveDir := "./uploads/questions"
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		_ = os.MkdirAll(saveDir, 0755)
	}

	ext := filepath.Ext(remoteURL)
	if ext == "" {
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "png") {
			ext = ".png"
		} else if strings.Contains(contentType, "gif") {
			ext = ".gif"
		} else {
			ext = ".jpg"
		}
	}
	if idx := strings.Index(ext, "?"); idx != -1 {
		ext = ext[:idx]
	}

	fileName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String(), ext)
	localPath := filepath.Join(saveDir, fileName)

	out, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return "/uploads/questions/" + fileName, nil
}

// =========================================================
// 🔐 核心鉴权辅助函数
// =========================================================
func checkAccess(c *gin.Context, source string, categoryPath string) bool {
	uidRaw, exists := c.Get("userID")
	if !exists {
		return false
	}
	userID := uidRaw.(uint)

	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)

	if role == "admin" || role == "agent" {
		return true
	}

	normalizedPath := strings.ReplaceAll(categoryPath, " > ", "/")
	parts := strings.Split(normalizedPath, "/")
	rootCategory := ""
	if len(parts) > 0 {
		rootCategory = strings.TrimSpace(parts[0])
	}

	return product.NewRepository().CheckPermission(userID, source, rootCategory)
}

func hardDeleteQuestions(tx *gorm.DB, questionIDs []uint) error {
	if len(questionIDs) == 0 {
		return nil
	}

	var childIDs []uint
	tx.Model(&Question{}).Where("parent_id IN ?", questionIDs).Pluck("id", &childIDs)
	allIDs := append(questionIDs, childIDs...)

	if err := tx.Exec("DELETE FROM user_favorites WHERE question_id IN ?", allIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM user_mistakes WHERE question_id IN ?", allIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM answer_records WHERE question_id IN ?", allIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM question_feedbacks WHERE question_id IN ?", allIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM note_likes WHERE note_id IN (SELECT id FROM notes WHERE question_id IN ?)", allIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM note_collects WHERE note_id IN (SELECT id FROM notes WHERE question_id IN ?)", allIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM notes WHERE question_id IN ?", allIDs).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("id IN ?", allIDs).Delete(&Question{}).Error; err != nil {
		return err
	}
	return nil
}

func cleanStem(text string) string {
	text = strings.ReplaceAll(text, "【共用主干】", "")
	text = strings.ReplaceAll(text, "【共用题干】", "")
	text = strings.ReplaceAll(text, "【案例描述】", "")
	return strings.TrimSpace(text)
}

func cleanStemForFingerprint(text string) string {
	text = cleanStem(text)
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "　", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.ReplaceAll(text, "，", "")
	return text
}

func getTypeWeight(t string) int {
	t = strings.ToUpper(t)
	if strings.Contains(t, "A1") {
		return 10
	}
	if strings.Contains(t, "A2") {
		return 20
	}
	if strings.Contains(t, "A3") {
		return 30
	}
	if strings.Contains(t, "A4") {
		return 40
	}
	if strings.Contains(t, "B1") {
		return 50
	}
	if strings.Contains(t, "X") {
		return 60
	}
	return 999
}

// =================================================================
// 🔥🔥🔥 核心强化：获取章节答题卡骨架 (含实时科学学习统计) 🔥🔥🔥
// =================================================================
func (h *Handler) GetChapterSkeleton(c *gin.Context) {
	category := c.Query("category")
	source := c.Query("source")

	uidRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	var userID uint
	switch id := uidRaw.(type) {
	case uint:
		userID = id
	case int:
		userID = uint(id)
	case float64:
		userID = uint(id)
	}

	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请指定章节"})
		return
	}

	if source != "" {
		if !checkAccess(c, source, category) {
			c.JSON(http.StatusForbidden, gin.H{"error": "FORBIDDEN", "message": "🔒 您尚未获得该科目的访问授权"})
			return
		}
	}

	// 1. 获取本章所有的题目结构 (超轻量查询)
	type QLite struct {
		ID       uint
		ParentID *uint
		Type     string
	}
	var allQs []QLite
	query := db.DB.Model(&Question{}).
		Select("id, parent_id, type").
		Where("category_path LIKE ?", category+"%").
		Where("deleted_at IS NULL")

	if source != "" {
		query = query.Where("source = ?", source)
	}
	if err := query.Order("id asc").Find(&allQs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取答题卡数据失败"})
		return
	}

	if len(allQs) == 0 {
		c.JSON(http.StatusOK, gin.H{"total": 0, "data": []interface{}{}, "summary": nil})
		return
	}

	// 2. 获取该用户对这些题目的作答记录
	var qIDs []uint
	for _, q := range allQs {
		qIDs = append(qIDs, q.ID)
	}

	type RecordStatus struct {
		QuestionID uint
		IsCorrect  bool
	}
	var records []RecordStatus
	db.DB.Table("answer_records").
		Select("question_id, is_correct").
		Where("user_id = ? AND question_id IN ?", userID, qIDs).
		Scan(&records)

	recordMap := make(map[uint]bool) // qID -> isCorrect
	for _, r := range records {
		recordMap[r.QuestionID] = r.IsCorrect
	}

	// 3. 组装父级骨架并智能聚合子题状态
	var skeletons []map[string]interface{}
	childMap := make(map[uint][]uint) // parentID -> []childID

	for _, q := range allQs {
		if q.ParentID != nil && *q.ParentID > 0 {
			childMap[*q.ParentID] = append(childMap[*q.ParentID], q.ID)
		}
	}

	// 准备统计变量
	var correctNum, attemptedNum int

	for _, q := range allQs {
		// 只有父题或独立单题才能上答题卡
		if q.ParentID == nil || *q.ParentID == 0 {
			status := "unfilled"
			childrenIDs := childMap[q.ID]

			if len(childrenIDs) > 0 {
				// 组合大题：只要有子题错就算错，全对才算对
				answeredSubCount := 0
				wrongSubCount := 0
				for _, cID := range childrenIDs {
					if isCorrect, ok := recordMap[cID]; ok {
						answeredSubCount++
						if !isCorrect {
							wrongSubCount++
						}
					}
				}
				if answeredSubCount > 0 {
					if wrongSubCount > 0 {
						status = "wrong"
						attemptedNum++ // 只要做了就算“已尝试”
					} else if answeredSubCount == len(childrenIDs) {
						status = "correct"
						attemptedNum++
						correctNum++ // 全对才计入“正确数”
					} else {
						status = "partial" // 做了但没做完
						attemptedNum++     // 部分完成也算“已尝试”
					}
				}
			} else {
				// 普通单题
				if isCorrect, ok := recordMap[q.ID]; ok {
					attemptedNum++
					if isCorrect {
						status = "correct"
						correctNum++
					} else {
						status = "wrong"
					}
				}
			}

			skeletons = append(skeletons, map[string]interface{}{
				"id":     q.ID,
				"type":   q.Type,
				"status": status,
			})
		}
	}

	// 4. 计算实时科学学习数据
	totalNum := len(skeletons)
	accuracyRate := 0.0 // 正确率：正确题数 / 已做题数
	if attemptedNum > 0 {
		accuracyRate = float64(correctNum) / float64(attemptedNum) * 100
	}

	masteryRate := 0.0 // 掌握率：正确题数 / 总题数
	if totalNum > 0 {
		masteryRate = float64(correctNum) / float64(totalNum) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total": totalNum,
		"data":  skeletons,
		"summary": gin.H{
			"correct_num":   correctNum,                       // 已答对大题数
			"attempted_num": attemptedNum,                     // 已尝试大题数
			"total_num":     totalNum,                         // 本章总大题数
			"accuracy_rate": fmt.Sprintf("%.1f", accuracyRate), // 格式化为字符串，方便前端直接展示
			"mastery_rate":  fmt.Sprintf("%.1f", masteryRate),
		},
	})
}

// =================================================================
// 🔥🔥🔥 核心强化：单题详情接口 (含子题与个人统计) 🔥🔥🔥
// =================================================================
func (h *Handler) GetDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	uidRaw, _ := c.Get("userID")
	var userID uint
	switch uid := uidRaw.(type) {
	case uint:
		userID = uid
	case float64:
		userID = uint(uid)
	}

	// 1. 获取题目本身(包含预加载的 Children)
	q, err := h.repo.GetDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目未找到"})
		return
	}

	// 2. 鉴权
	if !checkAccess(c, q.Source, q.CategoryPath) {
		c.JSON(http.StatusForbidden, gin.H{"error": "FORBIDDEN", "message": "🔒 您无权查看该题目详情"})
		return
	}

	// 3. 聚合个人专属状态 (作答记录、收藏、笔记)
	var allQIDs []uint
	allQIDs = append(allQIDs, q.ID)
	for _, child := range q.Children {
		allQIDs = append(allQIDs, child.ID)
	}

	favMap := make(map[uint]bool)
	recordMap := make(map[uint]interface{})
	noteCountMap := make(map[uint]int64)

	if userID > 0 {
		var favs []uint
		db.DB.Table("user_favorites").Where("user_id = ? AND question_id IN ?", userID, allQIDs).Pluck("question_id", &favs)
		for _, qid := range favs {
			favMap[qid] = true
		}

		type RecordDTO struct {
			QuestionID uint
			Choice     string
			IsCorrect  bool
		}
		var records []RecordDTO
		db.DB.Table("answer_records").Select("question_id, choice, is_correct").Where("user_id = ? AND question_id IN ?", userID, allQIDs).Order("created_at asc").Scan(&records)
		for _, r := range records {
			if r.Choice != "" {
				recordMap[r.QuestionID] = map[string]interface{}{"choice": r.Choice, "is_correct": r.IsCorrect}
			}
		}
	}

	type CountResult struct {
		QuestionID uint
		Total      int64
	}
	var counts []CountResult
	db.DB.Table("notes").Select("question_id, count(1) as total").Where("question_id IN (?)", allQIDs).Group("question_id").Scan(&counts)
	for _, cnt := range counts {
		noteCountMap[cnt.QuestionID] = cnt.Total
	}

	// 4. 将基础数据和个人状态捏合在一起，返回完美的大 JSON
	currentTotalNotes := noteCountMap[q.ID]
	var optionsMap map[string]string
	if len(q.Options) > 0 {
		_ = json.Unmarshal(q.Options, &optionsMap)
	}

	var childrenList []map[string]interface{}
	if len(q.Children) > 0 {
		for _, child := range q.Children {
			var childOpts map[string]string
			if !strings.Contains(q.Type, "B1") && len(child.Options) > 0 {
				_ = json.Unmarshal(child.Options, &childOpts)
			}
			childNoteCount := noteCountMap[child.ID]
			currentTotalNotes += childNoteCount

			childrenList = append(childrenList, map[string]interface{}{
				"id":              child.ID,
				"type":            child.Type,
				"stem":            child.Stem,
				"options":         childOpts,
				"correct":         child.Correct,
				"analysis":        child.Analysis,
				"user_record":     recordMap[child.ID], // 小题的作答记录
				"difficulty":      child.Difficulty,
				"diff_value":      child.DiffValue,
				"syllabus":        child.Syllabus,
				"cognitive_level": child.CognitiveLevel,
				"note_count":      childNoteCount,
				"category_path":   child.CategoryPath,
			})
		}
	}

	item := map[string]interface{}{
		"id":              q.ID,
		"type":            q.Type,
		"stem":            q.Stem,
		"options":         optionsMap,
		"correct":         q.Correct,
		"analysis":        q.Analysis,
		"difficulty":      q.Difficulty,
		"diff_value":      q.DiffValue,
		"syllabus":        q.Syllabus,
		"cognitive_level": q.CognitiveLevel,
		"source":          q.Source,
		"is_favorite":     favMap[q.ID],          // 大题的收藏状态
		"user_record":     recordMap[q.ID],       // 单题时的作答记录
		"note_count":      currentTotalNotes,     // 包含子题的总笔记数
		"children":        childrenList,
		"category_path":   q.CategoryPath,
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

// =================================================================
// List 获取题目列表 (保留原状，供后台管理使用)
// =================================================================
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}

	category := c.Query("category")
	keyword := c.Query("q")
	source := c.Query("source")

	if source != "" && category != "" {
		if !checkAccess(c, source, category) {
			c.JSON(http.StatusForbidden, gin.H{"error": "FORBIDDEN", "message": "🔒 您尚未获得该科目的访问授权，请联系管理员或购买相关课程"})
			return
		}
	}

	rawQuestions, total, err := h.repo.List(page, pageSize, category, keyword, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var finalQuestions []*Question
	var parentIDsToFetch []uint
	processedParentMap := make(map[uint]bool)

	for _, q := range rawQuestions {
		if q.ParentID != nil && *q.ParentID > 0 {
			if !processedParentMap[*q.ParentID] {
				parentIDsToFetch = append(parentIDsToFetch, *q.ParentID)
				processedParentMap[*q.ParentID] = true
			}
		}
	}

	parentMap := make(map[uint]*Question)
	if len(parentIDsToFetch) > 0 {
		var parents []*Question
		db.DB.Preload("Children").Where("id IN ?", parentIDsToFetch).Find(&parents)
		for _, p := range parents {
			parentMap[p.ID] = p
		}
	}

	addedParentIDs := make(map[uint]bool)

	for i := range rawQuestions {
		q := &rawQuestions[i]

		if q.ParentID != nil && *q.ParentID > 0 {
			if parent, exists := parentMap[*q.ParentID]; exists {
				if !addedParentIDs[parent.ID] {
					finalQuestions = append(finalQuestions, parent)
					addedParentIDs[parent.ID] = true
				}
			} else {
				finalQuestions = append(finalQuestions, q)
			}
		} else {
			finalQuestions = append(finalQuestions, q)
		}
	}

	var allQIDs []uint
	for _, q := range finalQuestions {
		allQIDs = append(allQIDs, q.ID)
		for _, child := range q.Children {
			allQIDs = append(allQIDs, child.ID)
		}
	}

	var userID uint
	if v, exists := c.Get("userID"); exists {
		switch id := v.(type) {
		case uint:
			userID = id
		case float64:
			userID = uint(id)
		}
	}

	favMap := make(map[uint]bool)
	recordMap := make(map[uint]interface{})
	noteCountMap := make(map[uint]int64)

	if len(allQIDs) > 0 {
		if userID > 0 {
			var favs []uint
			db.DB.Table("user_favorites").Where("user_id = ? AND question_id IN ?", userID, allQIDs).Pluck("question_id", &favs)
			for _, qid := range favs {
				favMap[qid] = true
			}

			type RecordDTO struct {
				QuestionID uint
				Choice     string
				IsCorrect  bool
			}
			var records []RecordDTO
			db.DB.Table("answer_records").Select("question_id, choice, is_correct").Where("user_id = ? AND question_id IN ?", userID, allQIDs).Order("created_at asc").Scan(&records)
			for _, r := range records {
				if r.Choice != "" {
					recordMap[r.QuestionID] = map[string]interface{}{"choice": r.Choice, "is_correct": r.IsCorrect}
				}
			}
		}

		type CountResult struct {
			QuestionID uint
			Total      int64
		}
		var counts []CountResult
		db.DB.Table("notes").Select("question_id, count(1) as total").Where("question_id IN (?)", allQIDs).Group("question_id").Scan(&counts)
		for _, c := range counts {
			noteCountMap[c.QuestionID] = c.Total
		}
	}

	var responseList []map[string]interface{}
	for _, q := range finalQuestions {
		currentTotalNotes := noteCountMap[q.ID]
		var optionsMap map[string]string
		if len(q.Options) > 0 {
			_ = json.Unmarshal(q.Options, &optionsMap)
		}

		var childrenList []map[string]interface{}
		if len(q.Children) > 0 {
			for _, child := range q.Children {
				var childOpts map[string]string
				if !strings.Contains(q.Type, "B1") && len(child.Options) > 0 {
					_ = json.Unmarshal(child.Options, &childOpts)
				}
				childNoteCount := noteCountMap[child.ID]
				currentTotalNotes += childNoteCount

				childrenList = append(childrenList, map[string]interface{}{
					"id":              child.ID,
					"type":            child.Type,
					"stem":            child.Stem,
					"options":         childOpts,
					"correct":         child.Correct,
					"analysis":        child.Analysis,
					"user_record":     recordMap[child.ID],
					"difficulty":      child.Difficulty,
					"diff_value":      child.DiffValue,
					"syllabus":        child.Syllabus,
					"cognitive_level": child.CognitiveLevel,
					"note_count":      childNoteCount,
					"category_path":   child.CategoryPath,
				})
			}
		}

		item := map[string]interface{}{
			"id":              q.ID,
			"type":            q.Type,
			"stem":            q.Stem,
			"options":         optionsMap,
			"correct":         q.Correct,
			"analysis":        q.Analysis,
			"difficulty":      q.Difficulty,
			"diff_value":      q.DiffValue,
			"syllabus":        q.Syllabus,
			"cognitive_level": q.CognitiveLevel,
			"source":          q.Source,
			"is_favorite":     favMap[q.ID],
			"user_record":     recordMap[q.ID],
			"note_count":      currentTotalNotes,
			"children":        childrenList,
			"category_path":   q.CategoryPath,
		}
		responseList = append(responseList, item)
	}

	if responseList == nil {
		responseList = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, gin.H{"data": responseList, "total": total, "page": page, "page_size": pageSize})
}

// ... SyncCategories ...
func (h *Handler) SyncCategories(c *gin.Context) {
	if err := h.repo.SyncCategories(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "目录已同步"})
}

func (h *Handler) GetTree(c *gin.Context) {
	parentIDStr := c.Query("parent_id")
	source := c.Query("source")
	
	// 🔥 核心修复：极致健壮的 userID 获取逻辑
	var userID uint
	if v, exists := c.Get("userID"); exists {
		switch id := v.(type) {
		case uint:
			userID = id
		case int:
			userID = uint(id)
		case int64:
			userID = uint(id)
		case float64:
			userID = uint(id)
		}
	}

	var pID *int
	if parentIDStr != "" && parentIDStr != "0" {
		id, _ := strconv.Atoi(parentIDStr)
		pID = &id
	}

	// 将 userID 传给仓库层
	nodes, err := h.repo.GetTree(pID, source, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取目录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": nodes})
}

// ... UpdateCategory ...
func (h *Handler) UpdateCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.UpdateCategory(uint(id), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// =================================================================
// 🔥 ImportQuestions 终极修复版
// =================================================================
func (h *Handler) ImportQuestions(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "请上传文件"}); return }
	bankName := c.PostForm("bank_name")
	if bankName == "" { c.JSON(http.StatusBadRequest, gin.H{"error": "必须指定题库分类名称"}); return }
	
	src, err := file.Open(); if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "文件打开失败"}); return }; defer src.Close()
	f, err := excelize.OpenReader(src); if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "Excel 解析失败"}); return }
	rows, err := f.GetRows(f.GetSheetName(0)); if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "读取行失败"}); return }

	var roots []*Question
	var lastFingerprint string = ""
	var currentParent *Question = nil
	var lastQType string = ""

	for i, row := range rows {
		if i == 0 || len(row) < 4 { continue }
		getCol := func(idx int) string {
			if idx < len(row) { return strings.TrimSpace(row[idx]) }
			return ""
		}

		originCategory := getCol(1); qType := getCol(2)
		
		fullStem := processContentImages(getCol(3)) 
		optA, optB, optC, optD, optE, optF := processContentImages(getCol(4)), processContentImages(getCol(5)), processContentImages(getCol(6)), processContentImages(getCol(7)), processContentImages(getCol(8)), processContentImages(getCol(9))
		rawCorrect := getCol(10)
		analysis := processContentImages(getCol(11))
		
		difficulty := getCol(12)
		diffVal, _ := strconv.ParseFloat(getCol(13), 64); if diffVal == 0 { diffVal = 0.5 }
		syllabus := getCol(14); cognitiveLevel := getCol(15)

		var finalOpts datatypes.JSON = nil
		optsMap := make(map[string]string)
		hasOption := false
		if optA != "" { optsMap["A"] = optA; hasOption = true }
		if optB != "" { optsMap["B"] = optB; hasOption = true }
		if optC != "" { optsMap["C"] = optC; hasOption = true }
		if optD != "" { optsMap["D"] = optD; hasOption = true }
		if optE != "" { optsMap["E"] = optE; hasOption = true }
		if optF != "" { optsMap["F"] = optF; hasOption = true }
		if hasOption { optsJson, _ := json.Marshal(optsMap); finalOpts = optsJson }

		optsFingerprint := fmt.Sprintf("%s|%s|%s|%s|%s", optA, optB, optC, optD, optE)
		topCategory := ""; if parts := strings.Split(originCategory, ">"); len(parts) > 0 { topCategory = strings.TrimSpace(parts[0]) } else { topCategory = originCategory }

		finalCorrect := ""
		isSubjective := strings.Contains(qType, "问答") || strings.Contains(qType, "论述") || strings.Contains(qType, "案例") || strings.Contains(qType, "名词解释")
		if isSubjective {
			finalCorrect = processContentImages(rawCorrect) 
		} else {
			finalCorrect = strings.TrimSpace(strings.ToUpper(rawCorrect))
		}

		var childQuestion *Question = nil

		if strings.Contains(qType, "A3") || strings.Contains(qType, "A4") {
			parts := strings.Split(fullStem, "\n")
			currentMainStem := ""; currentSubStem := ""

			if len(parts) > 1 {
				currentMainStem = strings.Join(parts[:len(parts)-1], "\n")
				currentSubStem = parts[len(parts)-1]
			} else if strings.Contains(fullStem, "【共用主干】") || strings.Contains(fullStem, "【共用题干】") {
				currentMainStem = fullStem
				currentSubStem = fullStem 
			} else {
				currentMainStem = fullStem; currentSubStem = "" 
			}

			currentFingerprint := cleanStemForFingerprint(currentMainStem)
			isSameGroup := false
			if currentParent != nil && (strings.Contains(lastQType, "A3") || strings.Contains(lastQType, "A4")) {
				if currentFingerprint != "" && lastFingerprint != "" {
					if strings.Contains(currentFingerprint, lastFingerprint) || strings.Contains(lastFingerprint, currentFingerprint) {
						isSameGroup = true
					}
				}
			}

			if isSameGroup {
				if currentSubStem == "" && len(parts) <= 1 { currentSubStem = "" }
				childQuestion = &Question{
					Type: qType, Stem: cleanStem(currentSubStem), Options: finalOpts, Correct: finalCorrect,
					Analysis: analysis, Category: topCategory, CategoryPath: originCategory, Source: bankName, 
					Difficulty: difficulty, DiffValue: diffVal, Syllabus: syllabus, CognitiveLevel: cognitiveLevel,
				}
				currentParent.Children = append(currentParent.Children, *childQuestion)
			} else {
				newParent := &Question{
					Type: qType, Stem: cleanStem(currentMainStem), Category: topCategory, CategoryPath: originCategory,
					Source: bankName, ParentID: nil, Children: []Question{},
				}
				roots = append(roots, newParent)
				currentParent = newParent; lastFingerprint = currentFingerprint; lastQType = qType

				childQuestion = &Question{
					Type: qType, Stem: cleanStem(currentSubStem), Options: finalOpts, Correct: finalCorrect,
					Analysis: analysis, Category: topCategory, CategoryPath: originCategory, Source: bankName, 
					Difficulty: difficulty, DiffValue: diffVal, Syllabus: syllabus, CognitiveLevel: cognitiveLevel,
				}
				currentParent.Children = append(currentParent.Children, *childQuestion)
			}

		} else if strings.Contains(qType, "B1") {
			isSameGroup := false
			if currentParent != nil && strings.Contains(lastQType, "B1") {
				if optsFingerprint == lastFingerprint { isSameGroup = true }
			}
			if isSameGroup {
				childQuestion = &Question{
					Type: qType, Stem: cleanStem(fullStem), Options: nil, Correct: finalCorrect,
					Analysis: analysis, Category: topCategory, CategoryPath: originCategory, Source: bankName, 
					Difficulty: difficulty, DiffValue: diffVal, Syllabus: syllabus, CognitiveLevel: cognitiveLevel,
				}
				currentParent.Children = append(currentParent.Children, *childQuestion)
			} else {
				newParent := &Question{
					Type: qType, Stem: cleanStem(fullStem), Options: finalOpts, Category: topCategory, CategoryPath: originCategory,
					Source: bankName, ParentID: nil, Children: []Question{},
				}
				roots = append(roots, newParent)
				currentParent = newParent; lastFingerprint = optsFingerprint; lastQType = qType
				childQuestion = &Question{
					Type: qType, Stem: cleanStem(fullStem), Options: nil, Correct: finalCorrect,
					Analysis: analysis, Category: topCategory, CategoryPath: originCategory, Source: bankName, 
					Difficulty: difficulty, DiffValue: diffVal, Syllabus: syllabus, CognitiveLevel: cognitiveLevel,
				}
				currentParent.Children = append(currentParent.Children, *childQuestion)
			}

		} else {
			lastFingerprint = ""; currentParent = nil; lastQType = qType
			q := &Question{
				Type: qType, Stem: cleanStem(fullStem), Options: finalOpts, Correct: finalCorrect,
				Analysis: analysis, Category: topCategory, CategoryPath: originCategory, Source: bankName,
				Difficulty: difficulty, DiffValue: diffVal, Syllabus: syllabus, CognitiveLevel: cognitiveLevel, ParentID: nil,
			}
			roots = append(roots, q)
		}
	}

	sort.SliceStable(roots, func(i, j int) bool { return getTypeWeight(roots[i].Type) < getTypeWeight(roots[j].Type) })
	insertCount := 0
	for _, root := range roots {
		if err := db.DB.Create(root).Error; err == nil {
			if len(root.Children) > 0 { insertCount += len(root.Children) } else { insertCount++ }
		}
	}
	h.repo.SyncCategories()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("成功导入 %d 道小题 (图片已转存)", insertCount)})
}

// ... Admin Ops ...
func (h *Handler) GetSources(c *gin.Context) {
	list, err := h.repo.GetSources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

type RenameSourceReq struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

func (h *Handler) RenameSource(c *gin.Context) {
	var req RenameSourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.RenameSource(req.OldName, req.NewName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "题库重命名成功"})
}

type TransferReq struct {
	FromSource string `json:"from_source"`
	ToSource   string `json:"to_source"`
	Category   string `json:"category"`
}

func (h *Handler) TransferCategory(c *gin.Context) {
	var req struct {
		FromSource string `json:"from_source" binding:"required"`
		ToSource   string `json:"to_source" binding:"required"`
		Category   string `json:"category" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx := db.DB.Begin()
	res := tx.Model(&Question{}).Where("source = ? AND category_path LIKE ?", req.FromSource, req.Category+"%").Update("source", req.ToSource)
	if res.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "题目迁移失败"})
		return
	}
	affectedQuestions := res.RowsAffected
	var rootCat Category
	if err := tx.Where("source = ? AND name = ?", req.FromSource, req.Category).First(&rootCat).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "在源题库中找不到该科目目录"})
		return
	}
	catIDs := []uint{rootCat.ID}
	getAllChildIDs(tx, rootCat.ID, &catIDs)
	if err := tx.Model(&Category{}).Where("id IN ?", catIDs).Update("source", req.ToSource).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "目录结构迁移失败"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("迁移成功！共移动 %d 道题，%d 个目录节点", affectedQuestions, len(catIDs))})
}

func getAllChildIDs(tx *gorm.DB, parentID uint, result *[]uint) {
	var children []Category
	tx.Where("parent_id = ?", parentID).Find(&children)
	for _, child := range children {
		*result = append(*result, child.ID)
		getAllChildIDs(tx, child.ID, result)
	}
}

type ReorderReq struct {
	Items []ReorderItem `json:"items"`
}

func (h *Handler) ReorderCategories(c *gin.Context) {
	var req ReorderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.ReorderCategories(req.Items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "顺序已更新"})
}

type UpdateQuestionReq struct {
	Stem           string            `json:"stem"`
	Type           string            `json:"type"`
	Options        map[string]string `json:"options"`
	Correct        string            `json:"correct"`
	Analysis       string            `json:"analysis"`
	Difficulty     string            `json:"difficulty"`
	DiffValue      float64           `json:"diff_value"`
	Syllabus       string            `json:"syllabus"`
	CognitiveLevel string            `json:"cognitive_level"`
}

func (h *Handler) UpdateQuestion(c *gin.Context) {
	id := c.Param("id")
	var req UpdateQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var q Question
	if err := db.DB.First(&q, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}
	q.Stem = req.Stem
	q.Type = req.Type
	q.Correct = req.Correct
	q.Analysis = req.Analysis
	q.DiffValue = req.DiffValue
	q.Difficulty = req.Difficulty
	q.Syllabus = req.Syllabus
	q.CognitiveLevel = req.CognitiveLevel
	if len(req.Options) > 0 {
		optsJson, _ := json.Marshal(req.Options)
		q.Options = optsJson
	} else {
		q.Options = nil
	}
	if err := db.DB.Save(&q).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "题目已更新", "data": q})
}

func (h *Handler) DeleteSource(c *gin.Context) {
	var req struct {
		SourceName string `json:"source_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := db.DB.Begin()

	var qIDs []uint
	tx.Model(&Question{}).Where("source = ?", req.SourceName).Pluck("id", &qIDs)

	if err := hardDeleteQuestions(tx, qIDs); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "题目数据清理失败"})
		return
	}

	if err := tx.Unscoped().Where("source = ?", req.SourceName).Delete(&Category{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "目录清理失败"})
		return
	}

	if err := tx.Where("source = ?", req.SourceName).Delete(&product.ProductContent{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "商品权益清理失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "题库已彻底删除，相关商品权益已同步移除"})
}

func (h *Handler) DeleteQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	tx := db.DB.Begin()
	if err := hardDeleteQuestions(tx, []uint{uint(id)}); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "题目已彻底删除"})
}

type BatchDeleteReq struct {
	IDs []uint `json:"ids" binding:"required"`
}

func (h *Handler) BatchDeleteQuestions(c *gin.Context) {
	var req BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未选择任何题目"})
		return
	}

	tx := db.DB.Begin()
	if err := hardDeleteQuestions(tx, req.IDs); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除失败: " + err.Error()})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("已彻底删除 %d 道题目及关联数据", len(req.IDs))})
}

func (h *Handler) DeleteByCategory(c *gin.Context) {
	categoryPath := c.Query("category_path")
	source := c.Query("source")

	if categoryPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须指定分类路径"})
		return
	}

	tx := db.DB.Begin()

	var qIDs []uint
	qQuery := tx.Model(&Question{}).Where("category_path LIKE ?", categoryPath+"%")
	if source != "" {
		qQuery = qQuery.Where("source = ?", source)
	}
	qQuery.Pluck("id", &qIDs)

	if len(qIDs) > 0 {
		if err := hardDeleteQuestions(tx, qIDs); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "题目清理失败"})
			return
		}
	}

	catQuery := tx.Unscoped().Where("full_path LIKE ?", categoryPath+"%")
	if source != "" {
		catQuery = catQuery.Where("source = ?", source)
	}
	if err := catQuery.Delete(&Category{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "目录结构清理失败"})
		return
	}

	parts := strings.Split(categoryPath, "/")
	if len(parts) == 1 { 
		if err := tx.Where("source = ? AND category = ?", source, parts[0]).Delete(&product.ProductContent{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "商品权益清理失败"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "章节及题目已彻底粉碎，相关商品权益已更新"})
}

// ==========================================
// 🔧 题目纠错反馈系统 (Feedback System)
// ==========================================

func (h *Handler) SubmitFeedback(c *gin.Context) {
	var req struct {
		QuestionID uint   `json:"question_id" binding:"required"`
		Type       string `json:"type" binding:"required"`
		Content    string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var count int64
	db.DB.Model(&QuestionFeedback{}).
		Where("user_id = ? AND question_id = ? AND created_at >= ?", userID, req.QuestionID, todayStart).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "您今天已经反馈过这道题了，请明天再来"})
		return
	}

	fb := QuestionFeedback{
		UserID:     userID,
		QuestionID: req.QuestionID,
		Type:       req.Type,
		Content:    req.Content,
		Status:     0, 
	}

	if err := db.DB.Create(&fb).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "反馈已提交，感谢您的纠错！"})
}

func (h *Handler) AdminListFeedbacks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	statusStr := c.Query("status") 
	offset := (page - 1) * pageSize

	query := db.DB.Model(&QuestionFeedback{}).
		Preload("User").
		Preload("Question").
		Order("created_at desc")

	if statusStr != "" {
		query = query.Where("status = ?", statusStr)
	}

	var total int64
	query.Count(&total)

	var list []QuestionFeedback
	if err := query.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page})
}

func (h *Handler) AdminResolveFeedback(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status     int    `json:"status"` 
		AdminReply string `json:"admin_reply"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData := map[string]interface{}{
		"status":      req.Status,
		"admin_reply": req.AdminReply,
	}

	if err := db.DB.Model(&QuestionFeedback{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "处理完成"})
}