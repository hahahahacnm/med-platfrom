package config

import (
	"log"

	"github.com/spf13/viper"
)

// 全局配置变量
var GlobalConfig Config

// 🔥🔥🔥 修复 1：定义 Cfg 指针，兼容旧代码 (jwt.go, db.go) 🔥🔥🔥
// 这样旧代码里的 config.Cfg.xxx 依然能工作，指向 GlobalConfig
var Cfg *Config = &GlobalConfig

type Config struct {
	App     AppConfig     `mapstructure:"app"`
	Log     LogConfig     `mapstructure:"log"`
	Data    DataConfig    `mapstructure:"data"`
	Jwt     JwtConfig     `mapstructure:"jwt"`
	Payment PaymentConfig `mapstructure:"payment"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type DataConfig struct {
	Database DatabaseConfig `mapstructure:"database"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	Source string `mapstructure:"source"`
}

type JwtConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

// 💰 支付配置结构体
type PaymentConfig struct {
	Mode string `mapstructure:"mode"` // dev 或 prod

	// 🔥🔥🔥 修复 2：必须补上 Driver 字段！🔥🔥🔥
	// service.go 靠这个字段判断是用 "mock" 还是 "alipay"
	Driver string `mapstructure:"driver"`

	Domain string       `mapstructure:"domain"` // 回调域名
	Alipay AlipayConfig `mapstructure:"alipay"`
	Wechat WechatConfig `mapstructure:"wechat"`
}

type AlipayConfig struct {
	AppID      string `mapstructure:"app_id"`
	PrivateKey string `mapstructure:"private_key"`
	PublicKey  string `mapstructure:"public_key"`
	NotifyURL  string `mapstructure:"notify_url"`
	ReturnURL  string `mapstructure:"return_url"`
}

type WechatConfig struct {
	AppID      string `mapstructure:"app_id"`
	MchID      string `mapstructure:"mch_id"`
	ApiV3Key   string `mapstructure:"api_v3_key"`
	PrivateKey string `mapstructure:"private_key"`
}

func Load() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("❌ 读取配置文件失败: %v", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		log.Fatalf("❌ 解析配置文件失败: %v", err)
	}

	// 🔥 确保 Cfg 指向加载好数据的 GlobalConfig
	Cfg = &GlobalConfig

	log.Printf("✅ 配置加载成功，当前环境: %s, 端口: %d, 支付模式: %s, 驱动: %s",
		GlobalConfig.App.Env,
		GlobalConfig.App.Port,
		GlobalConfig.Payment.Mode,
		GlobalConfig.Payment.Driver, // 打印一下驱动，确认读到了
	)
}