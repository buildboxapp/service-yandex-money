package main

import (
	"github.com/buildboxapp/logger"
	"html/template"
	"strings"
	"net/url"
	"time"
	"context"
	"sync"
	"net/http"
)


// Для НАВИГАТОРА
type Bottons struct {
	Label      string
	Class      string
	Icon       string
	DataToggle string
	DataPlace  string
	DataPlace2 string
	DataUrl    string
	DataUrl2   string
}

type Navigator struct {
	Title		string
	User,UserID,Photo string
	Structure   []*Items
	Buttons     []Bottons
	Roles       []Data
	CurrentRole Data
	ButtonsNavTop []Data
	Selected    string
	PayTo		string
	PayTitle	string
	Mode    	string
	UpdateData	[]Data
	UpdateFlag	bool
	BaseMode	map[string]string
	ClientPath  string
	Domain		string
	ApiPath     string
	GuiPath		string
	Maket		string
	RequestURI  string
	Referer		string
	Social		template.HTML
	CountLicense 	int
	CountSessions	int
	AccessLicense	string
	UrlLogin	string
	State 		map[string]string
}



var StatusCode = RStatus{
	"OK":                       {"Запрос выполнен", 200, "", ""},
	"OKLicenseActivation":      {"Лицензия была активирована", 200, "", ""},
	"NotStatus":                {"Ответ сервера не содержит статус выполнения запроса", 501, "", ""},
	"NotExtended":              {"На сервере отсутствует расширение, которое желает использовать клиент", 501, "", ""},
	"ErrorCheckupIdForAll":     {"Ошибка: Идентификатор уже используется. Повторите ввод.", 500, "", ""},
	"ErrorCheckupFieldForAll":  {"Ошибка: Значение поля не уникально для всей БД. Повторите ввод.", 500, "", ""},
	"ErrorCheckupFieldForTpl":  {"Ошибка: Знаение поля не уникально среди полей данного шаблона. Повторите ввод.", 500, "", ""},
	"ErrorFormatJson":          {"Ошибка формата JSON-запроса", 500, "", ""},
	"ErrorTransactionFalse":    {"Ошибка выполнения тразакции SQL", 500, "", ""},
	"ErrorBeginDB":             {"Ошибка подключения к БД", 500, "", ""},
	"ErrorPrepareSQL":          {"Ошибка подготовки запроса SQL", 500, "", ""},
	"ErrorNullParameter":       {"Ошибка! Не передан параметр", 503, "", ""},
	"ErrorQuery":               {"Ошибка запроса на выборку данных", 500, "", ""},
	"ErrorScanRows":            {"Ошибка переноса данных из запроса в объект", 500, "", ""},
	"ErrorNullFields":          {"Не все поля заполнены", 500, "", ""},
	"ErrorAccessType":          {"Ошибка доступа к элементу типа", 500, "", ""},
	"ErrorGetData":             {"Ошибка доступа данным объекта", 500, "", ""},
	"ErrorRevElement":          {"Значение было изменено ранее.", 409, "", ""},
	"ErrorForbiddenElement":    {"Значение занято другим пользователем.", 403, "", ""},
	"ErrorUnprocessableEntity": {"Необрабатываемый экземпляр", 422, "", ""},
	"ErrorNotFound":            {"Значение не найдено", 404, "", ""},
	"ErrorReadConfigDir":       {"Ошибка чтения директории конфигураций", 403, "", ""},
	"errorOpenConfigDir":       {"Ошибка открытия директории конфигураций", 403, "", ""},
	"ErrorReadConfigFile":      {"Ошибка чтения файла конфигураций", 403, "", ""},
	"ErrorPortBusy":            {"Указанный порт занят", 403, "", ""},
	"ErrorGone":            	{"Объект был удален ранее", 410, "", ""},
	"ErrorShema":            	{"Ошибка формата заданной схемы формирования запроса", 410, "", ""},
	"ErrorUpdateParams":        {"Не переданы параметры для обновления серверов (сервер источник, сервер получатель)", 410, "", ""},

	"ErrorBaseTarget":          {"Не задана целевая база для обновления данных", 410, "", ""},
	"ErrorBaseSource":          {"Не задана источник обновления данных", 410, "", ""},

	"ErrorYandexNotProduct":    {"Не передан идентификатор продукта", 500, "", ""},
}


