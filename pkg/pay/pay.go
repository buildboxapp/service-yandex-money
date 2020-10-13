// Платежный модуль Yandex.Kassa позволяет создавать платежи и учитывать факт оплаты
// получает данные из настроект проекта

// СОЗДАНИЕ СЧЕТА
// существует возможность перенаправление созданных платежей на стороннюю страницу
// может понадобиться при выставлении счета, когда счет выставляется на базе сформированного платежа для общего учета
// для редиректа необходимо указать доп.параметр &redirect_postcreate=/адрес_страницы_для_редиректа?obj=
// при редиректе будет добавлен uid сформированного платежа, на базе которого может быть сформирован счет
// пример вызова модуля с редиректорм на создание счета
// /buildbox/gui/tools/pay?product=2020-02-25T22-44-40z03-00-100241&redirect_postcreate=/view/page/orderb?obj=
// .... адрес страницы оплаты....................................../.. параметр редиректа ......место UID-а платежа


package pay

import (
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/services/yandex.pay"
	bblib "github.com/buildboxapp/lib"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Payment struct {
	Amount map[string]string `json:"amount"`
	Confirmation map[string]string `json:"confirmation"`
	Description string `json:"description"`
}

type AnswerGateRound struct {
	ID string `json:"id"`
	Status string `json:"status"`
	Paid bool `json:"paid"`
	CreatedAt   time.Time `json:"created_at"`
	Amount map[string]string `json:"amount"`
	Confirmation map[string]string
}

type AnswerConfirmation struct {
	Type   string `json:"type"`
	Event  string `json:"event"`
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Paid   bool   `json:"paid"`
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
		AuthorizationDetails struct {
			Rrn      string `json:"rrn"`
			AuthCode string `json:"auth_code"`
		} `json:"authorization_details"`
		CreatedAt   time.Time `json:"created_at"`
		Description string    `json:"description"`
		ExpiresAt   time.Time `json:"expires_at"`
		Metadata    struct {
		} `json:"metadata"`
		PaymentMethod struct {
			Type  string `json:"type"`
			ID    string `json:"id"`
			Saved bool   `json:"saved"`
			Card  struct {
				First6      string `json:"first6"`
				Last4       string `json:"last4"`
				ExpiryMonth string `json:"expiry_month"`
				ExpiryYear  string `json:"expiry_year"`
				CardType    string `json:"card_type"`
			} `json:"card"`
			Title string `json:"title"`
		} `json:"payment_method"`
		Test bool `json:"test"`
	} `json:"object"`
}




// Payment - платеж, который будет отправлен на платежный шлюз
// Order - объект счета, который создается в БД и где фиксируются атрибуты платежа
// Product - товар/услуга, которая оплачивается (только берем данные)


func GPayYandex(w http.ResponseWriter, r *http.Request) {
	var objProduct main.ResponseData
	redirect_postcreate := r.FormValue("redirect_postcreate") // редирект после создания объекта платежа (без оплаты в яндекс)
	product := r.FormValue("product")  // товар, за который оплачиваем

	bblib.Curl("GET", "_objs/"+product, "", &objProduct)


	// 0. формируем аттрибуты для платежа и счета
	amount, amount_string, currency, product_pointsrc, product_pointvalue := getProductAttr(objProduct)

	// 1. создаем платеж (формата шлюза)
	description := product_pointvalue
	//description := "Оплата товара:" + product_pointvalue + " на сумму " + strconv.Itoa(amount)
	payment, _ := setPayment(amount, description, currency)

	// 2. создаем объект счета
	OrderUID, OrderObj := createOrder(r, payment, product_pointsrc, product_pointvalue, amount_string)
	d, _ := json.Marshal(OrderUID)

	// 2.1. редиректим на страницу счета с указанием uid созданного платежа
	if redirect_postcreate != "" {
		http.Redirect(w, r, redirect_postcreate + OrderUID, 302)
	}

	// 3. отправляем запрос на бронирование платежа и перенаправление на страницу оплаты шлюза
	answer, err := postPayment(w, payment, OrderUID)

	//fmt.Println("Объект платежа, который пришел со шлуюза: ", answer)
	// если оплата провелась в один этап (ApplePay или тестовый), то проверяем на статус и
	// обновляем объект платежа в базе на выполенный (ставим дату списания)
	//if answer.Status == "succeeded" {
	//	datetime_expires := answer.CreatedAt
	//	DatatimeExpires := datetime_expires.Format("2006-01-02 15:04:05")
	//	OrderObj.Data[0].AttrUpdate("datetime_expires", DatatimeExpires, "", "")
	//}

	if err != nil {
		w.WriteHeader(503)
		w.Write([]byte(d))
	}

	// 4. сохраняю идентификатор платежа шлюза в объекте платеж
	OrderObj.Data[0].AttrUpdate("paymentID", answer.ID, "", "")

	redirect_url := answer.Confirmation["confirmation_url"]
	http.Redirect(w, r, redirect_url, 302)

	// получение формы оплаты и редирект должен произойти в postPayment
	http.Redirect(w, r, Redirect_errorpay, 302)

	// иначе пишем ошибку
	w.WriteHeader(503)
	w.Write([]byte(d))
}

