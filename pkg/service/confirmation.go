package service

import (
	"context"
	"fmt"
	"github.com/buildboxapp/yookassa/pkg/model"
	"net/http"
)

// сюда приходит подтверждение оплаты от шлюза
// ручка, которую дергает шлюз
func (s *service) Confirmation(ctx context.Context, answer model.AnswerConfirmation, in model.ConfirmationIn) (out model.ConfirmationOut, err error) {

	s.logger.Info("Пришел запрос подтверждения от Yandex-Шлюза на оплату")
	//fmt.Println("Пришел запрос подтверждения от Yandex-Шлюза на оплату ")

	// отправляем запрос на бронирование платежа
	PaymentID := answer.Object.ID
	PaymentMethod := answer.Object.PaymentMethod.Title
	PaymentStatus := answer.Object.Status

	datetime_expires := answer.Object.CreatedAt
	DatatimeExpires := datetime_expires.Format("2006-01-02 15:04:05")

	// ПРОБЕГАЮ ВСЕ ПЛАТЕЖИ И СМОТРЮ СОВПАДЕНИЕ
	var fullOrders model.ResponseData
	_, err = s.utils.Curl("POST", in.Configuration.ApiUrl + "/query/qtplobjs?obj="+in.Configuration.TplOrders, "", &fullOrders, map[string]string{})

	OrderUID := ""

	for _,v := range fullOrders.Data {
		if a, found := v.Attr("paymentID", "value"); found {

			// да, это объект платежа, транзакцию которого надо подтвердить
			if a == PaymentID {
				OrderUID = v.Uid

				// пришел запрос на подтверждение возможности отгрузки. клиент оплатил и ждет ответа, после чего зачислятся деньги на мой счет
				if PaymentStatus == "waiting_for_capture" {
					// захватываю деньги
					s.captureMoney(OrderUID, PaymentID, in)
				}

				// последействие - деньги захвачены, надо пополнить баланс
				if PaymentStatus == "succeeded" {
					// работаем с объектом платежа, если он еще не подписан (бывают дубли запросов от шлюза яндекса)
					if k, _ := v.Attr("datetime_expires", "value"); k == "" {

						// обновляем аттрибут paymentID для объекта платежа, у которого paymentID совпадает с полученным в ответе шлюза
						err = s.api.AttrUpdate(v.Uid, "datetime_expires", DatatimeExpires, "", "")
						if err != nil {
							fmt.Println("Error AttrUpdate ", err)
							s.logger.Error(err, "Error AttrUpdate")
							return
						}
						err = s.api.AttrUpdate(v.Uid, "payment_method", PaymentMethod, "", "")
						if err != nil {
							fmt.Println("Error AttrUpdate ", err)
							s.logger.Error(err, "Error AttrUpdate")
							return
						}
					}
				}
			}
		}
	}

	return
}

// после подтверждения оплаты пользователем я захватываю деньги
// если не захвачу то улетят через 7 дней — при оплате банковской картой; 2 часа — при оплате любым другим способом.
func (s *service) captureMoney(OrderUID, PaymentID string, in model.ConfirmationIn)  {
	client := &http.Client{}

	req, err := http.NewRequest("POST", in.Configuration.MoneyGate+"/"+PaymentID+"/capture", nil)
	req.Header.Set("Idempotence-Key", OrderUID+"_cap")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(in.Configuration.Shopid, in.Configuration.Shopkey)

	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error(err, "Error: Capture payment (%s), transaction (%s) failed! (%s): ", OrderUID, PaymentID)
		return
	} else {
		defer resp.Body.Close()
	}

	s.logger.Warning("Статус при захвате денег: ", resp.Status)
}
