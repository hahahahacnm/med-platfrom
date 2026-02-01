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
// 2. 核心写操作
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
		record := &AnswerRecord{
			UserID:     userID,
			QuestionID: uint(qID),
			Choice:     userChoice,
			IsCorrect:  isCorrect,
		}
		h.repo.CreateOrUpdate(record)

		// 2. 错题本逻辑
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

// Reset 重置单题
func (h *Handler) Reset(c *gin.Context) {
	userID := h.getUserID(c)
	qID, _ := strconv.Atoi(c.Param("id"))
	h.repo.Delete(userID, uint(qID))
	c.JSON(http.StatusOK, gin.H{"message": "已重置"})
}

// ResetChapter 重置章节
func (h *Handler) ResetChapter(c *gin.Context) {
	userID := h.getUserID(c)
	category := c.Query("category")
	h.repo.ResetCategory(userID, category)
	c.JSON(http.StatusOK, gin.H{"message": "本章记录已清空"})
}

// ---------------------------------------------------------
// 3. 核心读操作
// ---------------------------------------------------------

func (h *Handler) GetMistakes(c *gin.Context) {
	h.getPersonalList(c, "user_mistakes")
}

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
// 5. 🔥 仪表盘综合统计接口 (GetDashboardStats) 🔥
// ---------------------------------------------------------

type DashboardStatsResponse struct {
	TotalCount      int64           `json:"total_count"`
	TodayCount      int64           `json:"today_count"`
	Accuracy        float64         `json:"accuracy"`
	ConsecutiveDays int             `json:"consecutive_days"`
	ActivityMap     []DailyActivity `json:"activity_map"`
	SubjectAnalysis []SubjectGroup  `json:"subject_analysis"`
	RankList        []RankUser      `json:"rank_list"`
}

type DailyActivity struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Level int    `json:"level"`
}

type SubjectGroup struct {
	Name     string        `json:"name"`
	Total    int           `json:"total"`
	Accuracy float64       `json:"accuracy"`
	Chapters []ChapterStat `json:"chapters"`
}

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
	
	// 初始化，防止前端接收 null
	response := DashboardStatsResponse{
		ActivityMap:     []DailyActivity{},
		SubjectAnalysis: []SubjectGroup{},
		RankList:        []RankUser{},
	}

	// =========================================================
	// 1. 基础统计
	// =========================================================
	var dailyTotal int64
	var archivedTotal int64
	db.DB.Model(&question.UserDailyStat{}).Where("user_id = ?", uid).Select("COALESCE(SUM(count), 0)").Scan(&dailyTotal)
	db.DB.Model(&question.UserArchivedStat{}).Where("user_id = ?", uid).Select("COALESCE(total_count, 0)").Scan(&archivedTotal)
	response.TotalCount = dailyTotal + archivedTotal

	todayStr := time.Now().Format("2006-01-02")
	var todayStat question.UserDailyStat
	db.DB.Where("user_id = ? AND date_str = ?", uid, todayStr).First(&todayStat)
	response.TodayCount = int64(todayStat.Count)

	var currentTotal, currentCorrect int64
	db.DB.Model(&AnswerRecord{}).Where("user_id = ?", uid).Count(&currentTotal)
	db.DB.Model(&AnswerRecord{}).Where("user_id = ? AND is_correct = ?", uid, true).Count(&currentCorrect)

	if currentTotal > 0 {
		response.Accuracy = float64(currentCorrect) / float64(currentTotal) * 100
	}

	// =========================================================
	// 2. 学习热力图 (修复：适配前端 split('-')[2])
	// =========================================================
	var stats []question.UserDailyStat
	twoWeeksAgo := time.Now().AddDate(0, 0, -14).Format("2006-01-02")
	db.DB.Where("user_id = ? AND date_str > ?", uid, twoWeeksAgo).Find(&stats)

	activityMap := make(map[string]int)
	for _, s := range stats {
		activityMap[s.DateStr] = s.Count
	}

	for i := 13; i >= 0; i-- {
		// 这里必须用完整格式，否则前端 Dashboard.vue 里的 day.date.split('-')[2] 会越界报错
		fullDate := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		count := activityMap[fullDate]
		level := 0
		if count > 0 { level = 1 }
		if count > 20 { level = 2 }
		if count > 50 { level = 3 }
		if count > 100 { level = 4 }
		response.ActivityMap = append(response.ActivityMap, DailyActivity{Date: fullDate, Count: count, Level: level})
	}

	// =========================================================
	// 3. 计算连续打卡
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
	// 4. 学科能力分析 (修复核心：强制显示学科，拒绝"未知")
	// =========================================================
	type CatStatRaw struct {
		CategoryID int
		IsCorrect  bool
	}
	var rawStats []CatStatRaw
	
	// 查所有记录 (不区分软删，保证历史记录可查)
	db.DB.Table("answer_records").
		Select("questions.category_id, answer_records.is_correct").
		Joins("JOIN questions ON answer_records.question_id = questions.id").
		Where("answer_records.user_id = ?", uid).
		Scan(&rawStats)

	// 🔥 核心修复：Unscoped() 必须加！否则被删的分类查不到，链条断裂就会显示未知
	var allCats []question.Category
	db.DB.Unscoped().Find(&allCats)

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
		var subjectName string = "未知学科"
		var chapterName string = "综合练习"

		tempID := currID
		var pathNodes []CatNode
		loop := 0
		
		// 1. 向上追溯，构建路径 [根, ..., 父, 子]
		for tempID != 0 && loop < 10 {
			node, exists := catMap[tempID]
			if !exists {
				// ID对不上，可能是野题，break
				break
			}
			pathNodes = append([]CatNode{node}, pathNodes...) // 插到最前面
			tempID = node.ParentID
			loop++
		}

		// 2. 智能提取学科名
		// 优先找 Level 1 (大科目)，如果找不到，就用根节点 (pathNodes[0])
		foundSubject := false
		
		for _, node := range pathNodes {
			if node.Level == 1 {
				subjectName = node.Name
				foundSubject = true
			}
			if node.Level == 2 {
				chapterName = node.Name
			}
		}

		// 🔥 兜底逻辑：如果没找到标准 Level 1 (比如分类只有一层，或者题库是 Level 0)，直接用根节点名
		if !foundSubject && len(pathNodes) > 0 {
			subjectName = pathNodes[0].Name
		}
		
		// 聚合
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

	// 转为 Response 结构
	for subName, subData := range aggMap {
		subAcc := 0.0
		if subData.Total > 0 {
			subAcc = float64(subData.Correct) / float64(subData.Total) * 100
		}

		group := SubjectGroup{Name: subName, Total: subData.Total, Accuracy: subAcc, Chapters: []ChapterStat{}}

		for chapName, chapData := range subData.ChapterMap {
			chapAcc := 0.0
			if chapData.Total > 0 {
				chapAcc = float64(chapData.Correct) / float64(chapData.Total) * 100
			}
			group.Chapters = append(group.Chapters, ChapterStat{Name: chapName, Total: chapData.Total, Accuracy: chapAcc})
		}
		sort.Slice(group.Chapters, func(i, j int) bool { return group.Chapters[i].Total > group.Chapters[j].Total })

		response.SubjectAnalysis = append(response.SubjectAnalysis, group)
	}
	sort.Slice(response.SubjectAnalysis, func(i, j int) bool {
		return response.SubjectAnalysis[i].Total > response.SubjectAnalysis[j].Total
	})

	// =========================================================
	// 5. 今日卷王榜 (安全查询)
	// =========================================================
	rows, err := db.DB.Raw(`
		SELECT 
			u.username, 
			u.avatar, 
			s.count as total 
		FROM user_daily_stats s
		JOIN users u ON s.user_id = u.id 
		WHERE s.date_str = ? 
		ORDER BY total DESC 
		LIMIT 5
	`, todayStr).Rows()

	if err == nil && rows != nil {
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
				"id":              child.ID,
				"type":            child.Type,
				"stem":            child.Stem,
				"options":         childOpts,
				"correct":         child.Correct,
				"analysis":        child.Analysis,
				"user_record":     childRecord,
				"syllabus":        child.Syllabus,
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

// ---------------------------------------------------------
// 7. 🔥 新增：今日卷王榜分页接口 (独立接口，支持无限加载) 🔥
// ---------------------------------------------------------

// RankUserDetail 增加一些详细信息，比如学校，让榜单更好看
type RankUserDetail struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"` // 显示昵称更友好
	Avatar   string `json:"avatar"`
	School   string `json:"school"`   // 加上学校
	Count    int    `json:"count"`
	Rank     int    `json:"rank"`     // 绝对排名
}

