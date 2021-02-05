package handlers

import (
	"github.com/buildboxapp/service-yandex-money/pkg/config"
	"github.com/buildboxapp/service-yandex-money/pkg/jwt"
	"github.com/buildboxapp/service-yandex-money/pkg/service"
	bblog "github.com/buildboxapp/lib/log"
	"net/http"
)

type handlers struct {
	service service.Service
	logger bblog.Log
	cfg config.Config
	jwt jwt.Jwt
}

type Handlers interface {
	Alive(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	Pay(w http.ResponseWriter, r *http.Request)
	Confirmation(w http.ResponseWriter, r *http.Request)
}

func New(
	service service.Service,
	logger bblog.Log,
	cfg config.Config,
	jwt jwt.Jwt,
) Handlers {
	return &handlers{
		service,
		logger,
		cfg,
		jwt,
	}
}