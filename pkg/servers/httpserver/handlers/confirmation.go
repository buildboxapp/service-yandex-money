package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/service-yandex-money/pkg/model"
	"io/ioutil"
	"net/http"
)

// Service get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /confirmation [get]
func (h *handlers) Confirmation(w http.ResponseWriter, r *http.Request) {
	in, err := h.confirmationDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[Service] Error function execution (ServiceDecodeRequest).")
		return
	}
	serviceResult, err := h.service.Confirmation(r.Context(), in)
	if err != nil {
		h.logger.Error(err, "[Service] Error service execution (Service).")
		return
	}
	response, _ := h.confirmationEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.logger.Error(err, "[Service] Error function execution (ServiceEncodeResponse).")
		return
	}
	err = h.confirmationTransportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[Service] Error function execution (ServiceTransportResponse).")
		return
	}

	return
}

func (h *handlers) confirmationDecodeRequest(ctx context.Context, r *http.Request) (in model.AnswerConfirmation, err error)  {
	// читаем токен
	// запрашиваем профиль пользователя по юид-у
	// создаем in

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Errorf("Error: Parsing the answer Confirmation Gateway is not valid! (%s)", err)
		h.logger.Error(err, "Error YandexPay: Parsing the answer Confirmation Gateway is not valid! ")
		return
	}

	err = json.Unmarshal(data, &in)
	if err != nil {
		fmt.Errorf("Error: Unmarshal the answer Confirmation Gateway is not valid! (%s)", err)
		h.logger.Error(err, "Error YandexPay: Unmarshal the answer Confirmation Gateway is not valid ")
		return
	}


	return in, err
}

func (h *handlers) confirmationEncodeResponse(ctx context.Context, serviceResult *model.ConfirmationOut) (response string, err error)  {
	return response, err
}

func (h *handlers) confirmationTransportResponse(w http.ResponseWriter, response interface{}) (err error)  {
	d, err := json.Marshal(response)
	w.WriteHeader(200)

	if err != nil {
		w.WriteHeader(403)
	}
	w.Write(d)
	return err
}