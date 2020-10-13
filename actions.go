package main

import (
	"encoding/json"
	"fmt"
	liba "github.com/buildboxapp/app/lib"
	lib2 "github.com/buildboxapp/lib"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)


////////////////////////////////////////////////
//////////// ОПЕРАЦИИ С ОБЪЕКТАМИ //////////////
////////////////////////////////////////////////

// получение объекта
// params: fields - список полей, которые мы хотим получить
func (p *ResponseData) GetObj(objid string, params map[string]string) {
	lib2.Curl("GET", "_objs/"+objid, "", p)
	return
}

// получение объектов шаблона
// params: options - ОПИСАТЬ!!!!
func (p *ResponseData) GetObjFromTpl(objid string, params map[string]string) {
	option := ""
	if v, found := params["option"]; found {
		option = v
	}

	if option == "" {
		lib2.Curl("POST", "_data/"+objid, "", p)
	} else {
		lib2.Curl("POST", "_data/"+objid+"/"+option, "", p)
	}

	return
}
// перевод данных с формы в map slices
func FormToSlice(r *http.Request) map[string][]string {
	err := r.ParseForm()
	if err != nil {
		return nil
	}
	data := make(map[string][]string)
	for k, v := range r.PostForm {
		data[k] = v
	}
	return data
}


// создание объекта в запросе
// можем загружать как из:
// POST формы - отправлены поля (по-умолчанию, без доп. параметров)
// JSON-формат - получаем в теле поля, но в JSON-не (надо разобрать по-другому)
// FILE - загружаем из файла готовый объект
func CObjPost(w http.ResponseWriter, r *http.Request) {

	var data map[string]string
	var thisUser *ProfileData
	var _datecreate, _owner_pointsrc, _owner_pointvalue, _groups_pointsrc, _groups_pointvalue string
	var body []byte
	var err error

	// обработка запроса на создание в формате json
	// обычная мапка, где обязательно поле api_key = uid-приложения (потом сделать проверку)
	formatReq := false	// формат, когда приходят поля в json и их надо разобрать

	urlM := r.URL.Query()

	// если &file=путь_к_файлу_с_объектом
	// по загружаем этот файл и создаем объект
	if v, found := urlM["file"]; found {
		path := v[0]

		// 1 читаем файл
		res, err := lib.ReadFile(path)

		if err == nil {
			var objsData ResponseData
			json.Unmarshal([]byte(res), &objsData)

		// 2 создаем обьекты
			dataJ, _ := json.Marshal(CreateObj(objsData.Data))
			w.Write([]byte(dataJ))

			return
		}


	}


	if v, found := urlM["format"]; found {
		if v[0] == "json" {
			formatReq = true
		}

		body, err = ioutil.ReadAll(r.Body)
		defer r.Body.Close()
	}

	if err != nil {
		lib.ResponseJSON(w, body, "", err, nil)
		return
	}

	if formatReq {

		json.Unmarshal(body, &data)
		_owner_pointsrc = data["data-clientid"]
		_groups_pointsrc = data["data-groupid"]
		_groups_pointvalue = data["data-groupname"]


		// ДОДЕЛАТЬ - вставить проверку на наличие пользователя в базе
		if _owner_pointsrc == "" || len(_owner_pointsrc) < 20 {
			lib.ResponseJSON(w, _owner_pointsrc, "Unauthorized", err, nil)
			return
		}

		creator := data["data-clientname"]
		if creator == "" {
			creator = "user_" + _owner_pointsrc[:5]
		}
		_owner_pointvalue = creator

		if _groups_pointsrc != "" {
			data["_groups_pointsrc"] = _groups_pointsrc
			data["_groups_pointvalue"] = _groups_pointvalue
		}

	} else {
		// добавляем владельца объекта (кто его создал)
		thisUser = r.Context().Value("User").(*ProfileData)

		_datecreate = fmt.Sprintf("%v", time.Now().UTC())
		_owner_pointsrc = thisUser.Uid
		_owner_pointvalue = thisUser.First_name + " " + thisUser.Last_name
		data = FormToMap(r)

		data["_groups_pointsrc"] = thisUser.Groups
		data["_groups_pointvalue"] = thisUser.GroupsName
	}


	data["_datecreate"] = _datecreate
	data["_owner_pointsrc"] = _owner_pointsrc
	data["_owner_pointvalue"] = _owner_pointvalue


	// если получили значение data-type_pointsrc - значит тип задается создаваемого объекта создается вручнубю
	// значит задаем значение поля data-type
	_type_pointsrc := data["data-type_pointsrc"]
	if _type_pointsrc != "" {
		data["data-type"], _ = DivideSrc(_type_pointsrc)
	}



	newObj := GenerateObj(data, r)


	// передаем массив сформированных объектов (для группового создания)
	sliceData := []Data{}
	sliceData = append(sliceData, newObj)

	dataJ, _ := json.Marshal(CreateObj(sliceData))

	w.Write([]byte(dataJ))
}