// сюда приходит подтверждение оплаты от шлюза
// ручка, которую дергает шлюз
func GPayYandexСonfirmation(w http.ResponseWriter, r *http.Request)  {

	main.log.Info("Пришел запрос подтверждения от Yandex-Шлюза на оплату")
	fmt.Println("Пришел запрос подтверждения от Yandex-Шлюза на оплату ")


	var answerConfirmation AnswerConfirmation
	// отправляем запрос на бронирование платежа
	responseData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Errorf("Error: Parsing the answer Confirmation Gateway is not valid! (%s)", err)
		main.log.Error(err, "Error YandexPay: Parsing the answer Confirmation Gateway is not valid! ")
		return
	}

	err = json.Unmarshal(responseData, &answerConfirmation)
	if err != nil {
		fmt.Errorf("Error: Unmarshal the answer Confirmation Gateway is not valid! (%s)", err)
		main.log.Error(err, "Error YandexPay: Unmarshal the answer Confirmation Gateway is not valid ")
		return
	}

	PaymentID := answerConfirmation.Object.ID
	PaymentMethod := answerConfirmation.Object.PaymentMethod.Title
	PaymentStatus := answerConfirmation.Object.Status

	datetime_expires := answerConfirmation.Object.CreatedAt
	DatatimeExpires := datetime_expires.Format("2006-01-02 15:04:05")


	// ПРОБЕГАЮ ВСЕ ПЛАТЕЖИ И СМОТРЮ СОВПАДЕНИЕ
	// ПЕРЕДЕЛАТЬ НА ПРОСТОЙ ЗАПРОС
	var fullOrders main.ResponseData
	OrderUID := ""
	//lessonNum := 0 // кол-во бонусных уроков
	//currentBalance := 5 // текущий баланс

	fullOrders.GetObjFromTpl(PayTplOrders, nil)

	for _,v := range fullOrders.Data {
		if a, found := v.Attr("paymentID", "value"); found {

			// да, это объект платежа, транзакцию которого надо подтвердить
			if a == PaymentID {
				OrderUID = v.Uid

				// пришел запрос на подтверждение возможности отгрузки. клиент оплатил и ждет ответа, после чего зачислятся деньги на мой счет
				if PaymentStatus == "waiting_for_capture" {
					// захватываю деньги
					captureMoney(OrderUID, PaymentID)
				}


				// последействие - деньги захвачены, надо пополнить баланс
				if PaymentStatus == "succeeded" {

					// работаем с объектом платежа, если он еще не подписан (бывают дубли запросов от шлюза яндекса)
					if k, _ := v.Attr("datetime_expires", "value"); k == "" {

						// обновляем аттрибут paymentID для объекта платежа, у которого paymentID совпадает с полученным в ответе шлюза
						v.AttrUpdate("datetime_expires", DatatimeExpires, "", "")
						v.AttrUpdate("payment_method", PaymentMethod, "", "")



						// /////////////////////////////////////////////////////////////////////
						// ТОЛЬКО ДЛЯ UMNICK.RU
						// 1. получем последний урок, пройденный студентом
						// 2. по иду урока получаем его название и вычленяем номе
						// 3. из названия товара расчитываем сколько добавлять
						// 4. обновляем профиль студента
						// /////////////////////////////////////////////////////////////////////
						//
						//		// узнаем последний законченный урок
						//		var objExitLesson ResponseData
						//		UserUID, _ := v.Attr("user", "src")
						//		Curl("GET", "/query/query_closeLesson?filter="+UserUID, "", &objExitLesson)
						//
						//
						//		// берем профиль оплатившего пользователя
						//		var objProfile ResponseData
						//		Curl("GET", "/query/userprofiles?hash="+UserUID, "", &objProfile)
						//
						//		// получаем оплаченные уже значения
						//		for _, v3 := range objProfile.Data {
						//			if cc, found := v3.Attr("payto", "value"); found {
						//				currentBalance, _ = strconv.Atoi(cc)
						//			}
						//		}
						//
						//		// первый объект = самый последний, берем его и получаем чтобы взять номер
						//		if len(objExitLesson.Data) > 0 {
						//
						//				for _, v1 := range objExitLesson.Data {
						//
						//					// получаем UID-самого урока
						//					lastLessonSRC, _ := v1.Attr("lesson","src") // получаем занчение типа Урок #70
						//
						//
						//					// берем объект последнего сданного урока
						//					var objLesson ResponseData
						//					Curl("GET", "_objs/"+lastLessonSRC, "", &objLesson)
						//
						//					lastLessonName, _ := objLesson.Data[0].Attr("title","value") // получаем занчение типа Урок #70
						//					num := strings.Split(lastLessonName, "#")[1] // номер последнего пройденного урока
						//					lessonNum, err = strconv.Atoi(num)
						//
						//					break
						//				}
						//		}
						//
						//
						//		plusValue := 0 // сколько прибавлять уроков к текущему
						//		// получаем название заказанного продукта (ТОЛЬКО ДЛЯ УМНИКА)
						//		if a, found := v.Attr("product", "value"); found {
						//			if strings.Contains(a, "1/2") {
						//				plusValue = 37
						//			}
						//			if strings.Contains(a, "Полный") {
						//				plusValue = 74
						//			}
						//			if strings.Contains(a, "10") {
						//				plusValue = 10
						//			}
						//		}
						//
						//		// выбираем большее значение из текущего урока или проплаченных ранее уроков
						//		if lessonNum < currentBalance {
						//			lessonNum = currentBalance
						//		}
						//
						//
						//		if err == nil {
						//			payNum := lessonNum + plusValue // получили номер проплаченогоДО урока
						//
						//			// ограничение до 74-х уроков
						//			if payNum > 74 {
						//				payNum = 74
						//			}
						//
						//			// обновляем профиль оплатившего пользователя
						//			var objProfile ResponseData
						//			Curl("GET", "/query/userprofiles?hash="+UserUID, "", &objProfile)
						//
						//			for _, v3 := range objProfile.Data {
						//				v3.AttrUpdate("payto", strconv.Itoa(payNum), "", "")
						//			}
						//
						//		}
						//
						//
						//		//// заменяем в профиле текущего пользователя оплаченные уроки
						//		//ctx := r.Context()
						//		//d := ctx.Value("User").(*ProfileData)
						//
						//
						//
						//		//log.Error("======================")
						//		//log.Error("======================")
						//		//log.Error("======================")
						//		//log.Error(ctx)
						//		//log.Error(d)
						//		//log.Error("======================")
						//		//log.Error("======================")
						//		//log.Error("======================")
						//		//
						//		//
						//		//resflag := d.CurrentProfile.AttrSet("payto", "value", strconv.Itoa(lessonNum))
						//		//
						//		//
						//		//log.Error("======================")
						//		//log.Error("======================")
						//		//log.Error("======================")
						//		//	log.Error(resflag)
						//		//	log.Error(ctx.Value("User"))
						//		//log.Error("======================")
						//		//log.Error("======================")
						//		//log.Error("======================")
						//
						//
						//// /////////////////////////////////////////////////////////////////////
						//// /////////////////////////////////////////////////////////////////////
						//// /////////////////////////////////////////////////////////////////////

					}

				}

			}
		}
	}


	w.WriteHeader(200)

}

