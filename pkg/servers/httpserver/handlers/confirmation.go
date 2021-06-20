package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/yookassa/pkg/model"
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
	answer, in, err := h.confirmationDecodeRequest(r.Context(), r)
	if err != nil {
		h.transportError(w, 500, err, "[Confirmation] Error service execution (confirmationDecodeRequest)")
		return
	}
	serviceResult, err := h.service.Confirmation(r.Context(), answer, in)
	if err != nil {
		h.transportError(w, 500, err, "[Confirmation] Error service execution (Confirmation)")
		return
	}
	response, _ := h.confirmationEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.transportError(w, 500, err, "[Confirmation] Error service execution (confirmationEncodeResponse)")
		return
	}
	err = h.transportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[Confirmation] Error function execution (transportResponse).")
		return
	}

	return
}

func (h *handlers) confirmationDecodeRequest(ctx context.Context, r *http.Request) (answer model.AnswerConfirmation, in model.ConfirmationIn, err error)  {
	configToken := r.FormValue("token")  // идентификатор текущей конфигурации в конф.шлюза

	// ищем текущую конфигурацию для данного запроса исходя из переданного token
	for _, v := range h.cfg.Custom {
		// нашли текущую активную конфигурацию
		if v.Token == configToken && v.Active != "" {
			in.Configuration = v
		}
	}

	// отправляем запрос на бронирование платежа
	responseData, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Errorf("Error: Parsing the answer Confirmation Gateway is not valid! (%s)", err)
		h.logger.Error(err, "Error YandexPay: Parsing the answer Confirmation Gateway is not valid! ")
		return
	}

	err = json.Unmarshal(responseData, &answer)
	if err != nil {
		fmt.Errorf("Error: Unmarshal the answer Confirmation Gateway is not valid! (%s)", err)
		h.logger.Error(err, "Error YandexPay: Unmarshal the answer Confirmation Gateway is not valid ")
		return
	}

	return answer, in, err
}

func (h *handlers) confirmationEncodeResponse(ctx context.Context, serviceResult *model.ConfirmationOut) (response string, err error)  {
	return response, err
}