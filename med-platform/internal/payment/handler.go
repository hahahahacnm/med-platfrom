package payment

import (
	"fmt"
	"net/http"

	"med-platform/internal/common/db"
	"med-platform/internal/product"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct{}

func NewHandler() *Handler {
	// 自动检查并创建 orders 表
	db.DB.AutoMigrate(&Order{})
	return &Handler{}
}

// 1. CreatePay 创建订单接口
func (h *Handler) CreatePay(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	// 接收 SkuID
	var req struct {
		SkuID   uint   `json:"sku_id" binding:"required"`
		Channel string `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 先查 SKU (获取价格)
	var sku product.ProductSku
	if err := db.DB.First(&sku, req.SkuID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品规格不存在"})
		return
	}

	// 2. 再查 Product (获取商品名、检查上架状态)
	var prod product.Product
	if err := db.DB.First(&prod, sku.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "关联商品数据异常"})
		return
	}

	// 🛑 检查商品是否已下架
	if !prod.IsOnShelf {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该商品已下架，无法购买"})
		return
	}

	// 3. 落库（待支付）
	orderNo := uuid.New().String()
	order := Order{
		OrderNo:   orderNo,
		UserID:    userID,
		ProductID: prod.ID,  // 记录属于哪个大类商品
		SkuID:     sku.ID,   // 🔥 记录买了哪个规格
		Amount:    sku.Price, // 🔥 使用 SKU 的价格
		Status:    "PENDING",
	}

	if err := db.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
		return
	}

	// 4. 获取策略，生成支付链接
	// 直接调用同包下的 service.go 中的工厂方法，不需要 import strategy 包
	strat := GetPaymentStrategy() 

	// 生成支付描述：商品名 - 规格名
	subject := fmt.Sprintf("%s - %s", prod.Name, sku.Name)

	payUrl, err := strat.Pay(orderNo, sku.Price, subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "支付初始化失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pay_url":  payUrl,
		"order_no": orderNo,
	})
}

// 2. MockSuccess 模拟支付成功回调
func (h *Handler) MockSuccess(c *gin.Context) {
	orderNo := c.Query("out_trade_no")
	if orderNo == "" {
		c.String(http.StatusBadRequest, "订单号丢失")
		return
	}

	tradeNo := "MOCK_TRADE_" + uuid.New().String()

	err := SettleOrder(orderNo, tradeNo)
	if err != nil {
		c.String(http.StatusInternalServerError, "发货失败: "+err.Error())
		return
	}

	html := `
		<div style="text-align:center; padding-top:50px; font-family:sans-serif;">
			<h1 style="color:#18a058;">✅ 支付成功！</h1>
			<p>商品已自动到账，正在跳转回首页...</p>
			<script>
				setTimeout(function(){
					location.href = 'http://localhost:5173/payment-test'; 
				}, 2000);
			</script>
		</div>
	`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// 3. AlipayNotify 真实支付宝回调
func (h *Handler) AlipayNotify(c *gin.Context) {
	c.String(http.StatusOK, "success")
}