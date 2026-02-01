package forum

import (
	"med-platform/internal/common/db"
	"med-platform/internal/common/model"   // 🔥 引入通用模型
	"med-platform/internal/common/service" // 🔥 引入通用通知服务
	"med-platform/internal/common/uploader"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// =======================
// 1. 板块管理 (Board)
// =======================

// CreateBoard 创建板块 (Admin)
func (h *Handler) CreateBoard(c *gin.Context) {
	var req ForumBoard
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 图标固化逻辑
	if req.Icon != "" && strings.Contains(req.Icon, "/uploads/temp/") {
		finalPaths := uploader.ConfirmImages([]string{req.Icon}, "forum")
		if len(finalPaths) > 0 {
			req.Icon = finalPaths[0]
		}
	}

	if err := db.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建板块失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "板块创建成功", "data": req})
}

// ListBoards 获取板块列表
func (h *Handler) ListBoards(c *gin.Context) {
	var boards []ForumBoard
	if err := db.DB.Order("sort_order desc, id asc").Find(&boards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取板块失败"})
		return
	}

	type BoardWithStats struct {
		ForumBoard
		PostCount    int64 `json:"post_count"`
		CommentCount int64 `json:"comment_count"`
	}

	var response []BoardWithStats
	for _, board := range boards {
		var pCount int64
		var cCount int64
		db.DB.Model(&ForumPost{}).Where("board_id = ?", board.ID).Count(&pCount)
		db.DB.Model(&ForumPost{}).Where("board_id = ?", board.ID).Select("COALESCE(SUM(comment_count), 0)").Row().Scan(&cCount)

		response = append(response, BoardWithStats{
			ForumBoard:   board,
			PostCount:    pCount,
			CommentCount: cCount,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateBoard 更新板块
func (h *Handler) UpdateBoard(c *gin.Context) {
	id := c.Param("id")
	var req ForumBoard
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var board ForumBoard
	if err := db.DB.First(&board, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "板块不存在"})
		return
	}

	if req.Icon != "" && strings.Contains(req.Icon, "/uploads/temp/") {
		finalPaths := uploader.ConfirmImages([]string{req.Icon}, "forum")
		if len(finalPaths) > 0 {
			req.Icon = finalPaths[0]
		}
	}

	req.ID = board.ID
	db.DB.Model(&board).Select("*").Updates(req)
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": board})
}

// DeleteBoard 删除板块
func (h *Handler) DeleteBoard(c *gin.Context) {
	id := c.Param("id")
	var count int64
	db.DB.Model(&ForumPost{}).Where("board_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该板块下还有帖子，无法删除"})
		return
	}
	db.DB.Delete(&ForumBoard{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "板块已删除"})
}

// =======================
// 2. 帖子管理 (Post)
// =======================

