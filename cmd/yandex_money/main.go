package main

import (
	"context"
	"fmt"
	"github.com/buildboxapp/lib"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/metric"
	"github.com/buildboxapp/service-yandex-money/pkg/api"
	"github.com/buildboxapp/service-yandex-money/pkg/config"
	"github.com/buildboxapp/service-yandex-money/pkg/jwt"
	"github.com/buildboxapp/service-yandex-money/pkg/servers"
	"github.com/buildboxapp/service-yandex-money/pkg/servers/httpserver"
	"github.com/buildboxapp/service-yandex-money/pkg/service"
	"github.com/buildboxapp/service-yandex-money/pkg/utils"
	"github.com/labstack/gommon/color"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/urfave/cli"

	"io"
)

const sep = string(os.PathSeparator)

var fileLog *os.File
var outpurLog io.Writer


func main()  {
	//warning := color.Red("[Fail]")

	// закрываем файл с логами
	defer fileLog.Close()
	defaultConfig, err := lib.DefaultConfig()

	if err != nil {
		return
	}
	rootDir, err := lib.RootDir()
	if err != nil {
		return
	}

	appCLI := cli.NewApp()
	appCLI.Usage = "Demon Buildbox Proxy started"
	appCLI.Commands = []cli.Command{
		{
			Name:"start", ShortName: "",
			Usage: "Start single Buildbox-service process",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:	"config, c",
					Usage:	"Название файла конфигурации, с которым будет запущен сервис",
					Value:	defaultConfig,
				},
				cli.StringFlag{
					Name:	"dir, d",
					Usage:	"Путь к шаблонам",
					Value:	rootDir,
				},
				cli.StringFlag{
					Name:	"port, p",
					Usage:	"Порт, на котором запустить процесс",
					Value:	"",
				},
			},
			Action: func(c *cli.Context) error {
				configfile := c.String("config")
				port := c.String("port")
				dir := c.String("dir")

				Start(configfile, dir, port)

				return nil
			},
		},
	}

	appCLI.Run(os.Args)

	return
}

// стартуем сервис приложения
func Start(configfile, dir, port string) {
	done := color.Green("[OK]")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// инициируем пакеты
	var cfg = config.New(configfile)

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
		cfg.LogsLevel,
		lib.UUID(),
		cfg.Domain,
		"service-yandex-money",
		cfg.UidService,
		cfg.LogIntervalReload.Value,
		cfg.LogIntervalClearFiles.Value,
		cfg.LogPeriodSaveFiles,
	)
	logger.RotateInit(ctx)

	fmt.Printf("%s Enabled logs. Level:%s, Dir:%s\n", done, cfg.LogsLevel, cfg.LogsDir)
	logger.Info("Запускаем app-сервис: ",cfg.Domain)

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

	jwt := jwt.New(
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
		jwt,
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