// само создание объекта (JSON)
func CreateObj(sendData []Data) (objData ResponseData) {
	dataJ, _ := json.Marshal(sendData)

	// отправляем обьект в шаблон
	lib2.Curl("POST", "_objs", string(dataJ), &objData)

	// пост-обработка триггера (на создание)
	TriggerCreatePost(&objData)

	return objData
}


// само создание объекта, получая данные формы (старый вариант)
func CreateObjForm(data map[string]string, r *http.Request) (objData ResponseData) {

	newObj := GenerateObj(data, r)

	// передаем массив сформированных объектов (для группового создания)
	sliceData := []Data{}
	sliceData = append(sliceData, newObj)

	//fmt.Println("objData: ", objData)

	objData = CreateObj(sliceData)

	return objData
}


// генерация объекты из данных формы
func GenerateObj(data map[string]string, r *http.Request) Data {
	var sendData Data
	thisUserUID := ""

	// получаем текущего пользователя (кто создает)
	thisUser := r.Context().Value("User").(*ProfileData)
	if thisUser != nil {
		thisUserUID = thisUser.Uid
	}
	thisTime := fmt.Sprintf("%v", time.Now().UTC())

	sendData.Id = data["id"]
	sendData.Uid = data["data-uid"]
	sendData.Parent = data["data-parent"]
	sendData.Source = data["data-source"]
	sendData.Type = data["data-type"]
	sendData.Сopies = data["data-copies"]

	// если передан тип obj - то это объект, а у него в базе пустое значение
	if sendData.Type == "obj" {
		sendData.Type = ""
	}
	if sendData.Parent == "" {
		sendData.Parent = sendData.Source
	}


	// При создании элемента формы в Parent != Source поэтому добавляется data-type=element
	// если мы делаем просто объект (например кнопку), то Parent = Source и это не элемент формы
	// можно было добавлять тип element если Parent != Source, а не получать этот аттрибут с формы
	// но это менее гибко, ведь возможно создании одного объета для другого и это будет не элемент формы
	//if sendData.Parent == sendData.Source && sendData.Source != "" && sendData.Parent != "" {
	//	sendData.Type = ""
	//}


	attrData := make(map[string]Attribute)
	attrObj := &Attribute{}
	for k, v := range data {
		src := ""
		value := ""
		element := ""

		// если есть поля селектов, то добавляем их по-другому (scr и value)
		found_pointsrc := contains("true", k, "_pointsrc")
		found_pointvalue := contains("true", k, "_pointvalue")

		if found_pointsrc == "true" || found_pointvalue == "true" {
			if found_pointsrc == "true" {
				element = strings.Replace(k, "_pointsrc", "", -1) // название селект-элемента
			}
			if found_pointvalue == "true" {
				element = strings.Replace(k, "_pointvalue", "", -1) // название селект-элемента
			}
			src = data[element+"_pointsrc"]
			value = data[element+"_pointvalue"]
		} else {
			value = v
			element = k
		}
		// -----------------------------------------------------------------

		if mes := contains("true", k, "data-"); mes != "true" {
			attrObj.Value = value
			srcText, tplsText := DivideSrc(src)

			// Фиксим неполные данные связянного объекта
			// Бывает не указан tpl связанного объекта
			// запрашиваем его и добавляем

			if tplsText == "" && srcText != "" {
				var objObj ResponseData
				Curl("GET", "_objs/"+srcText, "", &objObj)
				if len(objObj.Data) != 0 {
					tplsText = objObj.Data[0].Source
				}
			}

			attrObj.Src = srcText
			attrObj.Tpls = tplsText
			attrObj.Status = ""
			attrObj.Rev = thisTime
			attrObj.Editor = thisUserUID


			attrData[element] = *attrObj
		}
		// -----------------------------------------------------------------

	}


	// при создании нового пользователя еще нет прав, но создается объект
	// в этом случае не выставляем права на объект
	if thisUser != nil {

	// добавляем аттрибуты прав доступа (при создании добавляем uid-роли пользователя)
	attrObjAccess := &Attribute{}
	attrObjAccess.Src = thisUser.CurrentRole.Uid

	attrObjAdmin := &Attribute{}
	attrObjAdmin.Src = thisUser.CurrentRole.Uid

	// если создается шаблон - добавляем системную роль ИБ в права управления
	if data["data-type"] == "tpl" {
		attrObjAdmin.Src = attrObjAdmin.Src + "," + RoleSecurity
	}

	//if data["data-type"] == "tpl" || data["data-type"] == "element" {
		attrData["access_read"] = *attrObjAccess
		attrData["access_write"] = *attrObjAccess
		attrData["access_delete"] = *attrObjAccess
		attrData["access_admin"] = *attrObjAdmin
	//}

	}

	sendData.Attributes = attrData

	// обработк @-функции через данные в самом объекте
	fconf, _ := json.Marshal(sendData)
	fconf1 := DogParse(string(fconf), r, &sendData)

	json.Unmarshal([]byte(fconf1), &sendData)

	return sendData
}


