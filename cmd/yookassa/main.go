package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/lib"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/metric"
	"github.com/buildboxapp/lib/config"
	"github.com/buildboxapp/yookassa/pkg/model"
	"strings"

	"github.com/buildboxapp/yookassa/pkg/api"
	"github.com/buildboxapp/yookassa/pkg/servers"
	"github.com/buildboxapp/yookassa/pkg/servers/httpserver"
	"github.com/buildboxapp/yookassa/pkg/service"
	"github.com/buildboxapp/yookassa/pkg/utils"

	"github.com/labstack/gommon/color"
	"os"
	"os/signal"
	"runtime/debug"

	"io"
)

const sep = string(os.PathSeparator)

var fileLog *os.File
var outpurLog io.Writer


func main()  {
	lib.RunServiceFuncCLI(Start)

	// закрываем файл с логами
	defer fileLog.Close()
}

// стартуем сервис приложения
func Start(configfile, dir, port, mode string) {
	var cfg model.Config
	done := color.Green("[OK]")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// инициируем пакеты
	err := config.Load(configfile, &cfg)
	if err != nil {
		fmt.Printf("%s (%s)", "Error. Load config is failed.", err)
		return
	}
	// ручная обработка конфигураций
	json.Unmarshal([]byte(cfg.Configuration), &cfg.Custom)

	cfg.UidService = strings.Split(configfile, ".")[0]

	///////////////// ЛОГИРОВАНИЕ //////////////////
	// формирование пути к лог-файлам и метрикам
	if cfg.LogsDir == "" {
		cfg.LogsDir = "logs"
	}
	// если путь указан относительно / значит задан абсолютный путь, иначе в директории
	if cfg.LogsDir[:1] != sep {
		rootDir, _ := lib.RootDir()
		cfg.LogsDir = rootDir + sep + "upload" + sep + cfg.Domain + sep + cfg.LogsDir
	}
	cfg.UrlProxy	= cfg.AddressProxyPointsrc

	// инициализировать лог и его ротацию
	var logger = log.New(
		cfg.LogsDir,
		cfg.ServiceLevelLogsPointsrc,
		lib.UUID(),
		cfg.Domain,
		"yookassa",
		cfg.UidService,
		cfg.LogIntervalReload.Value,
		cfg.LogIntervalClearFiles.Value,
		cfg.LogPeriodSaveFiles,
	)
	logger.RotateInit(ctx)

	fmt.Printf("%s Enabled logs. Level:%s, Dir:%s\n", done, cfg.ServiceLevelLogsPointsrc, cfg.LogsDir)
	logger.Info("Запускаем сервис: ",cfg.Domain)

	// создаем метрики
	metrics := metric.New(
		ctx,
		logger,
		cfg.LogIntervalMetric.Value,
	)

	defer func() {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			logger.Panic(fmt.Errorf("%s", b), "Recover panic from main function.")
			cancel()
			os.Exit(1)
		}
	}()

	utl := utils.New(
		cfg,
		logger,
	)
	port = utl.AddressProxy()
	cfg.PortService = port

	// клиент к api
	api := api.New(
		ctx,
		cfg,
		utl,
		logger,
	)

	// собираем сервис
	src := service.New(
		logger,
		cfg,
		metrics,
		utl,
		api,
	)

	// httpserver
	httpserver := httpserver.New(
		ctx,
		cfg,
		src,
		metrics,
		logger,
	)

	// для завершения сервиса ждем сигнал в процесс
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Kill)
	go ListenForShutdown(ch)

	srv := servers.New(
		"http",
		cfg,
		metrics,
		httpserver,
		src,
	)
	srv.Run()
}

func ListenForShutdown(ch <- chan os.Signal)  {
	<- ch
	os.Exit(0)
}
