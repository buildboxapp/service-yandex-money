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


package service

import (
	"context"
	"github.com/buildboxapp/yookassa/pkg/model"
	"net/http"
	"strconv"
	"encoding/json"
	"strings"
	"fmt"
	"time"
	"io/ioutil"
)


// Payment - платеж, который будет отправлен на платежный шлюз
// Order - объект счета, который создается в БД и где фиксируются атрибуты платежа
// Product - товар/услуга, которая оплачивается (только берем данные)


func (s *service) Pay(ctx context.Context, in model.PayIn) (out model.PayOut, err error) {
	var objProduct model.ResponseData

	// объект через системный запрос
	_, err = s.utils.Curl("GET", in.Configuration.ApiUrl + "/query/obj?obj="+in.Product, "", &objProduct, map[string]string{})
	if err != nil {
		return
	}
	if 	len(objProduct.Data) == 0 {
		err = fmt.Errorf("%s", "Error. Object product is empty")
		return
	}

	// 0. формируем аттрибуты для платежа и счета
	amount, amount_string, currency, product_pointsrc, product_pointvalue := s.getProductAttr(objProduct)

	// 1. создаем платеж (формата шлюза)
	description := product_pointvalue
	//description := "Оплата товара:" + product_pointvalue + " на сумму " + strconv.Itoa(amount)
	payment, err := s.setPayment(amount, description, currency, in)
	if err != nil {
		return
	}

	// 2. создаем объект счета
	OrderUID, OrderObj, err := s.createOrder(payment, product_pointsrc, product_pointvalue, amount_string, in)
	if err != nil {
		return
	}

	out.Body, _ = json.Marshal(OrderUID)

	// 2.1. редиректим на страницу счета с указанием uid созданного платежа
	if in.RedirectPostcreate != "" {
		out.RedirectUrl = in.RedirectPostcreate + OrderUID
		out.Code = 302
		return
	}

	// 3. отправляем запрос на бронирование платежа и перенаправление на страницу оплаты шлюза
	answer, err := s.postPayment(payment, OrderUID, in)
	if err != nil {
		return
	}

	//fmt.Println("Объект платежа, который пришел со шлуюза: ", answer)
	// если оплата провелась в один этап (ApplePay или тестовый), то проверяем на статус и
	// обновляем объект платежа в базе на выполенный (ставим дату списания)
	//if answer.Status == "succeeded" {
	//	datetime_expires := answer.CreatedAt
	//	DatatimeExpires := datetime_expires.Format("2006-01-02 15:04:05")
	//	OrderObj.Data[0].AttrUpdate("datetime_expires", DatatimeExpires, "", "")
	//}

	// 4. сохраняю идентификатор платежа шлюза в объекте платеж
	err = s.api.AttrUpdate(OrderObj.Data[0].Uid, "paymentID", answer.ID, "", "")
	if err != nil {
		return
	}

	redirect_url := answer.Confirmation["confirmation_url"]
	//http.Redirect(w, r, redirect_url, 302)
	out.RedirectUrl = redirect_url
	out.Code = 302

	// получение формы оплаты и редирект должен произойти в postPayment
	//http.Redirect(w, r, cfg.PayErrorRedirect, 302)
	//out.RedirectUrl = in.Configuration.RedirectError
	//out.Code = 302

	return
}

// данные заказанного/ых товара/ов (из списка товаров)
func (s *service) getProductAttr(objProduct model.ResponseData) (amount int, amount_string, currency, product_pointsrc, product_pointvalue string) {

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
func (s *service) setPayment(amount int, description, currency string, in model.PayIn) (payment model.Payment, err error) {

	payment.Amount = map[string]string{}
	payment.Amount["value"] = strconv.Itoa(amount)
	payment.Amount["currency"] = currency

	payment.Confirmation = map[string]string{}
	payment.Confirmation["type"] = "redirect"
	payment.Confirmation["return_url"] = in.Configuration.Redirecturl

	payment.Description = description

	return payment, err
}


// отправляем запрос на формирование платежа на шлюз Yandex
func (s *service) postPayment(payment model.Payment, OrderUID string, in model.PayIn) (answer model.AnswerGateRound, err error)  {
	client := &http.Client{}

	bodyJSON, err := json.Marshal(payment)
	if err != nil {
		return answer, err
	}

	req, err := http.NewRequest("POST", in.Configuration.MoneyGate, strings.NewReader(string(bodyJSON)))
	req.Header.Set("Idempotence-Key", OrderUID)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(in.Configuration.Shopid, in.Configuration.Shopkey)

	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error(err, "Error: Post is request MoneyGate failed! (%s)")
		return answer, err
	} else {
		defer resp.Body.Close()
	}

	// отправляем запрос на бронирование платежа
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error(err, "Error: Parsing the answer Gateway is not valid! (%s)")
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
func (s *service) createOrder(payment model.Payment, product_pointsrc, product_pointvalue, amount_string string, in model.PayIn) (OrderUID string, res model.ResponseData, err error) {
	data := map[string]string{}

	loc, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now().In(loc).Format("2006-01-02 15:04:05")

	data["data-source"] 		= in.Configuration.TplOrders
	data["data-parent"] 		= in.Configuration.TplOrders
	data["amount"] 				= payment.Amount["value"]
	data["amount_string"] 		= amount_string
	data["description"] 		= payment.Description
	data["product_pointsrc"] 	= product_pointsrc
	data["product_pointvalue"] 	= product_pointvalue
	data["user_pointsrc"] 		= in.UserUID
	data["user_pointvalue"] 	= in.UserName
	data["datetime_created"] 	= now

	res, err = s.api.CreateObjForm(data)
	if err != nil {
		err = fmt.Errorf("%s", "Error. Create object Order failed. " + fmt.Sprint(err))
		return
	}
	if len(res.Data) == 0 {
		err = fmt.Errorf("%s", "Error. Create object Order failed. Returned object is empty.")
		return
	}
	OrderUID = res.Data[0].Uid

	return
}