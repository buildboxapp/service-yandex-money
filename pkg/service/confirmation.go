package service

import (
	"context"
	"fmt"
	"github.com/buildboxapp/service-yandex-money/pkg/model"
)

// сюда приходит подтверждение оплаты от шлюза
// ручка, которую дергает шлюз
func (s *service) Confirmation(ctx context.Context, in model.AnswerConfirmation) (out model.ConfirmationOut, err error) {

	s.logger.Info("Пришел запрос подтверждения от Yandex-Шлюза на оплату")
	fmt.Println("Пришел запрос подтверждения от Yandex-Шлюза на оплату ")

	// отправляем запрос на бронирование платежа
	PaymentID := in.Object.ID
	PaymentMethod := in.Object.PaymentMethod.Title
	PaymentStatus := in.Object.Status

	datetime_expires := in.Object.CreatedAt
	DatatimeExpires := datetime_expires.Format("2006-01-02 15:04:05")


	// ПРОБЕГАЮ ВСЕ ПЛАТЕЖИ И СМОТРЮ СОВПАДЕНИЕ
	var fullOrders model.ResponseData
	s.utils.Curl("POST", "_data/"+s.cfg.Pay.Ordertpl, "", &fullOrders, map[string]string{})

	OrderUID := ""

	for _,v := range fullOrders.Data {
		if a, found := v.Attr("paymentID", "value"); found {

			// да, это объект платежа, транзакцию которого надо подтвердить
			if a == PaymentID {
				OrderUID = v.Uid

				// пришел запрос на подтверждение возможности отгрузки. клиент оплатил и ждет ответа, после чего зачислятся деньги на мой счет
				if PaymentStatus == "waiting_for_capture" {
					// захватываю деньги
					s.captureMoney(OrderUID, PaymentID)
				}

				// последействие - деньги захвачены, надо пополнить баланс
				if PaymentStatus == "succeeded" {

					// работаем с объектом платежа, если он еще не подписан (бывают дубли запросов от шлюза яндекса)
					if k, _ := v.Attr("datetime_expires", "value"); k == "" {

						// обновляем аттрибут paymentID для объекта платежа, у которого paymentID совпадает с полученным в ответе шлюза
						s.api.AttrUpdate(v.Uid, "datetime_expires", DatatimeExpires, "", "")
						s.api.AttrUpdate(v.Uid, "payment_method", PaymentMethod, "", "")
					}

				}

			}
		}
	}

	return
}