// CreatePost 发布帖子
func (h *Handler) CreatePost(c *gin.Context) {
	var req struct {
		BoardID  uint   `json:"board_id" binding:"required"`
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Summary  string `json:"summary"`
		IsPinned bool   `json:"is_pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)
	role := c.MustGet("role").(string)

	var board ForumBoard
	if err := db.DB.First(&board, req.BoardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "板块不存在"})
		return
	}

	if board.IsLocked && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "该板块仅限管理员发帖"})
		return
	}

	re := regexp.MustCompile(`src="([^"]*\/uploads\/temp\/[^"]+)"`)
	matches := re.FindAllStringSubmatch(req.Content, -1)
	var relativePaths []string
	var originalUrls []string
	for _, match := range matches {
		fullUrl := match[1]
		if idx := strings.Index(fullUrl, "/uploads/temp/"); idx != -1 {
			relativePaths = append(relativePaths, fullUrl[idx:])
			originalUrls = append(originalUrls, fullUrl)
		}
	}
	if len(relativePaths) > 0 {
		finalPaths := uploader.ConfirmImages(relativePaths, "forum")
		for i, oldUrl := range originalUrls {
			req.Content = strings.Replace(req.Content, oldUrl, finalPaths[i], -1)
		}
	}

	summary := req.Summary
	if summary == "" {
		plainText := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(req.Content, "")
		plainText = strings.ReplaceAll(plainText, "\n", "")
		runes := []rune(plainText)
		if len(runes) > 100 {
			summary = string(runes[:100]) + "..."
		} else {
			summary = string(runes)
		}
	}

	post := ForumPost{
		BoardID:  req.BoardID,
		AuthorID: userID,
		Title:    req.Title,
		Content:  req.Content,
		Summary:  summary,
		IsPinned: req.IsPinned && role == "admin",
	}

	if err := db.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "data": post})
}

// ListPosts 获取帖子列表
func (h *Handler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	boardID := c.Query("board_id")
	keyword := c.Query("q")

	query := db.DB.Model(&ForumPost{}).Preload("Author").Preload("Board").Order("is_global desc, is_pinned desc, created_at desc")
	if boardID != "" {
		query = query.Where("board_id = ?", boardID)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)
	var posts []ForumPost
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&posts)

	c.JSON(http.StatusOK, gin.H{"data": posts, "total": total, "page": page})
}

// GetPostDetail 获取详情
func (h *Handler) GetPostDetail(c *gin.Context) {
	id := c.Param("id")
	var post ForumPost
	if err := db.DB.Preload("Author").Preload("Board").First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}
	db.DB.Model(&post).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	post.ViewCount++
	c.JSON(http.StatusOK, gin.H{"data": post})
}

// DeletePost 删除帖子
func (h *Handler) DeletePost(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)
	role := c.MustGet("role").(string)

	var post ForumPost
	if err := db.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}

	if role != "admin" && post.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除"})
		return
	}

	db.DB.Unscoped().Where("post_id = ?", post.ID).Delete(&ForumComment{})
	db.DB.Unscoped().Delete(&post)
	c.JSON(http.StatusOK, gin.H{"message": "已彻底删除"})
}

// UploadImage 辅助功能
func (h *Handler) UploadImage(c *gin.Context) {
	url, err := uploader.SaveImageWithHash(c, "file", 5*1024*1024)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "上传成功", "url": url, "data": map[string]string{"url": url}})
}

// =======================
// 4. 评论与互动 (Comment & Report)
// =======================

// CreateComment 发表/回复评论 (🔥 已升级：接入通用通知系统)
func (h *Handler) CreateComment(c *gin.Context) {
	var req struct {
		PostID   uint   `json:"post_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
		ParentID *uint  `json:"parent_id"` // 可选：回复某条评论
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID := c.MustGet("userID").(uint)

	var post ForumPost
	if err := db.DB.First(&post, req.PostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}

	// 图片固化
	re := regexp.MustCompile(`src="([^"]*\/uploads\/temp\/[^"]+)"`)
	matches := re.FindAllStringSubmatch(req.Content, -1)
	var relativePaths []string
	var originalUrls []string
	for _, match := range matches {
		fullUrl := match[1]
		if idx := strings.Index(fullUrl, "/uploads/temp/"); idx != -1 {
			relativePaths = append(relativePaths, fullUrl[idx:])
			originalUrls = append(originalUrls, fullUrl)
		}
	}
	if len(relativePaths) > 0 {
		finalPaths := uploader.ConfirmImages(relativePaths, "forum")
		for i, oldUrl := range originalUrls {
			req.Content = strings.Replace(req.Content, oldUrl, finalPaths[i], -1)
		}
	}

	comment := ForumComment{
		PostID:   req.PostID,
		AuthorID: userID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := db.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "评论失败"})
		return
	}

	// 🔥🔥🔥 核心升级：调用通用通知服务 🔥🔥🔥
	// 1. 确定通知目标
	var targetUserID uint = post.AuthorID // 默认通知楼主
	// 2. 如果是回复楼中楼，则改为通知层主
	if req.ParentID != nil && *req.ParentID > 0 {
		var parentComment ForumComment
		if err := db.DB.Select("author_id").First(&parentComment, *req.ParentID).Error; err == nil {
			targetUserID = parentComment.AuthorID
		}
	}

	// 3. 发送通知 (sourceType="forum", sourceID=帖子ID)
	// 前端收到 "forum" 类型通知时，点击会跳转到 /post/:id
	service.SendNotification(
		targetUserID,
		userID,
		"forum",
		post.ID,
		req.Content, // 评论内容摘要
		post.Title,  // 帖子标题
	)

	// 更新评论数
	db.DB.Model(&post).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1))
	c.JSON(http.StatusOK, gin.H{"message": "评论成功"})
}