type RStatus map[string]RestStatus

type RestStatus struct {
	Description string `json:"description"`
	Status      int    `json:"status"`
	Error       string  `json:"error"`
	Code        string `json:"code"`
}

// ключ - UUID для sessionID
type Session struct {
	sync.Mutex
	Data map[string]Seance
	Logger  *logger.Log
}

type Seance struct {
	sync.Mutex
	Hash 		string 			`json:"hash"`
	Time 		time.Time 		`json:"time"`
	Context 	context.Context `json:"context"`
}

// элемент конфигурации
type Element struct {
	Type string
	Source interface{}
}

// ключ - MD5 от Uid пользователя (он постоянный)
// нужно, чтобы иметь возможность одному пользователю с разных браузеров иметь доступ к одному профилю
type Profile struct {
	Data map[string]*ProfileData
}

type ProfileData struct {
	Hash       		string
	Email       	string
	Uid         	string
	First_name  	string
	Last_name   	string
	Photo       	string
	Age       		string
	City        	string
	Country     	string
	Status 			string 	// - src поля Status в профиле (иногда необходимо для доп.фильтрации)
	Raw	       		[]Data	// объект пользователя (нужен при сборки проекта для данного юзера при добавлении прав на базу)
	Tables      	[]Data
	Roles       	[]Data
	Profiles       	[]Data
	Homepage		string
	Maket			string
	Groups			string
	GroupsName		string
	UpdateFlag 		bool
	UpdateData 		[]Data
	ButtonsNavTop	[]Data
	CurrentRole 	Data
	CurrentProfile 	Data
	Navigator   	[]*Items
	CountLicense	int
	BaseMode		map[string]string
}


type Page struct {
	Title        string
	Logo		 string
	FontColor 	 string
	Background 	 string
	Description	 string
	Year         string
	Maket		 string
	Navigator    template.HTML
	Napper       template.HTML
	Content      template.HTML
	Right        template.HTML
	Addons 		 template.HTML
	Rightsidebar template.HTML
	Path         template.HTML
	Tab          template.HTML
	Footer       template.HTML
	Metric		 template.HTML
	Social		 template.HTML
	Politic		 template.HTML
}

type Request struct {
	Data []interface{} `json:"data"`
}

type Response struct {
	Data   	[]interface{} 	`json:"data"`
	Res   	interface{} 	`json:"res"`
	Status 	RestStatus    	`json:"status"`
	Metrics Metrics 		`json:"metrics"`
}

type ResponseAPI struct {
	Data   	interface{} 	`json:"data"`
	Res   	interface{} 	`json:"res"`
	Status 	RestStatus    	`json:"status"`
	Metrics Metrics 		`json:"metrics"`
}

type Metrics struct {
	ResultSize     	int `json:"result_size"`
	ResultCount     int `json:"result_count"`
	ResultOffset    int `json:"result_offset"`
	ResultLimit     int `json:"result_limit"`
	ResultPage 		int `json:"result_page"`
	TimeExecution   string `json:"time_execution"`
	TimeQuery   	string `json:"time_query"`
}


type ResponseData struct {
	Data      []Data        `json:"data"`
	Res   	  interface{} 	`json:"res"`
	Status    RestStatus    `json:"status"`
	Metrics   Metrics 		`json:"metrics"`
}

