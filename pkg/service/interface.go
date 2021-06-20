package service

import (
	"context"
	"github.com/buildboxapp/yookassa/pkg/api"
	"github.com/buildboxapp/yookassa/pkg/model"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/metric"
	"github.com/buildboxapp/yookassa/pkg/utils"
)

type service struct {
	logger log.Log
	cfg model.Config
	metrics metric.ServiceMetric
	utils utils.Utils
	api api.Api
}

// Service interface
type Service interface {
	Ping(ctx context.Context) (result []model.Pong, err error)
	Pay(ctx context.Context, in model.PayIn) (out model.PayOut, err error)
	Confirmation(ctx context.Context, answer model.AnswerConfirmation, in model.ConfirmationIn) (out model.ConfirmationOut, err error)
}

func New(
	logger log.Log,
	cfg model.Config,
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
