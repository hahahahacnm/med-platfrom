package answer

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"med-platform/internal/common/db"
	"med-platform/internal/question"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ---------------------------------------------------------
// 1. 结构体与初始化
// ---------------------------------------------------------

type Handler struct {
	repo         *Repository
	questionRepo *question.Repository
}

func NewHandler() *Handler {
	return &Handler{
		repo:         NewRepository(),
		questionRepo: question.NewRepository(),
	}
}

type SubmitRequest struct {
	Choice string `json:"choice" binding:"required"`
}

// ---------------------------------------------------------
// 2. 核心写操作 (提交、收藏、移除、重置)
// ---------------------------------------------------------

// Submit 提交答案
func (h *Handler) Submit(c *gin.Context) {
	qID, _ := strconv.Atoi(c.Param("id"))
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := h.getUserID(c)

	// 查题目详情
	q, err := h.questionRepo.GetDetail(uint(qID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	userChoice := strings.TrimSpace(strings.ToUpper(req.Choice))
	correctChoice := strings.TrimSpace(strings.ToUpper(q.Correct))
	isCorrect := (userChoice == correctChoice)

	if userID > 0 {
		// 1. 记录流水
		// 注意：h.repo.CreateOrUpdate 内部现在的逻辑是：
		// 1. 更新 answer_records (最新状态)
		// 2. 增加 user_daily_stats (今日刷题数 +1)
		record := &AnswerRecord{
			UserID:     userID,
			QuestionID: uint(qID),
			Choice:     userChoice,
			IsCorrect:  isCorrect,
		}
		h.repo.CreateOrUpdate(record)

		// 2. 错题本逻辑 (保持原样)
		if !isCorrect {
			targetMistakeID := q.ID
			var mistake UserMistake
			err := db.DB.Where("user_id = ? AND question_id = ?", userID, targetMistakeID).First(&mistake).Error

			if err == gorm.ErrRecordNotFound {
				newMistake := UserMistake{
					UserID:     userID,
					QuestionID: targetMistakeID,
					Choice:     userChoice,
					CreatedAt:  time.Now(),
				}
				db.DB.Create(&newMistake)
			} else {
				mistake.Choice = userChoice
				mistake.UpdatedAt = time.Now()
				db.DB.Save(&mistake)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"is_correct":     isCorrect,
		"user_choice":    userChoice,
		"correct_answer": correctChoice,
		"analysis":       q.Analysis,
	})
}

// ToggleFavorite 收藏/取消
func (h *Handler) ToggleFavorite(c *gin.Context) {
	userID := h.getUserID(c)
	qID, _ := strconv.Atoi(c.Param("id"))

	var q question.Question
	if err := db.DB.First(&q, qID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	targetID := q.ID
	var fav UserFavorite
	result := db.DB.Where("user_id = ? AND question_id = ?", userID, targetID).First(&fav)

	if result.Error == gorm.ErrRecordNotFound {
		newFav := UserFavorite{UserID: userID, QuestionID: targetID}
		db.DB.Create(&newFav)
		c.JSON(http.StatusOK, gin.H{"is_favorite": true, "message": "收藏成功"})
	} else {
		db.DB.Delete(&fav)
		c.JSON(http.StatusOK, gin.H{"is_favorite": false, "message": "已取消收藏"})
	}
}

// RemoveMistake 移除错题
func (h *Handler) RemoveMistake(c *gin.Context) {
	userID := h.getUserID(c)
	qID, _ := strconv.Atoi(c.Param("id"))

	err := db.DB.Where("id = ? OR (user_id = ? AND question_id = ?)", qID, userID, qID).Delete(&UserMistake{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已移出错题本"})
}

// Reset 重置单题记录
func (h *Handler) Reset(c *gin.Context) {
	userID := h.getUserID(c)
	qID, _ := strconv.Atoi(c.Param("id"))
	h.repo.Delete(userID, uint(qID))
	c.JSON(http.StatusOK, gin.H{"message": "已重置"})
}

// ResetChapter 重置章节记录
func (h *Handler) ResetChapter(c *gin.Context) {
	userID := h.getUserID(c)
	category := c.Query("category")
	h.repo.ResetCategory(userID, category)
	c.JSON(http.StatusOK, gin.H{"message": "本章记录已清空"})
}

// ---------------------------------------------------------
// 3. 核心读操作 (错题列表、收藏列表)
// ---------------------------------------------------------

// GetMistakes 获取错题列表
func (h *Handler) GetMistakes(c *gin.Context) {
	h.getPersonalList(c, "user_mistakes")
}

// GetFavorites 获取收藏列表
func (h *Handler) GetFavorites(c *gin.Context) {
	h.getPersonalList(c, "user_favorites")
}

func (h *Handler) getPersonalList(c *gin.Context, tableName string) {
	userID := c.MustGet("userID").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	source := c.Query("source")
	keyword := c.Query("keyword")
	category := c.Query("category")

	var recordIDs []uint
	var questionIDs []uint
	var total int64

	baseQuery := db.DB.Table(tableName).
		Joins("JOIN questions ON "+tableName+".question_id = questions.id").
		Where(tableName+".user_id = ?", userID).
		Where("questions.deleted_at IS NULL")

	if source != "" {
		baseQuery = baseQuery.Where("questions.source = ?", source)
	}
	if category != "" {
		baseQuery = baseQuery.Where("questions.category_path LIKE ?", category+"%")
	}
	if keyword != "" {
		likeStr := "%" + keyword + "%"
		baseQuery = baseQuery.Where("questions.stem LIKE ? OR questions.analysis LIKE ?", likeStr, likeStr)
	}

	baseQuery.Count(&total)

	type Result struct {
		ID         uint
		QuestionID uint
	}
	var results []Result
	orderBy := tableName + ".updated_at desc"
	if tableName == "user_favorites" {
		orderBy = tableName + ".created_at desc"
	}

	baseQuery.Select(tableName+".id, "+tableName+".question_id").
		Order(orderBy).
		Limit(pageSize).Offset(offset).
		Scan(&results)

	for _, r := range results {
		recordIDs = append(recordIDs, r.ID)
		questionIDs = append(questionIDs, r.QuestionID)
	}

	if len(questionIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": total, "page": page, "page_size": pageSize})
		return
	}

	var rawQuestions []question.Question
	db.DB.Where("id IN ?", questionIDs).Find(&rawQuestions)

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

	var responseList []map[string]interface{}
	addedMap := make(map[uint]bool)

	for _, res := range results {
		rawQID := res.QuestionID
		targetID := rawQID
		if pid, ok := parentIDMap[rawQID]; ok {
			targetID = pid
		}

		if addedMap[targetID] {
			continue
		}

		if q, exists := qMap[targetID]; exists {
			item := h.buildQuestionMap(q, userID, tableName == "user_favorites")
			wrapper := map[string]interface{}{
				"id":         res.ID,
				"created_at": time.Now(),
				"question":   item,
			}
			if targetID != rawQID {
				wrapper["focus_child_id"] = rawQID
			}
			responseList = append(responseList, wrapper)
			addedMap[targetID] = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      responseList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ---------------------------------------------------------
// 4. 目录树接口
// ---------------------------------------------------------

func (h *Handler) GetMistakeTree(c *gin.Context) {
	h.getTreeData(c, "user_mistakes")
}

func (h *Handler) GetFavoriteTree(c *gin.Context) {
	h.getTreeData(c, "user_favorites")
}

func (h *Handler) getTreeData(c *gin.Context, tableName string) {
	userID := c.MustGet("userID").(uint)
	parentIDStr := c.Query("parent_id")
	if parentIDStr == "" {
		parentIDStr = c.Query("parent_key")
	}
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
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	var result []map[string]interface{}
	for _, cat := range currentCats {
		var count int64
		db.DB.Table(tableName).
			Joins("JOIN questions ON " + tableName + ".question_id = questions.id").
			Where(tableName + ".user_id = ?", userID).
			Where("questions.category_path LIKE ?", cat.FullPath+"%").
			Where("questions.deleted_at IS NULL").
			Count(&count)

		if count == 0 {
			continue
		}

		isLeaf := false
		if cat.Level >= MaxLevel {
			isLeaf = true
		} else {
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

// ---------------------------------------------------------
// 5. 🔥 仪表盘综合统计接口 (GetDashboardStats) - 嵌套版 🔥
// ---------------------------------------------------------

type DashboardStatsResponse struct {
	TotalCount      int64             `json:"total_count"`
	TodayCount      int64             `json:"today_count"`
	Accuracy        float64           `json:"accuracy"`
	ConsecutiveDays int               `json:"consecutive_days"`
	ActivityMap     []DailyActivity   `json:"activity_map"`
	SubjectAnalysis []SubjectGroup    `json:"subject_analysis"` // 嵌套结构
	RankList        []RankUser        `json:"rank_list"`
}

type DailyActivity struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Level int    `json:"level"`
}

// SubjectGroup 学科组 (Level 1)
type SubjectGroup struct {
	Name     string        `json:"name"`
	Total    int           `json:"total"`
	Accuracy float64       `json:"accuracy"`
	Chapters []ChapterStat `json:"chapters"`
}

// ChapterStat 章节统计 (Level 2)
type ChapterStat struct {
	Name     string  `json:"name"`
	Total    int     `json:"total"`
	Accuracy float64 `json:"accuracy"`
}

type RankUser struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Count    int    `json:"count"`
	Rank     int    `json:"rank"`
}

func (h *Handler) GetDashboardStats(c *gin.Context) {
	uid := c.MustGet("userID").(uint)
	response := DashboardStatsResponse{}

	// =========================================================
	// 1. 基础统计 (🔥 修改：从统计表查数量，从流水表查正确率)
	// =========================================================

	// A. 累计做题总数 = 每日统计表(近1年) + 归档表(1年前)
	var dailyTotal int64
	var archivedTotal int64
	db.DB.Model(&question.UserDailyStat{}).Where("user_id = ?", uid).Select("COALESCE(SUM(count), 0)").Scan(&dailyTotal)
	db.DB.Model(&question.UserArchivedStat{}).Where("user_id = ?", uid).Select("COALESCE(total_count, 0)").Scan(&archivedTotal)
	response.TotalCount = dailyTotal + archivedTotal

	// B. 今日做题数 (查 UserDailyStat)
	todayStr := time.Now().Format("2006-01-02")
	var todayStat question.UserDailyStat
	db.DB.Where("user_id = ? AND date_str = ?", uid, todayStr).First(&todayStat)
	response.TodayCount = int64(todayStat.Count)

	// C. 正确率 (查 AnswerRecord)
	// 正确率反映"当前水平"，所以只统计未被重置的有效记录
	var currentTotal, currentCorrect int64
	db.DB.Model(&AnswerRecord{}).Where("user_id = ?", uid).Count(&currentTotal)
	db.DB.Model(&AnswerRecord{}).Where("user_id = ? AND is_correct = ?", uid, true).Count(&currentCorrect)

	if currentTotal > 0 {
		response.Accuracy = float64(currentCorrect) / float64(currentTotal) * 100
	}

	// =========================================================
	// 2. 学习热力图 (🔥 修改：查 UserDailyStat，速度极快且保留历史)
	// =========================================================
	var stats []question.UserDailyStat
	twoWeeksAgo := time.Now().AddDate(0, 0, -14).Format("2006-01-02")
	db.DB.Where("user_id = ? AND date_str > ?", uid, twoWeeksAgo).Find(&stats)

	activityMap := make(map[string]int)
	for _, s := range stats {
		activityMap[s.DateStr] = s.Count
	}

	for i := 13; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		count := activityMap[d]
		level := 0
		if count > 0 {
			level = 1
		}
		if count > 20 {
			level = 2
		}
		if count > 50 {
			level = 3
		}
		if count > 100 {
			level = 4
		}
		response.ActivityMap = append(response.ActivityMap, DailyActivity{Date: time.Now().AddDate(0, 0, -i).Format("01-02"), Count: count, Level: level})
	}

	// =========================================================
	// 3. 计算连续打卡 (🔥 修改：查 UserDailyStat)
	// =========================================================
	streak := 0
	if activityMap[todayStr] > 0 {
		streak++
		for i := 1; i < 365; i++ {
			d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			var cnt int64
			db.DB.Model(&question.UserDailyStat{}).Where("user_id = ? AND date_str = ?", uid, d).Count(&cnt)
			if cnt > 0 {
				streak++
			} else {
				break
			}
		}
	}
	response.ConsecutiveDays = streak

	// =========================================================
	// 4. 学科能力分析 (依然查 AnswerRecord，反映当前水平)
	// 🔥🔥🔥 逻辑：L1为学科，L2为章节 🔥🔥🔥
	// =========================================================

	type CatStatRaw struct {
		CategoryID int
		IsCorrect  bool
	}
	var rawStats []CatStatRaw
	db.DB.Table("answer_records").
		Select("questions.category_id, answer_records.is_correct").
		Joins("JOIN questions ON answer_records.question_id = questions.id").
		Where("answer_records.user_id = ?", uid).
		Scan(&rawStats)

	var allCats []question.Category
	db.DB.Find(&allCats)

	type CatNode struct {
		Name     string
		ParentID uint
		Level    int
	}
	catMap := make(map[uint]CatNode)
	for _, c := range allCats {
		pid := uint(0)
		if c.ParentID != nil {
			pid = *c.ParentID
		}
		catMap[c.ID] = CatNode{Name: c.Name, ParentID: pid, Level: c.Level}
	}

	// 聚合容器
	type ChapterTemp struct {
		Total   int
		Correct int
	}
	type SubjectTemp struct {
		Total      int
		Correct    int
		ChapterMap map[string]*ChapterTemp
	}
	aggMap := make(map[string]*SubjectTemp)

	for _, r := range rawStats {
		currID := uint(r.CategoryID)
		var subjectName string = "" // Level 1 (例如：妇产科学)
		var chapterName string = "" // Level 2 (例如：产前检查)

		// 向上寻根
		tempID := currID
		var pathNodes []CatNode
		loop := 0
		for tempID != 0 && loop < 10 {
			node, exists := catMap[tempID]
			if !exists {
				break
			}
			pathNodes = append([]CatNode{node}, pathNodes...) // prepend
			tempID = node.ParentID
			loop++
		}

		// 🔥 核心层级定位
		for _, node := range pathNodes {
			if node.Level == 1 {
				subjectName = node.Name
			} // Level 1 = 学科
			if node.Level == 2 {
				chapterName = node.Name
			} // Level 2 = 章节
		}

		// 兜底：如果没找到 L1 (比如直接是根目录下的题)，就用最顶层节点名
		if subjectName == "" && len(pathNodes) > 0 {
			subjectName = pathNodes[0].Name
		}
		// 兜底：如果没有 L2，归入综合练习
		if chapterName == "" {
			chapterName = "综合练习"
		}

		if subjectName != "" {
			if _, ok := aggMap[subjectName]; !ok {
				aggMap[subjectName] = &SubjectTemp{ChapterMap: make(map[string]*ChapterTemp)}
			}
			st := aggMap[subjectName]
			st.Total++
			if r.IsCorrect {
				st.Correct++
			}

			if _, ok := st.ChapterMap[chapterName]; !ok {
				st.ChapterMap[chapterName] = &ChapterTemp{}
			}
			ct := st.ChapterMap[chapterName]
			ct.Total++
			if r.IsCorrect {
				ct.Correct++
			}
		}
	}

	// 格式化输出
	for subName, subData := range aggMap {
		subAcc := 0.0
		if subData.Total > 0 {
			subAcc = float64(subData.Correct) / float64(subData.Total) * 100
		}

		group := SubjectGroup{Name: subName, Total: subData.Total, Accuracy: subAcc}

		for chapName, chapData := range subData.ChapterMap {
			chapAcc := 0.0
			if chapData.Total > 0 {
				chapAcc = float64(chapData.Correct) / float64(chapData.Total) * 100
			}
			group.Chapters = append(group.Chapters, ChapterStat{Name: chapName, Total: chapData.Total, Accuracy: chapAcc})
		}
		// 章节排序：按做题量
		sort.Slice(group.Chapters, func(i, j int) bool { return group.Chapters[i].Total > group.Chapters[j].Total })

		response.SubjectAnalysis = append(response.SubjectAnalysis, group)
	}
	// 学科排序：按做题量
	sort.Slice(response.SubjectAnalysis, func(i, j int) bool {
		return response.SubjectAnalysis[i].Total > response.SubjectAnalysis[j].Total
	})

	// =========================================================
	// 5. 排行榜 (🔥 修改：UserDailyStat + UserArchivedStat)
	// =========================================================
	// 逻辑：热数据 + 冷数据 = 总战绩
	rows, _ := db.DB.Raw(`
		SELECT 
			u.username, 
			u.avatar, 
			(COALESCE(SUM(s.count), 0) + COALESCE(MAX(a.total_count), 0)) as total 
		FROM users u 
		LEFT JOIN user_daily_stats s ON u.id = s.user_id 
		LEFT JOIN user_archived_stats a ON u.id = a.user_id
		GROUP BY u.id 
		ORDER BY total DESC 
		LIMIT 5
	`).Rows()

	if rows != nil {
		defer rows.Close()
		rank := 1
		for rows.Next() {
			var r RankUser
			rows.Scan(&r.Username, &r.Avatar, &r.Count)
			r.Rank = rank
			response.RankList = append(response.RankList, r)
			rank++
		}
	}

	c.JSON(200, gin.H{"data": response})
}

// ---------------------------------------------------------
// 6. 辅助函数
// ---------------------------------------------------------

// GetStats 简单统计接口 (保留用于兼容)
func (h *Handler) GetStats(c *gin.Context) {
	h.GetDashboardStats(c)
}

func (h *Handler) getUserID(c *gin.Context) uint {
	if v, exists := c.Get("userID"); exists {
		if id, ok := v.(uint); ok {
			return id
		}
		if id, ok := v.(float64); ok {
			return uint(id)
		}
		if id, ok := v.(int); ok {
			return uint(id)
		}
	}
	return 0
}

func (h *Handler) buildQuestionMap(q question.Question, userID uint, isFav bool) map[string]interface{} {
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
			var childRecord map[string]interface{} = nil
			if userID > 0 {
				var crecord struct {
					Choice    string
					IsCorrect bool
				}
				db.DB.Table("answer_records").Select("choice, is_correct").Where("user_id = ? AND question_id = ?", userID, child.ID).Order("created_at desc").Limit(1).Scan(&crecord)
				if crecord.Choice != "" {
					childRecord = map[string]interface{}{"choice": crecord.Choice, "is_correct": crecord.IsCorrect}
				}
			}
			childrenList = append(childrenList, map[string]interface{}{
				"id":            child.ID,
				"type":          child.Type,
				"stem":          child.Stem,
				"options":       childOpts,
				"correct":       child.Correct,
				"analysis":      child.Analysis,
				"user_record":   childRecord,
				"syllabus":      child.Syllabus,
				"cognitive_level": child.CognitiveLevel,
			})
		}
	}

	var mainRecord map[string]interface{} = nil
	if userID > 0 && len(q.Children) == 0 {
		var mrecord struct {
			Choice    string
			IsCorrect bool
		}
		db.DB.Table("answer_records").Select("choice, is_correct").Where("user_id = ? AND question_id = ?", userID, q.ID).Order("created_at desc").Limit(1).Scan(&mrecord)
		if mrecord.Choice != "" {
			mainRecord = map[string]interface{}{"choice": mrecord.Choice, "is_correct": mrecord.IsCorrect}
		}
	}

	return map[string]interface{}{
		"id":              q.ID,
		"type":            q.Type,
		"stem":            q.Stem,
		"options":         optionsMap,
		"correct":         q.Correct,
		"analysis":        q.Analysis,
		"children":        childrenList,
		"is_favorite":     isFav,
		"user_record":     mainRecord,
		"difficulty":      q.Difficulty,
		"diff_value":      q.DiffValue,
		"syllabus":        q.Syllabus,
		"cognitive_level": q.CognitiveLevel,
		"category":        q.Category,
	}
}