// ListComments 获取评论列表
func (h *Handler) ListComments(c *gin.Context) {
	postID := c.Query("post_id")
	var comments []ForumComment
	db.DB.Where("post_id = ?", postID).Preload("Author").Order("created_at asc").Find(&comments)
	c.JSON(http.StatusOK, gin.H{"data": comments})
}

// DeleteComment 删除评论
func (h *Handler) DeleteComment(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)
	role := c.MustGet("role").(string)

	var comment ForumComment
	if err := db.DB.First(&comment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
		return
	}

	if role != "admin" && comment.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除"})
		return
	}

	db.DB.Unscoped().Delete(&comment)
	db.DB.Model(&ForumPost{ID: comment.PostID}).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1))
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// CreateReport 提交举报
func (h *Handler) CreateReport(c *gin.Context) {
	var req struct {
		TargetID   uint   `json:"target_id" binding:"required"`
		TargetType string `json:"target_type" binding:"required"`
		Reason     string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)
	report := ForumReport{
		TargetID:   req.TargetID,
		TargetType: req.TargetType,
		Reason:     req.Reason,
		ReporterID: userID,
		Status:     0,
	}
	db.DB.Create(&report)
	c.JSON(http.StatusOK, gin.H{"message": "举报已提交"})
}

// =======================
// 5. 管理员专属接口 (Admin Only)
// =======================

// AdminListComments 管理员获取评论列表
func (h *Handler) AdminListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("q")

	query := db.DB.Model(&ForumComment{}).Preload("Author").Order("created_at desc")
	if keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)
	var comments []ForumComment
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&comments)

	c.JSON(http.StatusOK, gin.H{"data": comments, "total": total, "page": page})
}

// AdminListReports 获取举报列表
func (h *Handler) AdminListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	query := db.DB.Model(&ForumReport{}).Preload("Reporter").Order("status asc, created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)
	var reports []ForumReport
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&reports)

	c.JSON(http.StatusOK, gin.H{"data": reports, "total": total})
}

// AdminResolveReport 处理举报
func (h *Handler) AdminResolveReport(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Model(&ForumReport{}).Where("id = ?", id).Update("status", 1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "举报已标记为处理"})
}

// AdminGetReportContent 举报预览
func (h *Handler) AdminGetReportContent(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID := c.Query("target_id")

	if targetType == "post" {
		var post ForumPost
		if err := db.DB.Preload("Author").First(&post, targetID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子已不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"content": post.Content,
			"author":  post.Author.Nickname,
			"title":   post.Title,
			"type":    "post",
		})
	} else if targetType == "comment" {
		var comment ForumComment
		if err := db.DB.Preload("Author").First(&comment, targetID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "评论已不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"content": comment.Content,
			"author":  comment.Author.Nickname,
			"title":   "评论内容",
			"type":    "comment",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未知类型"})
	}
}

// =======================
// 6. 通用消息通知接口 (🔥 新增：提供给前台铃铛使用)
// =======================

// GetNotifications 获取我的通知列表 (全局)
func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var notifs []model.Notification

	// 查询未读的和最近已读的
	// 注意：这里使用了 model.Notification (通用表)
	db.DB.Where("user_id = ?", userID).
		Preload("Sender").
		Order("is_read asc, created_at desc"). // 未读优先
		Limit(20).                             // 只取最近20条
		Find(&notifs)

	// 统计未读数
	var unreadCount int64
	db.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&unreadCount)

	c.JSON(http.StatusOK, gin.H{"data": notifs, "unread_count": unreadCount})
}

// ReadNotification 标记单条已读
func (h *Handler) ReadNotification(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id := c.Param("id")

	db.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)

	c.JSON(http.StatusOK, gin.H{"message": "已读"})
}

// ReadAllNotifications 全部已读
func (h *Handler) ReadAllNotifications(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	db.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)

	c.JSON(http.StatusOK, gin.H{"message": "全部已读"})
}