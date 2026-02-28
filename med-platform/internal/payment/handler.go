package payment

import (
	"fmt"
	"math"
	"net/http"
	"net/url"    // 🔥 修正点：由 "url" 改为 "net/url"
	"strconv"
	"strings"    // 🔥 确保导出功能需要的 strings 包已引入
	"time"

	"med-platform/internal/common/db"
	"med-platform/internal/sysconfig"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ==========================================
// 1. 原有支付相关接口 (在线直充)
// ==========================================

// CreatePay 创建赞助订单
func (h *Handler) CreatePay(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	// 1. 接收赞助金额
	var req struct {
		Amount  float64 `json:"amount" binding:"required,gt=0"`
		Channel string  `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "金额格式错误"})
		return
	}

	// 2. 规则校验
	if req.Amount < 1.0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "最低赞助 1 元起"})
		return
	}

	// 3. 计算获得积分 (1元 = 10积分)
	points := int(math.Floor(req.Amount * 10))

	// 4. 创建订单
	orderNo := uuid.New().String()
	order := Order{
		OrderNo:       orderNo,
		UserID:        userID,
		Amount:        req.Amount,
		PointsAwarded: points,
		Status:        "PENDING",
	}

	if err := db.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
		return
	}

	// 5. 发起支付 (GetPaymentStrategy 定义在 service.go 中)
	strat := GetPaymentStrategy()
	subject := fmt.Sprintf("赞助本站 - 获赠%d积分", points)

	payUrl, err := strat.Pay(orderNo, req.Amount, subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "支付初始化失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pay_url":  payUrl,
		"order_no": orderNo,
		"points":   points,
	})
}

func (h *Handler) MockSuccess(c *gin.Context) {
	orderNo := c.Query("out_trade_no")
	tradeNo := "MOCK_" + uuid.New().String()
	if err := SettleOrder(orderNo, tradeNo); err != nil {
		c.String(500, "error: "+err.Error())
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `<h1 style="color:green;text-align:center;margin-top:50px">🎉 赞助成功！积分已到账</h1><script>setTimeout(()=>window.close(), 2000)</script>`)
}

func (h *Handler) AlipayNotify(c *gin.Context) {
	req := c.Request
	strat := GetPaymentStrategy()
	outTradeNo, tradeNo, ok, _ := strat.HandleNotify(req)
	if ok {
		SettleOrder(outTradeNo, tradeNo)
		c.String(200, "success")
	} else {
		c.String(400, "fail")
	}
}

// ==========================================
// 2. 卡密/激活码相关接口
// ==========================================

// RedeemCode 用户兑换激活码 (含订单流水 & 动态代理分润)
func (h *Handler) RedeemCode(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "激活码不能为空"})
		return
	}

	codeStr := strings.TrimSpace(req.Code)

	// 开启事务处理兑换逻辑
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var code ActivationCode
		// 1. 悲观锁查码，防止并发多次兑换
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code = ?", codeStr).First(&code).Error; err != nil {
			return fmt.Errorf("激活码无效或不存在")
		}

		if code.Status == 1 {
			return fmt.Errorf("该激活码已被使用")
		}

		// 2. 标记激活码为已使用
		now := time.Now()
		code.Status = 1
		code.UsedByID = userID
		code.UsedAt = &now
		if err := tx.Save(&code).Error; err != nil {
			return fmt.Errorf("更新激活码状态失败")
		}

		// 3. 给用户加积分
		if err := tx.Table("users").Where("id = ?", userID).
			Update("points", gorm.Expr("points + ?", code.Points)).Error; err != nil {
			return fmt.Errorf("增加积分失败")
		}

		// 4. 生成一笔真实的支付订单记录
		orderAmount := float64(code.Points) / 10.0 // 1元 = 10积分
		syntheticOrder := Order{
			OrderNo:       "CD_" + code.Code, 
			TradeNo:       "REDEEM_SYS",      
			UserID:        userID,
			Amount:        orderAmount,
			PointsAwarded: code.Points,
			Status:        "PAID",            
			PayTime:       &now,
		}
		if err := tx.Create(&syntheticOrder).Error; err != nil {
			return fmt.Errorf("生成订单流水失败")
		}

		// 5. 🔥🔥🔥 激活码代理分润 (从配置中心动态获取) 🔥🔥🔥
		type SimpleUser struct {
			ID        uint
			InvitedBy uint
		}
		var currentUser SimpleUser
		if err := tx.Table("users").Select("id, invited_by").Where("id = ?", userID).Scan(&currentUser).Error; err == nil && currentUser.InvitedBy > 0 {
			
		var agentID uint
		if err := tx.Table("users").Select("id").Where("id = ?", currentUser.InvitedBy).Scan(&agentID).Error; err == nil && agentID > 0 {
			
			// 🔥 直接调用强类型获取函数，代码量减少 70%，逻辑更清晰
			rate := sysconfig.GetFloat(sysconfig.KeyAgentRateCard, 0.15)

			profit := orderAmount * rate

			if profit > 0 {
				commLog := CommissionLog{
					AgentID:        agentID,
					FromUserID:     userID,
					OrderNo:        syntheticOrder.OrderNo,
					OrderAmount:    orderAmount,
					Profit:         profit,
					AppliedRate:    rate,           
					Description:    "卡密兑换代理分润",
					WithdrawStatus: 0, 
				}
				tx.Create(&commLog)
			}
		}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "兑换成功！积分已到账"})
}

// GenerateCodes 管理员批量生成激活码
func (h *Handler) GenerateCodes(c *gin.Context) {
	var req struct {
		Count  int `json:"count" binding:"required,gt=0,lte=500"` // 单次最多生成500个
		Points int `json:"points" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，请检查数量和积分数"})
		return
	}

	var newCodes []ActivationCode
	for i := 0; i < req.Count; i++ {
		raw := strings.ReplaceAll(uuid.New().String(), "-", "")
		codeStr := fmt.Sprintf("TK-%s-%s", strings.ToUpper(raw[:4]), strings.ToUpper(raw[4:8]))
		newCodes = append(newCodes, ActivationCode{
			Code:   codeStr,
			Points: req.Points,
			Status: 0,
		})
	}

	if err := db.DB.Create(&newCodes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("成功生成 %d 个激活码", req.Count)})
}

