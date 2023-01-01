package main

import (
	"math/rand"
	"time"

	"github.com/cloudhut/common/logging"
	"go.uber.org/zap"

	"github.com/cloudhut/owl-shop/pkg/config"
	"github.com/cloudhut/owl-shop/pkg/shop"
)

func main() {
	// Initialize seed for fake service
	rand.Seed(time.Now().UTC().UnixNano())

	startupLogger := zap.NewExample()
	cfg, err := config.LoadConfig(startupLogger)
	if err != nil {
		startupLogger.Fatal("failed to parse config", zap.Error(err))
	}

	logger := logging.NewLogger(&cfg.Logger, "owl_shop")

	shopSvc, err := shop.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize shop", zap.Error(err))
	}
	err = shopSvc.Start()
	if err != nil {
		logger.Fatal("failed to start shop", zap.Error(err))
	}
}
