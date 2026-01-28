package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App  AppConfig
	Data DataConfig // 新增：数据配置
	Jwt  JwtConfig // 新增：JWT配置
}

type AppConfig struct {
	Name string
	Env  string
	Port int
}

type DataConfig struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Driver string
	Source string
}

type JwtConfig struct {
    Secret string
    Expire int
}

var Cfg Config

func Load() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("read config failed: %v", err)
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("unmarshal config failed: %v", err)
	}
}