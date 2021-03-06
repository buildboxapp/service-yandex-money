// запускаем указанные виды из поддерживаемых серверов
package servers

import (
	"github.com/buildboxapp/service-yandex-money/pkg/config"
	"github.com/buildboxapp/service-yandex-money/pkg/servers/httpserver"
	"github.com/buildboxapp/service-yandex-money/pkg/service"
	bbmetric "github.com/buildboxapp/lib/metric"
	"strings"
)

type servers struct {
	mode string
	cfg config.Config
	metrics bbmetric.ServiceMetric
	httpserver httpserver.Server
	service service.Service
}

type Servers interface {
	Run()
}

// запускаем указанные севрера
func (s *servers) Run() {
	if strings.Contains(s.mode, "http") {
		s.httpserver.Run()
	}
}

func New(
	mode string,
	cfg config.Config,
	metrics bbmetric.ServiceMetric,
	httpserver httpserver.Server,
service service.Service,
) Servers {
	return &servers{
		mode,
		cfg,
		metrics,
		httpserver,
		service,
	}
}