// расчет/вставка функциональных полей
func (p *ResponseData) Сalculate() {
	for k, v := range p.Data {
		attr := v.Attributes
		for k1, v1 := range attr {
			//if strings.Contains(v1.Src, "roleid") {
			//	p.Data[k].Attributes[k1] = Attribute{"", User.CurrentRole.Uid, "", "", ""}
			//}
			//if strings.Contains(v1.Src, "userid") {
			//	p.Data[k].Attributes[k1] = Attribute{"", User.Uid, "", "", ""}
			//}
			if strings.Contains(v1.Src, "uid") {
				p.Data[k].Attributes[k1] = Attribute{"", v.Uid, "", "", "", ""}
			}
			if strings.Contains(v1.Src, "tplid") {
				p.Data[k].Attributes[k1] = Attribute{"", v.Source, "", "", "", ""}
			}

		}
	}
}

// ------------------------------------------
// ------------------------------------------
// ------------------------------------------
// ------------------------------------------


type Attribute struct {
	Value  string `json:"value",reindex:"value"`
	Src    string `json:"src",reindex:"src"`
	Tpls   string `json:"tpls",reindex:"tpls"`
	Status string `json:"status",reindex:"status"`
	Rev    string `json:"rev",reindex:"rev"`
	Editor string `json:"editor",reindex:"editor"`
}

type Data struct {
	Uid        		string               `json:"uid"`
	Id         		string               `json:"id"`
	Source     		string               `json:"source"`
	Parent     		string               `json:"parent"`
	Type       		string               `json:"type"`
	Title      		string               `json:"title"`
	Rev        		string               `json:"rev"`
	Сopies			string 				 `json:"copies"`
	Attributes 		map[string]Attribute `json:"attributes"`
}

// заменяем значение аттрибутов в объекте профиля
func (p *Data) AttrSet(name, element, value string) bool  {
	g := Attribute{}

	for k, v := range p.Attributes {
		if k == name {
			g = v
		}
	}

	switch element {
	case "src":
		g.Src = value
	case "value":
		g.Value = value
	case "tpls":
		g.Tpls = value
	case "rev":
		g.Rev = value
	case "status":
		g.Status = value
	}

	f := p.Attributes

	for k, _ := range f {
		if k == name {
			f[k] = g
			return true
		}
	}

	// если ранее аттрибута не было, значит добавим его
	p.Attributes[element] = g

	return true
}

// возвращаем необходимый значение атрибута для объекта если он есть, инае пусто
// а также из заголовка объекта
func (p *Data) Attr(name, element string) (result string, found bool) {

	if _, found := p.Attributes[name]; found {

		// фикс для тех объектов, на которых добавлено скрытое поле Uid
		if name == "uid" {
			return p.Uid, true
		}

		switch element {
		case "src":
			return p.Attributes[name].Src, true
		case "value":
			return p.Attributes[name].Value, true
		case "tpls":
			return p.Attributes[name].Tpls, true
		case "rev":
			return p.Attributes[name].Rev, true
		case "status":
			return p.Attributes[name].Status, true
		case "uid":
			return p.Uid, true
		case "source":
			return p.Source, true
		case "id":
			return p.Id, true
		case "title":
			return p.Title, true
		case "type":
			return p.Type, true
		}
	} else {
		switch name {
		case "uid":
			return p.Uid, true
		case "source":
			return p.Source, true
		case "id":
			return p.Id, true
		case "title":
			return p.Title, true
		case "type":
			return p.Type, true
		}
	}
	return "", false
}

type Applications struct {
	Tpl    map[string]Data
	Obj    map[string]map[string]Data
	ObjUid map[string]map[string]Data
}

type Modal struct {
	Token             string
	Title             string
	TitleTpl		  string
	TitleTplCreate	  string
	HeaderDescription string
	HeaderIcon        string
	HeaderClose       string
	HeaderTitle       string
	Cols              interface{}
	Tpl               interface{}
	Elements          []interface{}

	Data              interface{}
	Pages             interface{}
	ButtonsA          interface{}
	ButtonsP          interface{}

	Configuration	  interface{}
	Profile			  interface{}
	Cookie			  interface{}

	RequestURI		  interface{}
	Referer			  interface{}
	RequestValue	  url.Values

	Request	  		  *http.Request

	Value			  map[string]interface{}

	Rand              string
	RandUid			  string
	Uid               string
	Id                string
	Source            string
	TypeForm          string
	Parent            string
	Copies            string
	ButtonMode        string
	HideCreateButton  string
	View              string
	Viewers           []map[string]string
	ClientPath        string
	Domain			  string
	ApiPath     	  string

	CountLicense 	  int
	CountSessions	  int
	AccessLicense	  string

	Message           string
	CurrentRole 	  Data

	TplFiles		  string

	Metric		 	 template.HTML
	Social			 template.HTML
	Politic			 template.HTML
	SystemFields	 string
	PointFields		 string
	State 			 map[string]string

}

