package zap

import (
	"testing"

	"go.uber.org/zap"
)

func TestZap(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger = logger.Named("main")
	logger.Error("oooooohhhhhh no", zap.String("hello", "motto"))

	logger = logger.Named("kafka")
	logger.Error("oooooohhhhhh no", zap.String("hello", "motto"))
}
