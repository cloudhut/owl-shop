package main

import (
	"github.com/cloudhut/common/logging"
	"github.com/cloudhut/owl-shop/pkg/shop"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

func main() {
	// Initialize seed for fake service
	rand.Seed(time.Now().UTC().UnixNano())

	startupLogger := zap.NewExample()
	cfg, err := LoadConfig(startupLogger)
	if err != nil {
		startupLogger.Fatal("failed to parse config", zap.Error(err))
	}

	logger := logging.NewLogger(&cfg.Logger, "owl_shop")

	shopSvc, err := shop.New(cfg.Shop, logger)
	if err != nil {
		logger.Fatal("failed to initialize shop", zap.Error(err))
	}
	err = shopSvc.Start()
	if err != nil {
		logger.Fatal("failed to start shop", zap.Error(err))
	}
}