// обновление элемента (через вызов внешнего запроса)
func (p *Data) AttrUpdate(name, value, src, editor string) bool {

	post := map[string]string{}
	thisTime := fmt.Sprintf("%v", time.Now().UTC())

	post["uid"] = p.Uid
	post["element"] = name
	post["value"] = value
	post["src"] = src
	post["rev"] = thisTime
	post["path"] = ""
	post["token"] = ""
	post["editor"] = editor

	dataJ, _ := json.Marshal(post)

	var objData Response
	Curl("POST", "_element/update", string(dataJ), &objData)

	// пост-обработка триггера (на изменение)
	TriggerUpdatePost(p.Uid, &objData)

	_, err := json.Marshal(objData)

	if err == nil {
		return true

	} else {
		return false
	}
}


////////////////////////////////////////////////
//////////// ОПЕРАЦИИ С ЛИНКАМИ //////////////
////////////////////////////////////////////////

// получение связей /link/{obj}/{mode}
// параметры URL:	{obj} 	- объект, к которому запрашиваем связи
//					&mode 	- in/out/all (по-умолчанию out - исходищие) out - я ссылаюсь на объекты; in - объект ссылается на меня
//					&source - через , список шаблонов, где искать связанные объекты. если нет, то ищем везде
//					&short	- сокращенный вывод (только UID-s) значение если не пустое
func CLinkGet(w http.ResponseWriter, r *http.Request) {
	var objData ResponseAPI

	vars := mux.Vars(r)
	obj := vars["obj"]
	mode := r.FormValue("mode")
	source := r.FormValue("source")
	short := r.FormValue("short")


	if mode == "" {
		mode = "out"
	}
	Curl("GET", "_link?obj="+obj+"&source="+source+"&mode="+mode+"&short="+short, "", &objData)

	dataJ, _ := json.Marshal(objData)

	w.Write([]byte(dataJ))

	return
}


