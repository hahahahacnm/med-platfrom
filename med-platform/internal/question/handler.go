package question

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"med-platform/internal/common/db"
	"med-platform/internal/product"

	"github.com/gin-gonic/gin"
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
// 🔐 [修复版] 核心鉴权辅助函数
// =========================================================
func checkAccess(c *gin.Context, source string, categoryPath string) bool {
	// 1. 获取当前用户身份
	uidRaw, exists := c.Get("userID")
	if !exists { return false }
	userID := uidRaw.(uint)

	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)

	// 2. 特权放行：超管和代理不需要买课
	if role == "admin" || role == "agent" {
		return true
	}

	// 3. 提取一级科目 (核心修复点)
	// 前端传来的可能是 "妇产科学 > 产前检查" (带空格和大于号)
	// 我们先把它标准化为 "妇产科学/产前检查"
	normalizedPath := strings.ReplaceAll(categoryPath, " > ", "/")
	
	// 然后按 "/" 切割
	parts := strings.Split(normalizedPath, "/")
	
	rootCategory := ""
	if len(parts) > 0 {
		// 取第一段，并去掉可能残留的空格
		rootCategory = strings.TrimSpace(parts[0])
	}

	// 此时 rootCategory 应该是干净的 "妇产科学"，能匹配上商品授权了
	return product.NewRepository().CheckPermission(userID, source, rootCategory)
}

// 🔥 [新增] 通用硬删除逻辑：物理删除题目 ID 列表，并清理所有关联表
func hardDeleteQuestions(tx *gorm.DB, questionIDs []uint) error {
	if len(questionIDs) == 0 {
		return nil
	}

	// 1. 查找这些题目的所有子题，一并加入删除列表
	var childIDs []uint
	tx.Model(&Question{}).Where("parent_id IN ?", questionIDs).Pluck("id", &childIDs)
	allIDs := append(questionIDs, childIDs...)

	// 2. 清理关联的用户数据 (收藏、错题、做题记录)
	if err := tx.Exec("DELETE FROM user_favorites WHERE question_id IN ?", allIDs).Error; err != nil { return err }
	if err := tx.Exec("DELETE FROM user_mistakes WHERE question_id IN ?", allIDs).Error; err != nil { return err }
	if err := tx.Exec("DELETE FROM answer_records WHERE question_id IN ?", allIDs).Error; err != nil { return err }

	// 3. 清理笔记系统 (先删点赞/收藏关联表，再删笔记主表)
	if err := tx.Exec("DELETE FROM note_likes WHERE note_id IN (SELECT id FROM notes WHERE question_id IN ?)", allIDs).Error; err != nil { return err }
	if err := tx.Exec("DELETE FROM note_collects WHERE note_id IN (SELECT id FROM notes WHERE question_id IN ?)", allIDs).Error; err != nil { return err }
	if err := tx.Exec("DELETE FROM notes WHERE question_id IN ?", allIDs).Error; err != nil { return err }

	// 4. 最后物理删除题目 (Unscoped 忽略 deleted_at，直接 DELETE)
	if err := tx.Unscoped().Where("id IN ?", allIDs).Delete(&Question{}).Error; err != nil { return err }

	return nil
}

// 辅助：清洗题干
func cleanStem(text string) string {
	text = strings.ReplaceAll(text, "【共用主干】", "")
	text = strings.ReplaceAll(text, "【共用题干】", "")
	text = strings.ReplaceAll(text, "【案例描述】", "")
	return strings.TrimSpace(text)
}

// 辅助：清洗用于比对的指纹 (去除标点、空格、换行)
func cleanStemForFingerprint(text string) string {
	text = cleanStem(text)
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "　", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.ReplaceAll(text, "，", "")
	return text
}

// 辅助：获取题型权重
func getTypeWeight(t string) int {
	t = strings.ToUpper(t)
	if strings.Contains(t, "A1") { return 10 }
	if strings.Contains(t, "A2") { return 20 }
	if strings.Contains(t, "A3") { return 30 }
	if strings.Contains(t, "A4") { return 40 }
	if strings.Contains(t, "B1") { return 50 }
	if strings.Contains(t, "X") { return 60 }
	return 999
}

