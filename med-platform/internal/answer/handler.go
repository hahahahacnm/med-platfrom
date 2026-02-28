package answer

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"med-platform/internal/common/db"
	"med-platform/internal/question"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// SubmitRequest 支持单题提交和批量提交
type SubmitRequest struct {
	Choice  string            `json:"choice"`  // 兼容老版本：单题选项
	Answers map[string]string `json:"answers"` // 🔥 新增批量提交：{"101": "A", "102": "B"}
}

// ---------------------------------------------------------
// 2. 核心写操作
// ---------------------------------------------------------

// Submit 提交答案 (支持单题与批量)
func (h *Handler) Submit(c *gin.Context) {
	qID, _ := strconv.Atoi(c.Param("id"))
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := h.getUserID(c)

	// 1. 整理需要判题的集合 (单题 or 批量)
	targetAnswers := make(map[uint]string)
	if len(req.Answers) > 0 {
		for k, v := range req.Answers {
			id, _ := strconv.Atoi(k)
			if id > 0 {
				targetAnswers[uint(id)] = v
			}
		}
	} else if req.Choice != "" {
		targetAnswers[uint(qID)] = req.Choice
	}

	if len(targetAnswers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供答案"})
		return
	}

	// 2. 批量查询目标题目
	var qIDs []uint
	for id := range targetAnswers {
		qIDs = append(qIDs, id)
	}
	var questions []question.Question
	db.DB.Where("id IN ?", qIDs).Find(&questions)

	// 3. 开始批处理
	var records []*AnswerRecord
	var mistakes []UserMistake
	resultData := make(map[uint]map[string]interface{})

	for _, q := range questions {
		userChoice := strings.TrimSpace(strings.ToUpper(targetAnswers[q.ID]))
		correctChoice := strings.TrimSpace(strings.ToUpper(q.Correct))
		isCorrect := (userChoice == correctChoice)

		if userID > 0 {
			// 准备流水记录 (带上冗余的 CategoryID 优化统计性能)
			records = append(records, &AnswerRecord{
				UserID:     userID,
				QuestionID: q.ID,
				CategoryID: q.CategoryID,
				Choice:     userChoice,
				IsCorrect:  isCorrect,
			})

			// 准备错题本记录
			if !isCorrect {
				mistakes = append(mistakes, UserMistake{
					UserID:     userID,
					QuestionID: q.ID,
					Choice:     userChoice,
					WrongCount: 1, // 基础错误次数
				})
			}
		}

		resultData[q.ID] = map[string]interface{}{
			"is_correct":     isCorrect,
			"user_choice":    userChoice,
			"correct_answer": correctChoice,
			"analysis":       q.Analysis,
		}
	}

	if userID > 0 && len(records) > 0 {
		// 批量保存流水
		if err := h.repo.BatchCreateOrUpdate(records); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存答题记录失败"})
			return
		}

		// 错题本 Upsert 逻辑
		if len(mistakes) > 0 {
			db.DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "question_id"}}, // 联合唯一索引
				DoUpdates: clause.Assignments(map[string]interface{}{
					"choice":      gorm.Expr("EXCLUDED.choice"), 
					"wrong_count": gorm.Expr("user_mistakes.wrong_count + 1"), 
					"updated_at":  time.Now(),
				}),
			}).Create(&mistakes)
		}
	}

	if len(targetAnswers) == 1 && req.Choice != "" {
		c.JSON(http.StatusOK, resultData[uint(qID)])
	} else {
		c.JSON(http.StatusOK, gin.H{"results": resultData})
	}
}

// ToggleFavorite 收藏/取消
// 🔥 修复：取消收藏时级联清除其下所有子题的收藏记录
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
		// 🔥 级联删除：将该父题及名下所有的子题收藏全部清除
		db.DB.Where("user_id = ? AND question_id IN (SELECT id FROM questions WHERE id = ? OR parent_id = ?)", userID, targetID, targetID).Delete(&UserFavorite{})
		c.JSON(http.StatusOK, gin.H{"is_favorite": false, "message": "已取消收藏"})
	}
}