// добавляем связь /link/add
// POST параметры: 	element - поле, на которое добавляем связь (если поле пустое - добавляем автоматическое _links)
//					from 	- объекты, которые связываются
//					to 		- объекты, с которыми связываем
func CLinkAdd(w http.ResponseWriter, r *http.Request) {
	var objData ResponseAPI

	post := map[string]string{}

	post["element"] = r.FormValue("element")
	post["from"] = r.FormValue("from")
	post["to"] = r.FormValue("to")

	dataJ, _ := json.Marshal(post)

	Curl("JSONTOPOST", "_link/add", string(dataJ), &objData)

	return
}

// удаляем связь /link/delete
// POST параметры: 	element - поле, на которое добавляем связь (если поле пустое - добавляем автоматическое _links)
//					from 	- объекты, которые отвязываются
//					to 		- объекты, с которыми отвязываем
func CLinkDelete(w http.ResponseWriter, r *http.Request) {
	var objData ResponseAPI

	post := map[string]string{}

	post["element"] = r.FormValue("element")
	post["from"] = r.FormValue("from")
	post["to"] = r.FormValue("to")

	dataJ, _ := json.Marshal(post)

	Curl("JSONTOPOST", "_link/delete", string(dataJ), &objData)

	return
}


////////////////////////////////////////////////
//////////////////// УТИЛИТЫ ///////////////////
////////////////////////////////////////////////

// перевод данных с формы в map (одноуровневую - ключ-значение)
// Внимание! теряются дублирующие значения, только первое значение по ключу
func FormToMap(r *http.Request) map[string]string {
	err := r.ParseForm()
	if err != nil {
		return nil
	}
	data := make(map[string]string)
	for k, v := range r.PostForm {
		data[k] = v[0]
	}
	return data
}


////////////////////////////////////////////////
//////////////////// ЛИЦЕНЗИРОВАНЕ /////////////
////////////////////////////////////////////////

// сущности
// ClientKey - это зашифрованный массив из адресов серверов, на которых мы запрашиваем активацию
// и хранится в скрытом поле clientkey при создании проекта
// ProjectKey - это hash-ключ вшитый в сборку при его компиляции через параметр, если он пустой,
// то генерится как hash(buildbox.app)

// clientkey - Генерируется утилитой /tools (вручную добавляем ProjectKey и массив)
// и прописывается в шаблоне Проекты в скрытом поле clientkey
// первый адрес - хост, у которого была куплена лицензия и который ее подтверждает
// второй/третий адреса - хосты для отправки фоновых уведомлений о лицензировании через другой адрес
// Для чего это надо? Для того, чтобы в дальнейшем сделать механизм отзыва у СубЛицензиата права лицензировать даже
// лицензии, которые он продал - это может понадобиться в том случае, если СубЛицензиат решил сливать в сеть бесплатные лиц.


// как работает лицензирование?
// клиент получает от провайдера код  (в файле) лицензии - это обычный объект и загружает к себе в систему
// далее нажимает Активировать -> идет запрос на сервер, адрес которого зашит в clientkey сборки
// clientkey - это зашифрованный (ключом ProjectKey) массив из адресов серверов, на которых мы запрашиваем активацию и которые могут
// выдать данную лицензию. т.е. лицензия должна активироваться только на сервере, на котором была сгенерированна
// если мы загрузим самодельную лицезнию, то она не сможет быть активирована, поскольку нет возможности адресовать
// запрос активации на сторонний сервер, вместо прописанного в clientkey

// теоретически можно сделать просто запрос к "псевдо-серверу", который сгенерировал лицензию и получить от него код
// но это возможно лишь в том случае, если "злоумышленник" знает секретный ключ и сам сгенерировал клиента, поскольку
// ключи клиента и сервера должны совпадать, иначе клиент не сможет расшифровать зашифрованный ответ активации и не сможет
// его проверить на валидность.

