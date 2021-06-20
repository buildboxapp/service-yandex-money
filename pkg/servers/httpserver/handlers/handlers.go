package handlers

import (
	"context"
	"encoding/json"
	bblog "github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/yookassa/pkg/model"
	"github.com/buildboxapp/yookassa/pkg/service"
	"net/http"
)

type handlers struct {
	service service.Service
	logger bblog.Log
	cfg model.Config
	ctx *context.Context
}

type Handlers interface {
	Alive(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	Pay(w http.ResponseWriter, r *http.Request)
	Confirmation(w http.ResponseWriter, r *http.Request)
}

func (h *handlers) transportResponse(w http.ResponseWriter, response interface{}) (err error)  {
	d, err := json.Marshal(response)
	var statusCode = 200
	if err != nil {
		statusCode = 403
	}
	w.WriteHeader(statusCode)
	w.Write(d)
	return err
}

func (h *handlers) transportError(w http.ResponseWriter, code int, error error, message string) (err error)  {
	var res = model.Response{}

	res.Status.Error = error
	res.Status.Description = message
	d, err := json.Marshal(res)

	h.logger.Error(err, message)

	w.WriteHeader(code)
	w.Write(d)
	return err
}

func New(
	service service.Service,
	logger bblog.Log,
	cfg model.Config,
	ctx *context.Context,
) Handlers {
	return &handlers{
		service,
		logger,
		cfg,
		ctx,
	}
}