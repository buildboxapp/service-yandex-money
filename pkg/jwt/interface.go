package jwt

import (
	"context"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/service-yandex-money/pkg/config"
	"github.com/buildboxapp/service-yandex-money/pkg/utils"
	"net/http"
)

type jwt struct {
	ctx context.Context
	cfg config.Config
	utl utils.Utils
	logger log.Log
}

type token struct {
	uid string `json:"uid"`
	roles string `json:"roles"`
	access struct{
		read string `json:"read"`
		write string `json:"write"`
		delete string `json:"delete"`
		admin string `json:"admin"`
	} `json:"access"`
	deny struct{
		read string `json:"read"`
		write string `json:"write"`
		delete string `json:"delete"`
		admin string `json:"admin"`
	} `json:"deny"`
	info struct {
		name string `json:"name"`
		clientType string `json:"client_type"`
	} `json:"info"`
}

type Jwt interface {
	Get(r *http.Request) (token token, err error)
}

type Token interface {
	Uid() (result string, err error)
	Name() (result string, err error)
	SetUid(value string) (err error)
	SetName(value string) (err error)
}

func New(
	ctx context.Context,
	cfg config.Config,
	utl utils.Utils,
	logger log.Log,
) Jwt {
	return &jwt{
		ctx,
		cfg,
		utl,
		logger,
	}
}