package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/lib"
	"github.com/buildboxapp/lib/log"
	bbmetric "github.com/buildboxapp/lib/metric"
	"github.com/buildboxapp/yookassa/pkg/model"
	"net/http"
	"runtime/debug"
	"time"
)

func (h *httpserver) MiddleLogger(next http.Handler, name string, logger log.Log, serviceMetrics bbmetric.ServiceMetric) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)
		timeInterval := time.Since(start)
		if name != "ProxyPing"  { //&& false == true
			mes := fmt.Sprintf("Query: %s %s %s %s",
				r.Method,
				r.RequestURI,
				name,
				timeInterval)
			logger.Info(mes)
		}

		// сохраняем статистику всех запросов, в том числе и пинга (потому что этот запрос фиксируется в количестве)
		serviceMetrics.SetTimeRequest(timeInterval)
	})
}

func (h *httpserver) AuthProcessor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenRaw string
		var token model.Token

		// пропускаем пинги и подтверждения оплаты от шлюза
		if r.URL.Path == "/ping" || r.URL.Path == "/confirmation" {
			next.ServeHTTP(w, r)
			return
		}

		authKeyHeader := r.Header.Get("X-Auth-Token")
		if authKeyHeader != "" {
			tokenRaw = authKeyHeader
		} else {
			authKeyCookie, err := r.Cookie("X-Auth-Token")
			if err == nil {
				tokenRaw = authKeyCookie.Value
			}
		}
		if tokenRaw == "" {
			lib.ResponseJSON(w, nil, "Unauthorized", nil, nil)
			return
		}

		// дешифруем токен ключом проекта
		tokenStr, err := lib.Decrypt([]byte(h.cfg.ProjectKey), tokenRaw)

		//fmt.Println(tokenStr, err)

		if err != nil {
			lib.ResponseJSON(w, nil, "Unauthorized", nil, nil)
			return
		}
		json.Unmarshal([]byte(tokenStr), &token)

		//fmt.Println(token)


		// НЕТ НИКАКИХ ПРОВЕРОК НА НАЛИЧИЕ ПРАВА
		// ПРОСТО РАСШИФРОВЫВАЕМ ТОКЕН И БЕРЕМ ИД

		// не передали ключ (пропускаем пинги)
		//if strings.TrimSpace(token.Uid) == "" {
		//	lib.ResponseJSON(w, nil, "Unauthorized", nil, nil)
		//	return
		//}

		// добавили в контектс значение текущего токена
		h.ctx = context.WithValue(h.ctx, "token", token)

		// не соответствие переданного ключа и UID-а API (пропускаем пинги)
		//if strings.TrimSpace(token.Uid) != h.cfg.UidService && r.URL.Path != "/ping" {
		//	lib.ResponseJSON(w, nil, "Unauthorized", nil, nil)
		//	return
		//}

		next.ServeHTTP(w, r)
	})
}

func (h *httpserver) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(r *http.Request) {
			rec := recover()
			if rec != nil {
				b := string(debug.Stack())
				//fmt.Println(r.URL.String())
				h.logger.Panic(fmt.Errorf("%s", b), "Recover panic from path: ", r.URL.String(), "; form: ", r.Form)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}(r)
		next.ServeHTTP(w, r)
	})
}

func (h *httpserver) JsonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
