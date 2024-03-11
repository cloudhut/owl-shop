package shop

import (
	"context"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/cloudhut/owl-shop/pkg/config"
)

// MetaService creates things that do not necessarily belong to one specific service,
// such as ACLs, Kafka connectors, SASL users, Quotas, inactive consumer groups etc.
// It may also create more resources that aren't actively used by any of the services
// of owlshop. For example, it may create additional schemas.
type MetaService struct {
	cfg        config.Shop
	logger     *zap.Logger
	kafkaCl    *kgo.Client
	kafkaAdmCl *kadm.Client
}

func NewMetaService(cfg config.Shop, logger *zap.Logger, kafkaCl *kgo.Client) (*MetaService, error) {
	return &MetaService{
		cfg:        cfg,
		logger:     logger,
		kafkaCl:    kafkaCl,
		kafkaAdmCl: kadm.NewClient(kafkaCl),
	}, nil
}

func (svc *MetaService) Initialize(ctx context.Context) error {
	if !svc.cfg.Meta.Enabled {
		return nil
	}

	// 1. Try to create ACLs
	// The ACL creation may still fail (e.g. due to missing permissions or disabled authorization)
	// but we won't inspect and log the error message in that case.
	metaServiceACLs := kadm.NewACLs().
		Allow("User:"+svc.cfg.GlobalPrefix+"meta-service").
		AllowHosts("*", "127.0.0.1").
		ResourcePatternType(kadm.ACLPatternLiteral).
		Groups("*").
		Topics("*").
		Clusters().
		Operations(kadm.OpAll)
	if _, err := svc.kafkaAdmCl.CreateACLs(ctx, metaServiceACLs); err != nil {
		svc.logger.Info("failed to create ACLs for meta-service", zap.Error(err))
	}

	deliveryServiceACLs := kadm.NewACLs().
		Allow("User:"+svc.cfg.GlobalPrefix+"delivery-service").
		AllowHosts("*").
		ResourcePatternType(kadm.ACLPatternLiteral).
		Topics("*").
		Groups("delivery-service").
		Operations(kadm.OpCreate, kadm.OpDescribe, kadm.OpRead, kadm.OpWrite)
	if _, err := svc.kafkaAdmCl.CreateACLs(ctx, deliveryServiceACLs); err != nil {
		svc.logger.Info("failed to create ACLs for delivery-service", zap.Error(err))
	}

	return nil
}