// List 获取题目列表
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 { page = 1 }
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if pageSize <= 0 { pageSize = 10 }
	if pageSize > 200 { pageSize = 200 }

	category := c.Query("category")
	keyword := c.Query("q")
	source := c.Query("source")

	// 🔥🔥🔥 鉴权守门员 🔥🔥🔥
	if source != "" && category != "" {
		if !checkAccess(c, source, category) {
			c.JSON(http.StatusForbidden, gin.H{"error": "FORBIDDEN", "message": "🔒 您尚未获得该科目的访问授权，请联系管理员或购买相关课程"})
			return
		}
	}

	// 1. 先按常规条件查询
	rawQuestions, total, err := h.repo.List(page, pageSize, category, keyword, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. 🧠 智能寻根逻辑
	var finalQuestions []*Question
	var parentIDsToFetch []uint
	processedParentMap := make(map[uint]bool)

	// A. 预处理：分离父ID
	for _, q := range rawQuestions {
		if q.ParentID != nil && *q.ParentID > 0 {
			if !processedParentMap[*q.ParentID] {
				parentIDsToFetch = append(parentIDsToFetch, *q.ParentID)
				processedParentMap[*q.ParentID] = true
			}
		}
	}

	// B. 批量抓取父题
	parentMap := make(map[uint]*Question)
	if len(parentIDsToFetch) > 0 {
		var parents []*Question
		db.DB.Preload("Children").Where("id IN ?", parentIDsToFetch).Find(&parents)
		for _, p := range parents {
			parentMap[p.ID] = p
		}
	}

	// C. 重组列表
	for i := range rawQuestions {
		q := &rawQuestions[i]

		if q.ParentID != nil && *q.ParentID > 0 {
			// 这是一个子题
			if parent, exists := parentMap[*q.ParentID]; exists {
				// 找到父题 -> 替换并去重
				alreadyAdded := false
				for _, fq := range finalQuestions {
					if fq.ID == parent.ID {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					finalQuestions = append(finalQuestions, parent)
				}
			} else {
				// 🔥 [核心修复] 找不到父题？(可能是脏数据) -> 原样保留子题
				// 如果不加这一步，这道题就凭空消失了，导致前端列表数量 < total
				finalQuestions = append(finalQuestions, q)
			}
		} else {
			// 父题或单题 -> 直接保留
			finalQuestions = append(finalQuestions, q)
		}
	}

	// -------------------------------------------------------
	// 3. 统计数据聚合 (保持不变)
	// -------------------------------------------------------
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
		case uint: userID = id
		case float64: userID = uint(id)
		}
	}

	favMap := make(map[uint]bool)
	recordMap := make(map[uint]interface{})
	noteCountMap := make(map[uint]int64)

	if len(allQIDs) > 0 {
		if userID > 0 {
			var favs []uint
			db.DB.Table("user_favorites").Where("user_id = ? AND question_id IN ?", userID, allQIDs).Pluck("question_id", &favs)
			for _, qid := range favs { favMap[qid] = true }

			type RecordDTO struct { QuestionID uint; Choice string; IsCorrect bool }
			var records []RecordDTO
			db.DB.Table("answer_records").Select("question_id, choice, is_correct").Where("user_id = ? AND question_id IN ?", userID, allQIDs).Order("created_at asc").Scan(&records)
			for _, r := range records {
				if r.Choice != "" { recordMap[r.QuestionID] = map[string]interface{}{"choice": r.Choice, "is_correct": r.IsCorrect} }
			}
		}

		type CountResult struct { QuestionID uint; Total int64 }
		var counts []CountResult
		db.DB.Table("notes").Select("question_id, count(1) as total").Where("question_id IN (?)", allQIDs).Group("question_id").Scan(&counts)
		for _, c := range counts { noteCountMap[c.QuestionID] = c.Total }
	}

	// 4. 最终 JSON 组装
	var responseList []map[string]interface{}
	for _, q := range finalQuestions {
		currentTotalNotes := noteCountMap[q.ID]
		var optionsMap map[string]string
		if len(q.Options) > 0 { _ = json.Unmarshal(q.Options, &optionsMap) }

		var childrenList []map[string]interface{}
		if len(q.Children) > 0 {
			for _, child := range q.Children {
				var childOpts map[string]string
				if !strings.Contains(q.Type, "B1") && len(child.Options) > 0 { _ = json.Unmarshal(child.Options, &childOpts) }
				childNoteCount := noteCountMap[child.ID]
				currentTotalNotes += childNoteCount

				childrenList = append(childrenList, map[string]interface{}{
					"id": child.ID, "type": child.Type, "stem": child.Stem, "options": childOpts, "correct": child.Correct,
					"analysis": child.Analysis, "user_record": recordMap[child.ID], "difficulty": child.Difficulty,
					"diff_value": child.DiffValue, "syllabus": child.Syllabus, "cognitive_level": child.CognitiveLevel,
					"note_count": childNoteCount,
					"category_path": child.CategoryPath, 
				})
			}
		}

		item := map[string]interface{}{
			"id": q.ID, "type": q.Type, "stem": q.Stem, "options": optionsMap, "correct": q.Correct,
			"analysis": q.Analysis, "difficulty": q.Difficulty, "diff_value": q.DiffValue, "syllabus": q.Syllabus,
			"cognitive_level": q.CognitiveLevel, "source": q.Source, "is_favorite": favMap[q.ID],
			"user_record": recordMap[q.ID], "note_count": currentTotalNotes, "children": childrenList,
			"category_path": q.CategoryPath,
		}
		responseList = append(responseList, item)
	}

	if responseList == nil { responseList = []map[string]interface{}{} }
	c.JSON(http.StatusOK, gin.H{"data": responseList, "total": total, "page": page, "page_size": pageSize})
}
// GetDetail 获取详情 (修改版：增加鉴权)
func (h *Handler) GetDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	
	// 1. 先查出题目，拿到它的 Source 和 Category
	q, err := h.repo.GetDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目未找到"})
		return
	}

	// 🔥🔒 2. 鉴权守门员 (新增)
	// 详情页也必须检查，防止用户通过猜ID直接访问
	if !checkAccess(c, q.Source, q.CategoryPath) {
		c.JSON(http.StatusForbidden, gin.H{"error": "FORBIDDEN", "message": "🔒 您无权查看该题目详情"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": q})
}

// ... SyncCategories ...
func (h *Handler) SyncCategories(c *gin.Context) {
	if err := h.repo.SyncCategories(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "目录已同步"})
}

// ... GetTree ...
func (h *Handler) GetTree(c *gin.Context) {
	parentIDStr := c.Query("parent_id")
	source := c.Query("source")
	const MaxLevel = 5
	query := db.DB.Model(&Category{})
	if parentIDStr == "" || parentIDStr == "0" {
		query = query.Where("parent_id IS NULL")
		if source != "" { query = query.Where("source = ?", source) }
	} else {
		query = query.Where("parent_id = ?", parentIDStr)
	}
	query = query.Where("level <= ?", MaxLevel)
	var cats []Category
	if err := query.Order("sort_order asc").Order("id asc").Find(&cats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取目录失败"})
		return
	}
	for i := range cats {
		if cats[i].Level >= MaxLevel {
			cats[i].IsLeaf = true
		} else {
			var count int64
			db.DB.Model(&Category{}).Where("parent_id = ? AND level <= ?", cats[i].ID, MaxLevel).Count(&count)
			cats[i].IsLeaf = (count == 0)
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": cats})
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
// 🔥 ImportQuestions 修复：实事求是，子题目录独立 🔥
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

		originCategory := getCol(1)
		qType := getCol(2)
		fullStem := getCol(3)
		optA, optB, optC, optD, optE, optF := getCol(4), getCol(5), getCol(6), getCol(7), getCol(8), getCol(9)
		correct := getCol(10)
		analysis := getCol(11)
		difficulty := getCol(12)
		diffVal, _ := strconv.ParseFloat(getCol(13), 64)
		if diffVal == 0 { diffVal = 0.5 }
		syllabus := getCol(14)
		cognitiveLevel := getCol(15)

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

		var childQuestion *Question = nil

		// A3/A4 场景
		if strings.Contains(qType, "A3") || strings.Contains(qType, "A4") || strings.Contains(qType, "案例") {
			parts := strings.Split(fullStem, "\n")
			currentMainStem := ""; currentSubStem := fullStem
			if len(parts) > 1 {
				currentMainStem = strings.Join(parts[:len(parts)-1], "\n")
				currentSubStem = parts[len(parts)-1]
			} else if strings.Contains(fullStem, "【共用主干】") || strings.Contains(fullStem, "【共用题干】") {
				currentMainStem = fullStem
			} else {
				currentMainStem = fullStem; currentSubStem = fullStem
			}

			currentFingerprint := cleanStemForFingerprint(currentMainStem)
			isSameGroup := false
			if currentParent != nil && (strings.Contains(lastQType, "A3") || strings.Contains(lastQType, "A4") || strings.Contains(lastQType, "案例")) {
				if currentFingerprint != "" && lastFingerprint != "" {
					if strings.Contains(currentFingerprint, lastFingerprint) || strings.Contains(lastFingerprint, currentFingerprint) {
						isSameGroup = true
					}
				}
			}

			if isSameGroup {
				childQuestion = &Question{
					Type: qType, Stem: cleanStem(currentSubStem), Options: finalOpts, Correct: strings.TrimSpace(strings.ToUpper(correct)),
					Analysis: analysis, 
					Category: topCategory, CategoryPath: originCategory, // 👈 真实数据
					Source: bankName, Difficulty: difficulty, DiffValue: diffVal, Syllabus: syllabus, CognitiveLevel: cognitiveLevel,
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
					Type: qType, Stem: cleanStem(currentSubStem), Options: finalOpts, Correct: strings.TrimSpace(strings.ToUpper(correct)),
					Analysis: analysis, 
					Category: topCategory, CategoryPath: originCategory, // 👈 真实数据
					Source: bankName, Difficulty: difficulty, DiffValue: diffVal, Syllabus: syllabus, CognitiveLevel: cognitiveLevel,
				}
				currentParent.Children = append(currentParent.Children, *childQuestion)
			}

		// B1 场景
		} else if strings.Contains(qType, "B1") {
			isSameGroup := false
			if currentParent != nil && strings.Contains(lastQType, "B1") {
				if optsFingerprint == lastFingerprint { isSameGroup = true }
			}

			if isSameGroup {
				childQuestion = &Question{
					Type: qType, Stem: cleanStem(fullStem), Options: nil, Correct: strings.TrimSpace(strings.ToUpper(correct)),
					Analysis: analysis, 
					Category: topCategory, CategoryPath: originCategory, // 👈 真实数据
					Source: bankName, Difficulty: difficulty, DiffValue: diffVal, Syllabus: syllabus, CognitiveLevel: cognitiveLevel,
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
					Type: qType, Stem: cleanStem(fullStem), Options: nil, Correct: strings.TrimSpace(strings.ToUpper(correct)),
					Analysis: analysis, 
					Category: topCategory, CategoryPath: originCategory, // 👈 真实数据
					Source: bankName, Difficulty: difficulty, DiffValue: diffVal, Syllabus: syllabus, CognitiveLevel: cognitiveLevel,
				}
				currentParent.Children = append(currentParent.Children, *childQuestion)
			}

		// 单题场景
		} else {
			lastFingerprint = ""; currentParent = nil; lastQType = qType
			q := &Question{
				Type: qType, Stem: cleanStem(fullStem), Options: finalOpts, Correct: strings.TrimSpace(strings.ToUpper(correct)),
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
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("成功导入 %d 道小题 (知识点完整性已保留)", insertCount)})
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
	if res.Error != nil { tx.Rollback(); c.JSON(http.StatusInternalServerError, gin.H{"error": "题目迁移失败"}); return }
	affectedQuestions := res.RowsAffected
	var rootCat Category
	if err := tx.Where("source = ? AND name = ?", req.FromSource, req.Category).First(&rootCat).Error; err != nil { tx.Rollback(); c.JSON(http.StatusNotFound, gin.H{"error": "在源题库中找不到该科目目录"}); return }
	catIDs := []uint{rootCat.ID}
	getAllChildIDs(tx, rootCat.ID, &catIDs)
	if err := tx.Model(&Category{}).Where("id IN ?", catIDs).Update("source", req.ToSource).Error; err != nil { tx.Rollback(); c.JSON(http.StatusInternalServerError, gin.H{"error": "目录结构迁移失败"}); return }
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

type ReorderReq struct { Items []ReorderItem `json:"items"` }
func (h *Handler) ReorderCategories(c *gin.Context) {
	var req ReorderReq
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	if err := h.repo.ReorderCategories(req.Items); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, gin.H{"message": "顺序已更新"})
}


type UpdateQuestionReq struct {
	Stem string `json:"stem"`; Type string `json:"type"`; Options map[string]string `json:"options"`; Correct string `json:"correct"`
	Analysis string `json:"analysis"`; Difficulty string `json:"difficulty"`; DiffValue float64 `json:"diff_value"`
	Syllabus string `json:"syllabus"`; CognitiveLevel string `json:"cognitive_level"`
}

func (h *Handler) UpdateQuestion(c *gin.Context) {
	id := c.Param("id"); var req UpdateQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	var q Question
	if err := db.DB.First(&q, id).Error; err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"}); return }
	q.Stem = req.Stem; q.Type = req.Type; q.Correct = req.Correct; q.Analysis = req.Analysis
	q.DiffValue = req.DiffValue; q.Difficulty = req.Difficulty; q.Syllabus = req.Syllabus; q.CognitiveLevel = req.CognitiveLevel
	if len(req.Options) > 0 { optsJson, _ := json.Marshal(req.Options); q.Options = optsJson } else { q.Options = nil }
	if err := db.DB.Save(&q).Error; err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"}); return }
	c.JSON(http.StatusOK, gin.H{"message": "题目已更新", "data": q})
}

// 删库硬删除 (修改版：增加清理商品绑定逻辑)
func (h *Handler) DeleteSource(c *gin.Context) {
    var req struct { SourceName string `json:"source_name"` }
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
    
    tx := db.DB.Begin()
    
    // 1. 找出该题库所有题目ID
    var qIDs []uint
    tx.Model(&Question{}).Where("source = ?", req.SourceName).Pluck("id", &qIDs)
    
    // 2. 物理删除题目
    if err := hardDeleteQuestions(tx, qIDs); err != nil { 
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "题目数据清理失败"})
        return 
    }
    
    // 3. 物理删除目录
    if err := tx.Unscoped().Where("source = ?", req.SourceName).Delete(&Category{}).Error; err != nil { 
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "目录清理失败"})
        return 
    }

    // 4. 🔥 [新增] 清理所有绑定该题库源的商品内容 (级联清理 Path A)
    // 删除了这个源，所有包含这个源的商品里的“肉”都要去掉
    if err := tx.Where("source = ?", req.SourceName).Delete(&product.ProductContent{}).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "商品权益清理失败"})
        return
    }

    tx.Commit()
    c.JSON(http.StatusOK, gin.H{"message": "题库已彻底删除，相关商品权益已同步移除"})
}

