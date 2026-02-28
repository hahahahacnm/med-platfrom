package sysconfig

import (
	"log"
	"med-platform/internal/common/db"
	"strconv"
	"sync"
)

// 🔥 定义系统标准配置 Key 常量，防止拼写错误
const (
	KeyAgentRateDirect = "AGENT_COMMISSION_RATE_DIRECT" // 在线支付分润比例
	KeyAgentRateCard   = "AGENT_COMMISSION_RATE_CARD"   // 卡密兑换分润比例
)

var (
	configCache = make(map[string]string)
	cacheMutex  sync.RWMutex
)

// InitConfig 系统启动时调用：加载配置并初始化默认值
func InitConfig() {
	// 1. 检查并初始化默认值（Seeding）
	seedDefaults()
	// 2. 加载到内存
	LoadAllConfigs()
}

// seedDefaults 检查关键配置是否存在，不存在则写入初始值
func seedDefaults() {
	defaults := []SysConfig{
		{Key: KeyAgentRateDirect, Value: "0.20", Description: "在线支付代理分润比例 (0.0-1.0)"},
		{Key: KeyAgentRateCard, Value: "0.15", Description: "卡密兑换代理分润比例 (0.0-1.0)"},
	}

	for _, d := range defaults {
		var count int64
		db.DB.Model(&SysConfig{}).Where("key = ?", d.Key).Count(&count)
		if count == 0 {
			db.DB.Create(&d)
			log.Printf("🌱 初始化系统配置项: %s = %s", d.Key, d.Value)
		}
	}
}

// LoadAllConfigs 从数据库加载所有配置到内存
func LoadAllConfigs() {
	var configs []SysConfig
	if err := db.DB.Find(&configs).Error; err != nil {
		log.Printf("❌ 无法加载系统配置: %v", err)
		return
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	configCache = make(map[string]string)
	for _, c := range configs {
		configCache[c.Key] = c.Value
	}
	log.Println("✅ 系统动态配置已成功加载到内存")
}

// GetConfig 原有的获取字符串方法
func GetConfig(key string) string {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return configCache[key]
}

// 🔥 新增：强类型获取 Float64 的工具函数，带兜底逻辑
func GetFloat(key string, defaultValue float64) float64 {
	valStr := GetConfig(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		log.Printf("⚠️ 配置项 %s 格式错误(%s)，使用兜底值: %v", key, valStr, defaultValue)
		return defaultValue
	}
	return val
}