// ------------------------------------------
// ------------------------------------------
// ------------------------------------------
// ------------------------------------------

// для сложных запросов, объединяющих несколько
//type Queryes struct {
//	Filters   []Filter       `json:"querys"`
//	Shema     string         `json:"shema"`
//	Attribute QueryAttribute `json:"attribite"`
//}

// ДЛЯ ЗАПРОСОВ
type Query struct {
	Id        string         	`json:"id"`
	Tpl       string 		 	`json:"tpl"`
	Filters   []Filter        	`json:"filters"`
	Shema     Shema		 		`json:"shema"`
	Attribute QueryAttribute 	`json:"attribute"`
	Request   url.Values 	`json:"request"`
}


type Shema struct {
	Value     string 		 `json:"value"`
}

//type Filter struct {
//	Name      	string `json:"name"`
//	SubQuery   	string `json:"sub_query"`
//	Src      	string `json:"src"`
//	Value     	string `json:"value"`
//	Element     string 	`json:"element"`
//	SrcUrl      string `json:"src_url"`
//	ValueUrl	string `json:"value_url"`
//	Dynamic 	string `json:"dynamic"`
//}

type Filter struct {
	Tpls      	string `json:"tpls"`
	Name      	string `json:"name"`
	Src      	string `json:"src"`
	Value     	string `json:"value"`
	Element     string 	`json:"element"`
	Uid     	string 	`json:"uid"`
	Mode 		string `json:"mode"`
	Type 		string `json:"type"`
}

type QueryAttribute struct {
	Sort      		string `json:"sort"`
	Format	  		string `json:"format"`
	Order     		string `json:"order"`
	Limit     		string `json:"limit"`
	Page      		string `json:"page"`
	LinkId      	string `json:"linkid"`
	LinkObj      	string `json:"linkobj"`
	Shortmode 		string `json:"shortmode"`
	ExtQuery  		string `json:"extQuery"`
	TypeExtQuery  	string `json:"typeExtQuery"`

	Groups			string `json:"groups"`

	Fields       	string 		 `json:"fields"`
	FieldsFiltermode string `json:"fields_filtermode"`
	FieldsListservices string `json:"fields_listservices"`
	FieldsListaccess string	`json:"fields_listaccess"`
	FieldsTitles 	string `json:"fields_titles"`
}


type Items struct {
	Title  			string   	`json:"title"`
	ExtentedLink 	string 		`json:"extentedLink"`
	Uid    			string   	`json:"uid"`
	Source 			string   	`json:"source"`
	Icon   			string   	`json:"icon"`
	Leader 			string   	`json:"leader"`
	Order  			string   	`json:"order"`
	Type   			string   	`json:"type"`
	Preview			string   	`json:"preview"`
	Url    			string   	`json:"url"`
	Sub    			[]string 	`json:"sub"`
	Incl   			[]*Items 	`json:"incl"`
	Class 			string 		`json:"class"`
	FinderMode 		string 		`json:"finder_mode"`
}

type Widget struct {
	Template string                       `json:"template"`
	Rand     string                       `json:"rand"`
	Elements map[string]map[string]string `json:"elements"`
	Data     interface{}                  `json:"data"`
}

type Message struct {
	Message 	string `json:"message"`
	From		string `json:"from"`
	To 			string `json:"to"`
	Transport 	string `json:"transport"`
}

var MessageTemplate = map[string]map[string]string{
	"RU":{
		"LimitSession":"Внимание, достигнут лимит сессий пользователей.",
	},
}
