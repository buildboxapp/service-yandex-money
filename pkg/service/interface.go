package service

import (
	"context"
	"github.com/buildboxapp/service-yandex-money/pkg/api"
	"github.com/buildboxapp/service-yandex-money/pkg/config"
	"github.com/buildboxapp/service-yandex-money/pkg/model"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/metric"
	"github.com/buildboxapp/service-yandex-money/pkg/utils"
)

type service struct {
	logger log.Log
	cfg config.Config
	metrics metric.ServiceMetric
	utils utils.Utils
	api api.Api
}

// Service interface
type Service interface {
	Ping(ctx context.Context) (result []model.Pong, err error)
	Pay(ctx context.Context, in model.PayIn) (out model.PayOut, err error)
	Confirmation(ctx context.Context, in model.AnswerConfirmation) (out model.ConfirmationOut, err error)
}

func New(
	logger log.Log,
	cfg config.Config,
	metrics metric.ServiceMetric,
	utils utils.Utils,
	api api.Api,
) Service {
	return &service{
		logger,
		cfg,
		metrics,
		utils,
		api,
	}
}
