// обработчики запроса сервиса
package service

import (
	"net/http"
)

// Собираем страницу (параметры из конфига) и пишем в w.Write
func (c *Service) PIndex(w http.ResponseWriter, r *http.Request) {
	result := "Service done"

	w.WriteHeader(200)
	w.Write([]byte(result))
}

