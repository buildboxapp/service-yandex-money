package handlers

import (
	"context"
	"encoding/json"
	"github.com/buildboxapp/service-yandex-money/pkg/model"
	"net/http"
)

// Service get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /pay [get]
func (h *handlers) Pay(w http.ResponseWriter, r *http.Request) {
	in, err := h.payDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[Service] Error function execution (ServiceDecodeRequest).")
		return
	}
	serviceResult, err := h.service.Pay(r.Context(), in)
	if err != nil {
		h.logger.Error(err, "[Service] Error service execution (Service).")
		return
	}
	response, _ := h.payEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.logger.Error(err, "[Service] Error function execution (ServiceEncodeResponse).")
		return
	}
	err = h.payTransportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[Service] Error function execution (ServiceTransportResponse).")
		return
	}

	return
}

func (h *handlers) payDecodeRequest(ctx context.Context, r *http.Request) (in model.PayIn, err error)  {
	in.RedirectPostcreate = r.FormValue("redirect_postcreate") // редирект после создания объекта платежа (без оплаты в яндекс)
	in.Product = r.FormValue("product")  // товар, за который оплачиваем

	token, _ := h.jwt.Get(r)
	res, _ := token.Uid()
	in.Token.SetUid(res)

	if err != nil {
		return
	}

	// читаем токен
	// запрашиваем профиль пользователя по юид-у
	// создаем in
	return in, err
}

func (h *handlers) payEncodeResponse(ctx context.Context, serviceResult *model.PayOut) (response string, err error)  {
	return response, err
}

func (h *handlers) payTransportResponse(w http.ResponseWriter, response interface{}) (err error)  {
	d, err := json.Marshal(response)
	w.WriteHeader(200)

	if err != nil {
		w.WriteHeader(403)
	}
	w.Write(d)
	return err
}