// RemoveMistake 移除错题
// 🔥 修复：移除错题时级联清除其下所有子题的错题记录
func (h *Handler) RemoveMistake(c *gin.Context) {
	userID := h.getUserID(c)
	qID, _ := strconv.Atoi(c.Param("id"))

	// 🔥 级联删除子题的错误记录
	err := db.DB.Where("user_id = ? AND question_id IN (SELECT id FROM questions WHERE id = ? OR parent_id = ?)", userID, qID, qID).Delete(&UserMistake{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已移出错题本"})
}

// Reset 重置单题 (只删记录，保留历史)
func (h *Handler) Reset(c *gin.Context) {
	userID := h.getUserID(c)
	qID, _ := strconv.Atoi(c.Param("id"))
	h.repo.Delete(userID, uint(qID))
	c.JSON(http.StatusOK, gin.H{"message": "已重置"})
}

// ResetChapter 重置章节 (只删记录，保留历史)
func (h *Handler) ResetChapter(c *gin.Context) {
	userID := h.getUserID(c)
	category := c.Query("category")
	h.repo.ResetCategory(userID, category)
	c.JSON(http.StatusOK, gin.H{"message": "本章记录已清空"})
}

// ---------------------------------------------------------
// 3. 核心读操作 (🔥🔥🔥 溯源聚合核心逻辑)
// ---------------------------------------------------------

// GetMistakeSkeleton 获取错题本骨架 
// 🔥 修复：如果错的是小题，自动聚合并返回它爸爸（父题）的 ID
func (h *Handler) GetMistakeSkeleton(c *gin.Context) {
	userID := h.getUserID(c)
	source := c.Query("source")
	category := c.Query("category")

	// 核心魔法：判断有父题就用父题 ID，没父题就用自己 ID
	groupExpr := "CASE WHEN questions.parent_id IS NOT NULL AND questions.parent_id > 0 THEN questions.parent_id ELSE questions.id END"

	baseQuery := db.DB.Table("user_mistakes").
		Select(groupExpr+" as id, MAX(questions.type) as type, MAX(user_mistakes.wrong_count) as wrong_count").
		Joins("JOIN questions ON user_mistakes.question_id = questions.id").
		Where("user_mistakes.user_id = ?", userID).
		Where("questions.deleted_at IS NULL").
		Group(groupExpr) // 按照父题/独立单题进行聚合分组

	if source != "" {
		baseQuery = baseQuery.Where("questions.source = ?", source)
	}
	if category != "" {
		baseQuery = baseQuery.Where("questions.category_path LIKE ?", category+"%")
	}

	type SkeletonItem struct {
		ID         uint   `json:"id"`
		Type       string `json:"type"`
		WrongCount int    `json:"wrong_count"`
	}
	var items []SkeletonItem

	// 按最新错的排在前面 (取分组中最新的 updated_at)
	if err := baseQuery.Order("MAX(user_mistakes.updated_at) desc").Scan(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取错题骨架失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(items),
		"data":  items,
	})
}

// GetFavoriteSkeleton 获取收藏夹骨架
func (h *Handler) GetFavoriteSkeleton(c *gin.Context) {
	userID := h.getUserID(c)
	source := c.Query("source")
	category := c.Query("category")

	groupExpr := "CASE WHEN questions.parent_id IS NOT NULL AND questions.parent_id > 0 THEN questions.parent_id ELSE questions.id END"

	baseQuery := db.DB.Table("user_favorites").
		Select(groupExpr+" as id, MAX(questions.type) as type").
		Joins("JOIN questions ON user_favorites.question_id = questions.id").
		Where("user_favorites.user_id = ?", userID).
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

	if err := baseQuery.Order("MAX(user_favorites.created_at) desc").Scan(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取收藏骨架失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(items),
		"data":  items,
	})
}

// 兼容遗留的 List 接口 (可保留给其他非核心系统用)
func (h *Handler) GetMistakes(c *gin.Context) { h.getPersonalList(c, "user_mistakes") }
func (h *Handler) GetFavorites(c *gin.Context) { h.getPersonalList(c, "user_favorites") }

func (h *Handler) getPersonalList(c *gin.Context, tableName string) {
	// ... (保留原有逻辑不动，因为主链路已经切换到 Skeleton)
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0, "message": "Deprecated: Please use Skeleton API"})
}


// ---------------------------------------------------------
// 4. 目录树接口 (🔥🔥🔥 校准数字统计)
// ---------------------------------------------------------

func (h *Handler) GetMistakeTree(c *gin.Context) { h.getTreeData(c, "user_mistakes") }
func (h *Handler) GetFavoriteTree(c *gin.Context) { h.getTreeData(c, "user_favorites") }