// ListCodes 管理员获取激活码列表
func (h *Handler) ListCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status") // 0=未使用, 1=已使用, 空=全部

	var codes []ActivationCode
	var total int64

	dbQuery := db.DB.Model(&ActivationCode{})
	if status != "" {
		s, _ := strconv.Atoi(status)
		dbQuery = dbQuery.Where("status = ?", s)
	}

	if err := dbQuery.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计失败"})
		return
	}

	err := dbQuery.Order("created_at desc, id desc").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&codes).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  codes,
		"total": total,
	})
}

// ManualUpdatePoints 管理员手动给用户加减积分
func (h *Handler) ManualUpdatePoints(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
		Points int  `json:"points" binding:"required"` // 正数增加，负数扣除
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var currentPoints int
		if err := tx.Table("users").Select("points").Where("id = ?", req.UserID).Scan(&currentPoints).Error; err != nil {
			return fmt.Errorf("用户不存在")
		}

		if currentPoints+req.Points < 0 {
			return fmt.Errorf("用户积分不足以扣除")
		}

		return tx.Table("users").Where("id = ?", req.UserID).
			Update("points", gorm.Expr("points + ?", req.Points)).Error
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// 注意确保你的 import 里包含了 "net/url" 和 "strings"

// ExportCodes 导出未使用卡密为 TXT 文件
func (h *Handler) ExportCodes(c *gin.Context) {
	pointsStr := c.Query("points")
	if pointsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请指定要导出的积分额度"})
		return
	}

	points, err := strconv.Atoi(pointsStr)
	if err != nil || points <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的积分额度"})
		return
	}

	var codes []ActivationCode
	// 🔥 核心逻辑：只查对应额度，且 status = 0 (未使用) 的卡密
	if err := db.DB.Where("points = ? AND status = 0", points).Find(&codes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if len(codes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有找到该额度的可用卡密"})
		return
	}

	// 拼接文本，一行一个 (使用 \r\n 兼容 Windows 记事本)
	var sb strings.Builder
	for _, code := range codes {
		sb.WriteString(code.Code + "\r\n")
	}

	// 设置下载用的 Header，并解决中文文件名乱码问题
	filename := fmt.Sprintf("卡密-%d积分.txt", points)
	encodedFilename := url.QueryEscape(filename)

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encodedFilename))
	c.Header("Content-Type", "text/plain; charset=utf-8")
	
	// 直接返回纯文本内容
	c.String(http.StatusOK, sb.String())
}