// после подтверждения оплаты пользователем я захватываю деньги
// если не захвачу то улетят через 7 дней — при оплате банковской картой; 2 часа — при оплате любым другим способом.
func captureMoney(OrderUID, PaymentID string)  {
	client := &http.Client{}

	req, err := http.NewRequest("POST", MoneyGate+"/"+PaymentID+"/capture", nil)
	req.Header.Set("Idempotence-Key", OrderUID+"_cap")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(PayShopid, PaySecretKey)

	fmt.Println("Захват денег (url): ", req)

	resp, err := client.Do(req)
	if err != nil {
		main.log.Error(err, "Error: Capture payment (%s), transaction (%s) failed! (%s): ", OrderUID, PaymentID)
		return
	} else {
		defer resp.Body.Close()
	}


	main.log.Warning("Статус при захвате денег: ", resp.Status)
}


// данные заказанного/ых товара/ов (из списка товаров)
func getProductAttr(objProduct main.ResponseData) (amount int, amount_string, currency, product_pointsrc, product_pointvalue string) {

	product_v := []string{}
	product_s := []string{}

	for _, v := range objProduct.Data {

		// стоимость
		amount_string, _ = v.Attr("credit_string", "value")

		jj, found := v.Attr("credit", "value")
		if !found {
			return
		}
		kk, err := strconv.Atoi(jj)
		if err != nil {
			return
		}
		amount = amount + kk

		// описание - списка товаров
		dv, found 	:= v.Attr("title", "value")
		if found {
			product_v = append(product_v, dv)
			product_s = append(product_s, v.Uid)
		}


		currency, found = v.Attr("currency", "value")
		if !found {
			currency = "RUB"
		}

	}

	product_pointsrc = strings.Join(product_s, ",")
	product_pointvalue = strings.Join(product_v, ";")

	return
}