// таким образом, сама уязвимость подмены сервера невозможна, поскольку сборка "сервера с лицензией" поставляется лично
// клиенту с зашитим в нее clientkey (где вторым адресом идет адрес buildbox.app - для возможности дублировать запросы,
// чтобы иметь статистику и контролировать партнеров, сборки которых раздают пустые лицензии.

// клиент делает запрос серверу на активацию лицензии и получает зашифрованный ProjectKey-ем слайс данных о текущей лицензии
// где uid-пользоватяля который активирует лицензию (мастер-юзер)
// uid-объекта лицензии
// uid-проекта, для которого генерируется лицензия (лицензия может быть активна только раз для одного проекта), теоретически
// лицензию можно отозвать и активировать на другой проект, но это очень трудоемко, ибо активированная лицензия больше не
// проверяется и может работать локально без доступа к Интернету, поэтому отзыв ее невозможен
// ну и дата протухания лицензии

// сервер при старте каждый раз расшифровывает все доступные и активные лицензии и исходя из этого назначает лимиты для
// количества одновременных сессий
// других ограничений пока нет, потому как сложно контролировать и распределять лицензии между пользователями.

// лицензия может быть активирована один раз. при повторном запросе активации возвращается ошибка и активация не происходит

// единственная угроза - декомпиляция и возможность узнать ProjectKey

// СТОРОНА КЛИЕНТА
// получаем запрос на активацию лицензии и делаем запрос на сервер активации
func  LicensePush(w http.ResponseWriter, r *http.Request) {
	var objLicense ResponseAPI
	var objLicenseClient ResponseData
	var sliceLicenseHost []string
	var bodyLicense map[string]string


	vars := mux.Vars(r)
	obj := vars["obj"]

	// 1. получаем проект и владельца лицензии (кто её активирует -> системный аккаунт клинета)
	project := r.FormValue("project")
	thisUser := r.Context().Value("User").(*ProfileData)

	// 2. получаем адрес сервера лицензирования из конфигурации
	licensekey := app.Get("licensekey")
	if licensekey == "" {
		lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. License key is failed."), nil)
		return
	}
	strlicenseHost, _ := lib.Decrypt(PK, licensekey)
	json.Unmarshal([]byte(strlicenseHost), &sliceLicenseHost)

	if len(sliceLicenseHost) == 0 {
		lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. License key is failed."), nil)
		return
	}

	// чистим от случайных /
	licenseHost := strings.Trim(sliceLicenseHost[0], "/")

	// 3. отправляю данные на сервер выдали лицензий
	ress, err := Curl("GET", licenseHost+"/license/activate/"+obj+"?project="+project+"&owner="+thisUser.Uid, "", &objLicense)
	if err != nil || ress == nil {
		lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. License server is not responding"), nil)
		return
	}

	if objLicense.Status.Error != "" {
		lib2.ResponseJSON(w, nil, objLicense.Status.Code, fmt.Errorf(objLicense.Status.Error), objLicense.Metrics)
		return
	}

	// 4. обновляем текущую лицензию (дописывают ЭЦП в лицензию на стороне клиента)
	_, err = Curl("GET", "_objs/"+obj, "", &objLicenseClient)
	if len(objLicenseClient.Data) == 0 {
		lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. License with this ID not found"), nil)
		return
	}

	objLicenseClient.Data[0].AttrUpdate("activation", fmt.Sprint(objLicense.Data), "", "")


	// расшифровываем полученную ЭЦП и
	// сохраняем дату оплаты (полученную через ЭЦП лицензии)
	mapLicense, err := lib.Decrypt(PK, fmt.Sprint(objLicense.Data))
	if err != nil {
		lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. License is failed"), nil)
		return
	}
	json.Unmarshal([]byte(mapLicense), &bodyLicense)
	datetime_expires 	:= bodyLicense["date_pay"]
	objLicenseClient.Data[0].AttrUpdate("datetime_expires", datetime_expires, "", "")


	lib2.ResponseJSON(w, nil, "OKLicenseActivation", err, nil)
}


