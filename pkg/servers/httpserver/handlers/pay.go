package handlers

import (
	"context"
	"fmt"
	"github.com/buildboxapp/yookassa/pkg/model"
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
		h.transportError(w, 500, err, "[Pay] Error service execution (payDecodeRequest)")
		return
	}
	serviceResult, err := h.service.Pay(r.Context(), in)
	if err != nil {
		h.transportError(w, 500, err, "[Pay] Error service execution (Pay)")
		return
	}
	if serviceResult.Code == 302 && serviceResult.RedirectUrl != "" {
		http.Redirect(w, r, serviceResult.RedirectUrl, serviceResult.Code)
	}
	response, _ := h.payEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.transportError(w, 500, err, "[Pay] Error service execution (payEncodeResponse)")
		return
	}
	err = h.transportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[Pay] Error function execution (transportResponse).")
		return
	}

	return
}

func (h *handlers) payDecodeRequest(ctx context.Context, r *http.Request) (in model.PayIn, err error)  {
	in.RedirectPostcreate = r.FormValue("redirect_postcreate") // редирект после создания объекта платежа (без оплаты в яндекс)
	in.Product = r.FormValue("product")  // товар, за который оплачиваем
	configToken := r.FormValue("token")  // идентификатор текущей конфигурации в конф.шлюза

	// ищем текущую конфигурацию для данного запроса исходя из переданного token
	for _, v := range h.cfg.Custom {
		// нашли текущую активную конфигурацию
		if v.Token == configToken && v.Active == "true" {
			in.Configuration = v
		}
	}

	if in.Product == "" {
		err = fmt.Errorf("%s", "Error. Param product is empty")
		return
	}

	// создаем in
	ctx1 := *h.ctx
	token, ok := ctx1.Value("token").(model.Token)
	if ok {
		in.UserUID = token.Uid
		in.UserName = token.Info.Name
	}
	if err != nil {
		return
	}

	return in, err
}

func (h *handlers) payEncodeResponse(ctx context.Context, serviceResult *model.PayOut) (response string, err error)  {
	return response, err
}