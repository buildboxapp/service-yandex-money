package api

import (
	"context"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/yookassa/pkg/model"
	"github.com/buildboxapp/yookassa/pkg/utils"
)

type api struct {
	ctx context.Context
	cfg model.Config
	utl utils.Utils
	logger log.Log
}

type Api interface {
	AttrUpdate(uid, name, value, src, editor string) (err error)
	CreateObjForm(data map[string]string) (res model.ResponseData, err error)
}

func New(
	ctx context.Context,
	cfg model.Config,
	utl utils.Utils,
	logger log.Log,
) Api {
	return &api{
		ctx,
		cfg,
		utl,
		logger,
	}
}