// 🔥 [新增] 1. 单题硬删除
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

// 🔥 [新增] 2. 批量硬删除 (Batch Delete)
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
	// 调用通用硬删除函数
	if err := hardDeleteQuestions(tx, req.IDs); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除失败: " + err.Error()})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("已彻底删除 %d 道题目及关联数据", len(req.IDs))})
}

// 按章节彻底删除 (修改版：增加清理商品绑定逻辑)
func (h *Handler) DeleteByCategory(c *gin.Context) {
    categoryPath := c.Query("category_path")
    source := c.Query("source")

    if categoryPath == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "必须指定分类路径"})
        return
    }

    tx := db.DB.Begin()

    // 1. 找出该目录下所有题目 ID
    var qIDs []uint
    qQuery := tx.Model(&Question{}).Where("category_path LIKE ?", categoryPath+"%")
    if source != "" {
        qQuery = qQuery.Where("source = ?", source)
    }
    qQuery.Pluck("id", &qIDs)

    // 2. 物理删除题目
    if len(qIDs) > 0 {
        if err := hardDeleteQuestions(tx, qIDs); err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "题目清理失败"})
            return
        }
    }

    // 3. 物理删除目录节点 (Unscoped)
    catQuery := tx.Unscoped().Where("full_path LIKE ?", categoryPath+"%")
    if source != "" {
        catQuery = catQuery.Where("source = ?", source)
    }
    if err := catQuery.Delete(&Category{}).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "目录结构清理失败"})
        return
    }

    // 4. 🔥 [新增] 清理商品绑定 (仅当删除一级目录时触发)
    // 逻辑：我们的商品绑定只精确到“科目”(一级目录)，不精确到章节。
    // 如果您删除了“内科学/呼吸系统”，不应该影响“内科学”商品的持有。
    // 但如果您删除了“内科学”整个大类，那么所有包含“内科学”的商品都要把这个肉剔除。
    parts := strings.Split(categoryPath, "/")
    if len(parts) == 1 { // 判断是否是一级目录
        if err := tx.Where("source = ? AND category = ?", source, parts[0]).Delete(&product.ProductContent{}).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "商品权益清理失败"})
            return
        }
    }

    tx.Commit()
    c.JSON(http.StatusOK, gin.H{"message": "章节及题目已彻底粉碎，相关商品权益已更新"})
}