// СТОРОНА СЕРВЕРА
// активация лицензии на стороне сервера
func  LicenseActivate(w http.ResponseWriter, r *http.Request)  {
	vars := mux.Vars(r)
	obj := vars["obj"]
	project := r.FormValue("project")
	owner := r.FormValue("owner")

	var objLicense, objProduct ResponseData
	var hashBody = map[string]string{}
	var countLicense int


	// 1 получаю данные объекта
	_, err := Curl("GET", "_objs/"+obj, "", &objLicense)
	if len(objLicense.Data) == 0 {
		lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. License with this ID not found"), nil)
		return
	}

		// 1.1 проверяем наличие подписи у запрашиваемой лицензии
		activation, found := objLicense.Data[0].Attr("activation", "value")
		if found && activation != "" {
			lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. This license has been activated before."), nil)
			return
		}

		// 1.2 проверяем наличие оплаты у запрашиваемой лицензии
		datetime_expires, found := objLicense.Data[0].Attr("datetime_expires", "value")
		if !found || datetime_expires == "" {
			lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. This license was not paid!"), nil)
			return
		}


	// 2 создаю объект для кеша
	hashBody["license"] = obj

		// 1 получаю данные объекта
		product_id, found := objLicense.Data[0].Attr("product", "src")

		_, err = Curl("GET", "_objs/"+product_id, "", &objProduct)
		if len(objProduct.Data) == 0 {
			lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. Linked product is not found"), nil)
			return
		}

		// количество лицензий
		formula_countLicense, found := objProduct.Data[0].Attr("option2", "value")
		if !found || formula_countLicense == "" {
			lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. the option number of licenses is set incorrectly"), nil)
			return
		}
		countLicense, err = strconv.Atoi(formula_countLicense)
		if err != nil {
			lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. the option number of licenses is set incorrectly"), nil)
			return
		}


		// формируем дату смерти
		formula_datedeath, found := objProduct.Data[0].Attr("option1", "value")
		if !found {
			lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. Time option is not found."), nil)
			return
		}
		f1 := strings.Split(formula_datedeath, ".")
		fd1, err := strconv.Atoi(f1[0])
		fd2, err := strconv.Atoi(f1[1])
		fd3, err := strconv.Atoi(f1[2])
		if err != nil {
			lib2.ResponseJSON(w, nil, "NotStatus", fmt.Errorf("Error. Time option have not true format."), nil)
			return
		}
		datedeath := time.Now().AddDate(fd1,fd2,fd3)


	hashBody["owner"] = owner
	hashBody["project"] = project
	hashBody["date_death"] = fmt.Sprint(datedeath)
	hashBody["date_pay"] = datetime_expires
	hashBody["count"] = fmt.Sprint(countLicense)


	jsonLicense, err := json.Marshal(hashBody)
	res, err := lib.Encrypt(PK, string(jsonLicense))

	// 3 дописывают ЭЦП в лицензию на стороне сервера
	objLicense.Data[0].AttrUpdate("activation", res, "", "")


	lib2.ResponseJSON(w, res, "OK", err, nil)
	//lib2.ResponseJSON(w, hashBody, "OK", err, nil)

}