func (h *Handler) GetDailyRank(c *gin.Context) {
	// 1. 分页参数解析
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	// 2. 安全限制：最大只允许查前 100 名
	// 也就是 offset 不能超过 100
	maxRankLimit := 100
	offset := (page - 1) * pageSize

	if offset >= maxRankLimit {
		// 超过100名，直接返回空，告诉前端到底了
		c.JSON(http.StatusOK, gin.H{
			"data":      []interface{}{},
			"page":      page,
			"has_more":  false,
			"message":   "仅展示前100名卷王",
		})
		return
	}

	// 修正 pageSize，防止最后一页溢出 100
	// 例如：当前 offset 是 90，pageSize 是 20，那只能取 10 个
	if offset+pageSize > maxRankLimit {
		pageSize = maxRankLimit - offset
	}

	todayStr := time.Now().Format("2006-01-02")
	var rankList []RankUserDetail

	// 3. 查询数据库 (关联 users 表获取头像、昵称、学校)
	// Order: count DESC (做题多在前), updated_at ASC (同样多，先做完的在前)
	rows, err := db.DB.Table("user_daily_stats").
		Select("users.id, users.username, users.nickname, users.avatar, users.school, user_daily_stats.count").
		Joins("JOIN users ON user_daily_stats.user_id = users.id").
		Where("user_daily_stats.date_str = ?", todayStr).
		Order("user_daily_stats.count DESC").
		Order("user_daily_stats.updated_at ASC"). 
		Limit(pageSize).
		Offset(offset).
		Rows()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排行榜失败"})
		return
	}
	defer rows.Close()

	// 4. 组装数据，计算绝对排名
	currentRank := offset + 1 // 排名 = 偏移量 + 1
	for rows.Next() {
		var r RankUserDetail
		// 注意 Scan 的顺序要和 Select 一致
		rows.Scan(&r.UserID, &r.Username, &r.Nickname, &r.Avatar, &r.School, &r.Count)
		
		// 如果没有昵称，显示用户名
		if r.Nickname == "" {
			r.Nickname = r.Username
		}
		
		r.Rank = currentRank
		rankList = append(rankList, r)
		currentRank++
	}

	// 5. 判断是否还有更多
	// 如果取出来的数据量 < pageSize，说明不够分了，肯定是最后一页
	// 或者已经达到了 100 名的界限
	hasMore := len(rankList) == pageSize && (offset+pageSize) < maxRankLimit

	c.JSON(http.StatusOK, gin.H{
		"data":      rankList,
		"page":      page,
		"has_more":  hasMore,
	})
}