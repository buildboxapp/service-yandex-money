// системный файл содержит фукнции, которые отвечают за интеграцию с инфраструктуру buildbox
package main

import (
	"encoding/json"
	"fmt"
	bblib "github.com/buildboxapp/lib"
	"github.com/buildboxapp/logger"
	"github.com/gorilla/mux"
	"github.com/labstack/gommon/color"
	"github.com/urfave/cli"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	stdlog "github.com/labstack/gommon/log"

	"io"
)

// тип ответа, который сервис отдает прокси при периодическом опросе (ping-е)
type Pong struct {
	Name string
	Version string
	Port int
	Pid  string
	State string
}
// этот тип нужен для передачи конфигурации State выполнения в обработчики
// и объекта Logger для логирования (без глобальных переменных)
type Service struct {
	Logger  *logger.Log
	State 	map[string]string
}
type Routes []Route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var fileLog *os.File
var outpurLog io.Writer

var log = logger.Log{}
var lib = bblib.Lib{}
var srv = Service{}

// инициализация логирования и стейта сервиса
func init() {
	// создаем/открываем файл логирования и назначаем его логеру
	fileLog, err := os.OpenFile("log_secvices.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		return
	}
	log.Init(fileLog, "All", "", "", "service")

	// задаем настройки логирования выполнения функций библиотеки
	lib.Logger = &log
	srv.Logger = &log
}

func main()  {

	// закрываем файл с логами
	defer fileLog.Close()

	defaultConfig, err := lib.DefaultConfig()
	if err != nil {
		lib.Logger.Warning("Warning! The default configuration directory was not found.")
	}

	appCLI := cli.NewApp()
	appCLI.Usage = "Demon Buildbox Proxy started"
	appCLI.Commands = []cli.Command{
		{
			Name:"run", ShortName: "r",
			Usage: "Run demon service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:	"config, c",
					Usage:	"Название файла конфигурации, с которым будет запущен сервис",
					Value:	defaultConfig,
				},
				cli.StringFlag{
					Name:	"dir, d",
					Usage:	"Путь к шаблонам",
					Value:	lib.RootDir(),
				},
				cli.StringFlag{
					Name:	"port, p",
					Usage:	"Порт, на котором запустить процесс",
					Value:	"",
				},
			},
			Action: func(c *cli.Context) error {
				configfile := c.String("config")
				dir := c.String("dir")
				lib.RunProcess(configfile, dir, "app", "start","services")

				return nil
			},
		},
		{
			Name:"start", ShortName: "",
			Usage: "Start single service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:	"config, c",
					Usage:	"Название файла конфигурации, с которым будет запущен сервис",
					Value:	defaultConfig,
				},
				cli.StringFlag{
					Name:	"dir, d",
					Usage:	"Путь к шаблонам",
					Value:	lib.RootDir(),
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
	router := NewRouter() //.StrictSlash(true)
	done := color.Green("OK")

	params, _, err := lib.ReadConf(configfile)
	if err != nil {
		lib.Logger.Error(err, "Reading file configuration is fail: "+configfile)
	}

	if port == "" || port == "0" {
		port = params["port_service"]
	}

	// автоматическая настройка портов
	addressProxy	:= params["address_proxy_pointsrc"]
	portInterval	:= params["port_auto_interval"]

	if (port == "" || port == "0") && addressProxy != "" && portInterval != "" {
		var portDataAPI bblib.Response
		// запрашиваем порт у указанного прокси-сервера
		proxy_url := addressProxy + "port?interval=" + portInterval
		lib.Curl("GET", proxy_url, "", &portDataAPI)
		port = fmt.Sprint(portDataAPI.Data)
	}

	params["port"] = port
	srv.State = params

	router.PathPrefix("/upload/").Handler(http.StripPrefix("/upload/", http.FileServer(http.Dir(dir + "/upload"))))
	router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(dir + "/templates"))))

	fmt.Printf("%s Starting service: %s\n", done, port)
	log.Info("Starting service: ", port)

	stdlog.Fatal(http.ListenAndServe(":"+port, router))
}

// ответ на запрос прокси
func (c *Service) ProxyPing(w http.ResponseWriter, r *http.Request) {

	pp := strings.Split(c.State["domain"] , "/")
	pg, _ := strconv.Atoi(c.State["port"])

	name := "ru"
	version := "ru"
	if len(pp) == 1 {
		name = pp[0]
	}
	if len(pp) == 2 {
		name = pp[0]
		version = pp[1]
	}

	pid := strconv.Itoa(os.Getpid())+":"+c.State["data-uid"]
	state, _ := json.Marshal(map[string]int{"cpu":0,"memory":0,"queue":0,"connection":0})

	var pong = []Pong{
		{name, version, pg, pid, string(state)},
	}

	// заменяем переменную домена на правильный формат для использования в Value.Prefix при генерации страницы
	c.State["domain"] 		= name + "/" + version
	c.State["client_path"] 	= "/" + c.State["domain"]

	res, _ := json.Marshal(pong)

	w.WriteHeader(200)
	w.Write([]byte(res))
}

// логирование запросов (активирование через параметр при старте -l --log console)
func (c *Service) LoggerHTTP(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		if name != "ProxyPing"  && false == true {
			c.Logger.Info(
				"Query: %s %s %s %s",
				r.Method,
				r.RequestURI,
				name,
				time.Since(start),
			)
		}
	})
}

// функция струтуры обработчиков
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = srv.LoggerHTTP(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
