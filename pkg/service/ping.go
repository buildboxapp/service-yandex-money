package service

import (
	"context"
	"encoding/json"
	"github.com/buildboxapp/yookassa/pkg/model"
	"os"
	"strconv"
	"strings"
)

// Ping ...
func (s *service) Ping(ctx context.Context) (result []model.Pong, err error) {
	pp := strings.Split(s.cfg.Domain, "/")
	name := "yookassa"
	version := "ru"

	if len(pp) == 1 {
		name = pp[0]
	}
	if len(pp) == 2 {
		name = pp[0]
		version = pp[1]
	}

	pg, _ := strconv.Atoi(s.cfg.PortService)
	pid := strconv.Itoa(os.Getpid()) + ":" + s.cfg.UidService
	state, _ := json.Marshal(s.metrics.Get())

	var r = []model.Pong{
		{name, version, "run", pg, pid, string(state), s.cfg.ReplicasService.Value},
	}

	return r, err
}