// формируем объект платежа
func setPayment(amount int, description, currency string) (payment Payment, err error) {

	payment.Amount = map[string]string{}
	payment.Amount["value"] = strconv.Itoa(amount)
	payment.Amount["currency"] = currency

	payment.Confirmation = map[string]string{}
	payment.Confirmation["type"] = "redirect"
	payment.Confirmation["return_url"] = PayRedirect

	payment.Description = description

	return payment, err
}


// отправляем запрос на формирование платежа на шлюз Yandex
func postPayment(w http.ResponseWriter, payment Payment, OrderUID string) (answer AnswerGateRound, err error)  {
	client := &http.Client{}

	bodyJSON, err := json.Marshal(payment)
	if err != nil {
		return answer, err
	}

	req, err := http.NewRequest("POST", MoneyGate, strings.NewReader(string(bodyJSON)))
	req.Header.Set("Idempotence-Key", OrderUID)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(PayShopid, PaySecretKey)

	resp, err := client.Do(req)
	if err != nil {
		main.log.Error(err, "Error: Post is request MoneyGate failed! (%s)")
		return answer, err
	} else {
		defer resp.Body.Close()
	}


	// отправляем запрос на бронирование платежа
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		main.log.Error(err, "Error: Parsing the answer Gateway is not valid! (%s)")
		return answer, err
	}
	responseString := string(responseData)

	err = json.Unmarshal([]byte(responseString), &answer)

	if err != nil {
		return answer, err
	}


	return answer,nil
}


// cоздаем объект платежа в базе
func createOrder(r *http.Request, payment Payment, product_pointsrc, product_pointvalue, amount_string string) (OrderUID string, res main.ResponseData) {
	data := map[string]string{}

	user := r.Context().Value("User").(*main.ProfileData)

	loc, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now().In(loc).Format("2006-01-02 15:04:05")

	data["data-source"] 		= PayTplOrders
	data["data-parent"] 		= PayTplOrders
	data["amount"] 				= payment.Amount["value"]
	data["amount_string"] 		= amount_string
	data["description"] 		= payment.Description
	data["product_pointsrc"] 	= product_pointsrc
	data["product_pointvalue"] 	= product_pointvalue
	data["user_pointsrc"] 		= user.Uid
	data["user_pointvalue"] 	= user.First_name + " " + user.Last_name
	data["datetime_created"] 	= now


	res = main.CreateObjForm(data, r)

	//fmt.Println("res: ", res)

	if len(res.Data) == 0 {
		return
	}

	OrderUID = res.Data[0].Uid

	return OrderUID, res
}
