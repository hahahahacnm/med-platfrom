package logger

import (
	"log"

	"go.uber.org/zap"
)

var Log *zap.Logger

func Init(env string) {
	var err error

	if env == "prod" {
		Log, err = zap.NewProduction()
	} else {
		Log, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatalf("init logger failed: %v", err)
	}
}
