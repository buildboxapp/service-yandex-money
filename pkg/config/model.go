package config

// Custom configuration Pay from service (parse TOML-format from field 'configuration')
type Pay struct {

	Shopid         					string         	`envconfig:"SHOPID" default:""`
	Shopkey         				string         	`envconfig:"SHOPKEY" default:""`
	Redirect         				string         	`envconfig:"PAY_REDIRECT" default:""`
	TplOrders         				string         	`envconfig:"PAY_TPL_ORDERS" default:""`
	ErrorRedirect         			string         	`envconfig:"PAY_ERROR_REDIRECT" default:"list/page/errorpay"`
	MoneyGate         				string         	`envconfig:"MONEY_GATE" default:"https://payment.yandex.net/api/v3/payments"`
	YandexRedirecturl				string			`envconfig:"YANDEX_REDIRECTURL" default:""`

	Ordertpl						string			`envconfig:"ORDERTPL" default:"" description:"uid-шаблона где размещаются платежи"`
}

type Config struct {

	ProjectKey         				string         	`envconfig:"PROJECT_KEY" default:"LKHlhb899Y09olUi"`

	ClientPath         				string         	`envconfig:"CLIENT_PATH" default:""`
	UrlProxy						string         	`envconfig:"URL_PROXY" default:""`
	UrlGui         					string         	`envconfig:"URL_GUI" default:""`
	UrlApi							string         	`envconfig:"URL_API" default:""`
	UidService         				string         	`envconfig:"UID_SERVICE" default:""`

	// Config
	ConfigName         				string         	`envconfig:"CONFIG_NAME" default:""`
	RootDir         				string         	`envconfig:"ROOT_DIR" default:""`

	// Logger
	LogsDir         				string         	`envconfig:"LOGS_DIR" default:"logs"`
	LogsLevel         				string         	`envconfig:"LOGS_LEVEL" default:""`
	LogIntervalReload         		Duration  		`envconfig:"LOG_INTERVAL_RELOAD" default:"10m" description:"интервал проверки необходимости пересозданния нового файла"`
	LogIntervalClearFiles         	Duration  		`envconfig:"LOG_INTERVAL_CLEAR_FILES" default:"30m" description:"интервал проверка на необходимость очистки старых логов"`
	LogPeriodSaveFiles         		string  		`envconfig:"LOG_PERION_SAVE_FILES" default:"0-1-0" description:"период хранения логов"`
	LogIntervalMetric         		Duration  		`envconfig:"LOG_INTERVAL_METRIC" default:"10s" description:"период сохранения метрик в файл логирования"`

	// Http
	MaxRequestBodySize 				Int       		`envconfig:"MAX_REQUEST_BODY_SIZE" default:"10485760"`
	ReadTimeout        				Duration 		`envconnfig:"READ_TIMEOUT" default:"10s"`
	WriteTimeout        			Duration 		`envconnfig:"WRITE_TIMEOUT" default:"10s"`
	ReadBufferSize     				Int    			`envconfig:"READ_BUFFER_SIZE" default:"16384"`



	// Params from .cfg
	SlashDatecreate	string `envconfig:"SLASH_DATECREATE" default:""`
	SlashOwnerPointsrc	string `envconfig:"SLASH_OWNER_POINTSRC" default:""`
	SlashOwnerPointvalue	string `envconfig:"SLASH_OWNER_POINTVALUE" default:""`
	SlashTitle	string `envconfig:"SLASH_TITLE" default:""`

	AccessAdminPointsrc	string `envconfig:"ACCESS_ADMIN_POINTSRC" default:""`
	AccessAdminPointvalue	string `envconfig:"ACCESS_ADMIN_POINTVALUE" default:""`
	AccessDeletePointsrc	string `envconfig:"ACCESS_DELETE_POINTSRC" default:""`
	AccessDeletePointvalue	string `envconfig:"ACCESS_DELETE_POINTVALUE" default:""`
	AccessReadPointsrc	string `envconfig:"ACCESS_READ_POINTSRC" default:""`
	AccessReadPointvalue	string `envconfig:"ACCESS_READ_POINTVALUE" default:""`
	AccessWritePointsrc	string `envconfig:"ACCESS_WRITE_POINTSRC" default:""`
	AccessWritePointvalue	string `envconfig:"ACCESS_WRITE_POINTVALUE" default:""`
	AddressProxyPointsrc	string `envconfig:"ADDRESS_PROXY_POINTSRC" default:""`
	AddressProxyPointvalue	string `envconfig:"ADDRESS_PROXY_POINTVALUE" default:""`

	Checkrun	string `envconfig:"CHECKRUN" default:""`
	CheckServiceext	string `envconfig:"CHECK_SERVICEEXT" default:""`
	Pay		Pay `envconfig:"PAY" default:""`
	Configuration	string `envconfig:"CONFIGURATION" default:""`

	DataUid	string `envconfig:"DATA_UID" default:""`
	Domain	string `envconfig:"DOMAIN" default:""`
	Driver	string `envconfig:"DRIVER" default:""`

	Pathrun	string `envconfig:"PATHRUN" default:""`
	PortAutoInterval	string `envconfig:"PORT_AUTO_INTERVAL" default:""`
	PortService	string `envconfig:"PORT_SERVICE" default:""`
	Projectuid	string `envconfig:"PROJECTUID" default:""`
	ProjectPointsrc	string `envconfig:"PROJECT_POINTSRC" default:""`
	ProjectPointvalue	string `envconfig:"PROJECT_POINTVALUE" default:""`

	ServiceExec	string `envconfig:"SERVICE_EXEC" default:""`
	ServiceLevelLogsPointsrc	string `envconfig:"SERVICE_LEVEL_LOGS_POINTSRC" default:""`
	ServiceLevelLogsPointvalue	string `envconfig:"SERVICE_LEVEL_LOGS_POINTVALUE" default:""`
	ServiceLogs	string `envconfig:"SERVICE_LOGS" default:""`
	ServiceMetricInterval	string `envconfig:"SERVICE_METRIC_INTERVAL" default:""`
	ServiveLevelLogsPointsrc	string `envconfig:"SERVIVE_LEVEL_LOGS_POINTSRC" default:""`
	ServiveLevelLogsPointvalue	string `envconfig:"SERVIVE_LEVEL_LOGS_POINTVALUE" default:""`

	Title	string `envconfig:"TITLE" default:""`
	ToBuild	string `envconfig:"TO_BUILD" default:""`
	ToUpdate	string `envconfig:"TO_UPDATE" default:""`

	Workingdir	string `envconfig:"WORKINGDIR" default:""`
}