// СТОРОНА КЛИЕНТА
// расчитываем количество активных лицензий
func LicenseCheck() (countActiveLicense int) {
	var m2, objUser ResponseData
	var bodyLicense map[string]string

	Curl("GET", "/query/sys_license_activated", "", &m2)
	for _, v := range m2.Data {

		activation_code, found := v.Attr("activation", "value")
		if !found || activation_code == "" {
			continue
		}

		mapLicense, err := lib.Decrypt(PK, activation_code)
		if err != nil {
			continue
		}
		json.Unmarshal([]byte(mapLicense), &bodyLicense)

		countStr 	:= bodyLicense["count"]
		project 	:= bodyLicense["project"]
		license 	:= bodyLicense["license"]
		date_pay 	:= bodyLicense["date_pay"]
		owner 		:= bodyLicense["owner"]

		// ПРОВЕРКИ ВАЛИДНОСТИ ЭЦП:

		// на дату окончания
		if liba.Timefresh(bodyLicense["date_death"]) == "false" {
			continue
		}
		// на наличие оплаты
		if date_pay == "" {
			continue
		}
		// на количество
		count, err := strconv.Atoi(countStr)
		if err != nil {
			continue
		}
		// на проект
		if project != app.State["projectUid"] {
			continue
		}
		// на принадлежность к данной лицензии
		// проверяем что ЭЦП принадлежит именно этой лицензии
		if license != v.Uid {
			continue
		}
		// проверка на наличие системного пользователя в системе (того, кто активировал)
		_, err = Curl("GET", "_objs/"+owner, "", &objUser)
		if len(objUser.Data) == 0 {
			continue
		}

		countActiveLicense += count
	}

	// бесплатные 5 лицензий
	if countActiveLicense < 5 {
		countActiveLicense = 5
	}

	return
}

// функция расчета статуса нагрузки выделелнных лицензиями сессий
// countLicense - количество доступных сессий пользователей
// countSessions - текущее количество сессий пользователей
// на выходе слово для класса плашки и редирект если превышение
// превышение до хх-% от максимального числа - норма, но надо информировать администратора (задается через глоб.переменную)
// четыре режима: зеленая/желтая/красная/блокировка
// пороги предупреждения:
// до 70% - зеленый
// 70-90% - желтый
// 90-100(+порог) - красный
// выше порога - редирект
func AccessCheck(w http.ResponseWriter, r *http.Request, countLicense, countSessions int) (color string) {

	var count = float64(countLicense)
	var stopLimit = ProcToleranceExcessLimitSession * count
	var sessions = float64(countSessions)
	var message = Message{MessageTemplate[Lang]["LimitSession"], "", "", "sms"}

	// выше порога - редирект
	if sessions > stopLimit {
		message.Send()
		url_signin := app.State["url_signin"]
		if url_signin != "" {
			http.Redirect(w, r, url_signin+"?mode=limit&ref="+ClientPath+r.RequestURI, 302)
		} else {
			http.Redirect(w, r, ClientPath+"/login?ref="+ClientPath+r.RequestURI+"&mode=limit", 302)
		}
	}

	// 90-100(+порог) - красный
	if (sessions <= stopLimit) && (sessions > 0.9*count) {
		return "danger"
	}

	// 70-90% - желтый
	if (sessions <= 0.9*count) && (sessions > 0.7*count) {
		return "warning"
	}

	// 100-70% - зеленый
	if (sessions < 0.7*count) {
		return "primary"
	}

	return ""
}


// фукнции КОРЗИНЫ
// получаем option - действие
// и ?uids - для штучного удаления - через , перечень uid-ов удаляемых обьектов
func ETrash(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	option := vars["option"] // может принимать значения: _elements (получить список объектов полей (type=element))
	uids := vars["uids"] // может принимать значения: _elements (получить список объектов полей (type=element))
	var objSource ResponseData

	// выводим список удаленных объектов
	if strings.ToUpper(option) == "GET" {
		lib2.Curl("GET", "_trash/get", "", &objSource)
	}

	if strings.ToUpper(option) == "DELETE" {
		lib2.Curl("DELETE", "_objs/"+uids, "", &objSource)
	}

	if strings.ToUpper(option) == "CLEAR" {
		lib2.Curl("POST", "_trash/clear", "", &objSource)
	}

	c, _ := json.Marshal(objSource)
	w.WriteHeader(200)
	w.Write(c)
}