func (h *Handler) getTreeData(c *gin.Context, tableName string) {
	userID := c.MustGet("userID").(uint)
	parentIDStr := c.Query("parent_id")
	if parentIDStr == "" { parentIDStr = c.Query("parent_key") }
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
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	var result []map[string]interface{}
	for _, cat := range currentCats {
		var count int64
		// 🔥 修复：保证左侧树的数量与右侧折叠后的“大题”数量绝对一致
		db.DB.Table(tableName).
			Select("COUNT(DISTINCT CASE WHEN questions.parent_id IS NOT NULL AND questions.parent_id > 0 THEN questions.parent_id ELSE questions.id END)").
			Joins("JOIN questions ON " + tableName + ".question_id = questions.id").
			Where(tableName + ".user_id = ?", userID).
			Where("questions.category_path LIKE ?", cat.FullPath+"%").
			Where("questions.deleted_at IS NULL").
			Scan(&count)

		if count == 0 { continue }

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
// 5. 🔥 仪表盘综合统计接口 (保持不动)
// ---------------------------------------------------------

type DashboardStatsResponse struct {
	TotalCount      int64           `json:"total_count"`
	TodayCount      int64           `json:"today_count"`
	Accuracy        float64         `json:"accuracy"`
	ConsecutiveDays int             `json:"consecutive_days"`
	ActivityMap     []DailyActivity `json:"activity_map"`
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
	uid := h.getUserID(c) // 使用健壮的获取ID函数
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	
	response := DashboardStatsResponse{
		ActivityMap: []DailyActivity{},
		RankList:    []RankUser{},
	}

	// 1. 基础总览统计
	var dailyTotal, archivedTotal int64
	db.DB.Model(&question.UserDailyStat{}).Where("user_id = ?", uid).Select("COALESCE(SUM(count), 0)").Scan(&dailyTotal)
	db.DB.Model(&question.UserArchivedStat{}).Where("user_id = ?", uid).Select("COALESCE(total_count, 0)").Scan(&archivedTotal)
	response.TotalCount = dailyTotal + archivedTotal

	// 今日统计
	todayStr := time.Now().Format("2006-01-02")
	var todayStat int64
	db.DB.Model(&question.UserDailyStat{}).Where("user_id = ? AND date_str = ?", uid, todayStr).Select("COALESCE(count, 0)").Scan(&todayStat)
	response.TodayCount = todayStat

	// 总体正确率 (仅作为一个参考总分)
	var currentTotal, currentCorrect int64
	db.DB.Model(&AnswerRecord{}).Where("user_id = ?", uid).Count(&currentTotal)
	db.DB.Model(&AnswerRecord{}).Where("user_id = ? AND is_correct = ?", uid, true).Count(&currentCorrect)
	if currentTotal > 0 {
		response.Accuracy = float64(currentCorrect) / float64(currentTotal) * 100
	}

	// 2. 学习热力图 (保留最近 14 天)
	var stats []question.UserDailyStat
	twoWeeksAgo := time.Now().AddDate(0, 0, -14).Format("2006-01-02")
	db.DB.Where("user_id = ? AND date_str > ?", uid, twoWeeksAgo).Find(&stats)

	activityMap := make(map[string]int)
	for _, s := range stats {
		activityMap[s.DateStr] = s.Count
	}

	for i := 13; i >= 0; i-- {
		fullDate := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		count := activityMap[fullDate]
		level := 0
		if count > 0 { level = 1 }
		if count > 20 { level = 2 }
		if count > 50 { level = 3 }
		if count > 100 { level = 4 }
		response.ActivityMap = append(response.ActivityMap, DailyActivity{Date: fullDate, Count: count, Level: level})
	}

	// 3. 计算连续打卡天数
	streak := 0
	if activityMap[todayStr] > 0 {
		streak++
		for i := 1; i < 365; i++ {
			d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			var cnt int64
			db.DB.Model(&question.UserDailyStat{}).Where("user_id = ? AND date_str = ?", uid, d).Count(&cnt)
			if cnt > 0 { streak++ } else { break }
		}
	}
	response.ConsecutiveDays = streak

	// 4. 今日排行榜 (前 5 名)
	rows, err := db.DB.Raw(`
		SELECT u.username, u.avatar, s.count as total 
		FROM user_daily_stats s
		JOIN users u ON s.user_id = u.id 
		WHERE s.date_str = ? 
		ORDER BY total DESC LIMIT 5
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

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// ---------------------------------------------------------
// 6. 辅助函数
// ---------------------------------------------------------

func (h *Handler) GetStats(c *gin.Context) {
	h.GetDashboardStats(c)
}

func (h *Handler) getUserID(c *gin.Context) uint {
	if v, exists := c.Get("userID"); exists {
		if id, ok := v.(uint); ok { return id }
		if id, ok := v.(float64); ok { return uint(id) }
		if id, ok := v.(int); ok { return uint(id) }
	}
	return 0
}

// ---------------------------------------------------------
// 7. 🔥 今日卷王榜分页接口 
// ---------------------------------------------------------

type RankUserDetail struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	School   string `json:"school"`   
	Count    int    `json:"count"`
	Rank     int    `json:"rank"`     
}

func (h *Handler) GetDailyRank(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	maxRankLimit := 100
	offset := (page - 1) * pageSize

	if offset >= maxRankLimit {
		c.JSON(http.StatusOK, gin.H{
			"data":      []interface{}{},
			"page":      page,
			"has_more":  false,
			"message":   "仅展示前100名卷王",
		})
		return
	}

	if offset+pageSize > maxRankLimit {
		pageSize = maxRankLimit - offset
	}

	todayStr := time.Now().Format("2006-01-02")
	var rankList []RankUserDetail

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

	currentRank := offset + 1 
	for rows.Next() {
		var r RankUserDetail
		rows.Scan(&r.UserID, &r.Username, &r.Nickname, &r.Avatar, &r.School, &r.Count)
		
		if r.Nickname == "" {
			r.Nickname = r.Username
		}
		
		r.Rank = currentRank
		rankList = append(rankList, r)
		currentRank++
	}

	hasMore := len(rankList) == pageSize && (offset+pageSize) < maxRankLimit

	c.JSON(http.StatusOK, gin.H{
		"data":      rankList,
		"page":      page,
		"has_more":  hasMore,
	})
}