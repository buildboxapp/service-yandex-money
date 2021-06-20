package api

import (
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/yookassa/pkg/model"
	"time"
)

// обновление элемента (через вызов внешнего запроса)
func (p *api) AttrUpdate(uid, name, value, src, editor string) (err error) {

	post := map[string]string{}
	thisTime := fmt.Sprintf("%v", time.Now().UTC())

	post["uid"] = uid
	post["element"] = name
	post["value"] = value
	post["src"] = src
	post["rev"] = thisTime
	post["path"] = ""
	post["token"] = ""
	post["editor"] = editor

	dataJ, _ := json.Marshal(post)

	var objData model.Response
	_, err = p.utl.Curl("POST", p.cfg.Custom[0].ApiUrl+"/element/update?format=json", string(dataJ), &objData, map[string]string{})
	if err != nil {
		return
	}

	var dataObjs model.ResponseData
	b1, _ := json.Marshal(objData)
	json.Unmarshal(b1, &dataObjs)

	/////////////////   ОБРАБОТКА ТРИГГЕРА НА ИЗМЕНИЕ ОБЪЕКТА (ПОСТ)   /////////////////
	//go TriggerRun(dataObjs.Data, nil, "get", "after", "")
	/////////////////////////////////////////////////////////////////////////////////

	_, err = json.Marshal(objData)

	return
}

func (p *api) CreateObjForm(data map[string]string) (res model.ResponseData, err error) {
	dataJ, _ := json.Marshal(data)
	_, err = p.utl.Curl("POST", p.cfg.Custom[0].ApiUrl+"/objs?format=json", string(dataJ), &res, map[string]string{})

	if len(res.Data) == 0 {
		err = fmt.Errorf("%s", res.Status.Description)
	}

	return res, err
}
