package jwt

import (
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/service-yandex-money/pkg/model"
	"net/http"
	"strings"
)

// обновление элемента (через вызов внешнего запроса)
func (j *jwt) Get(r *http.Request) (token token, err error) {

	// 1 пробуем получить токен из куки 'token'
	token, err = j.getTokenFromCookie(r, "token")
	if err == nil {
		return
	}

	// 2 пробуем получить токен из контекста для совместимости со старыми сессиями GUI
	token, err = j.getTokenFromContext(r)
	if err == nil {
		return
	}

	// 3 пробуем получить токен из заголовка
	token, err = j.getTokenFromHeader(r)

	return
}

// получаем токен из куки
func (j *jwt) getTokenFromCookie(r *http.Request, name string) (token token, err error) {
	cookieToken, err := r.Cookie(name)
	if cookieToken.Value == "" {
		err = fmt.Errorf("%s", "Cookie '"+name+"' is empty")
		return
	}
	token, err = j.parseValidateJWT(cookieToken.Value)
	if err != nil {
		return
	}

	return
}

// получаем токен из заголовка
func (j *jwt) getTokenFromHeader(r *http.Request) (token token, err error) {
	ha := r.Header.Get("Authorization")
	if len(ha) == 0 {
		err = fmt.Errorf("%s", "Header 'Authorization' is empty")
		return
	}
	splitToken := strings.Split(ha, "Bearer ")
	if len(splitToken) < 2 {
		err = fmt.Errorf("Authorization required")
		return
	}
	token, err = j.parseValidateJWT(splitToken[1])

	return
}

// получаем токен из контекста
func (j *jwt) getTokenFromContext(r *http.Request) (token token, err error) {
	user := r.Context().Value("User").(*model.ProfileData)

	token.SetUid(user.Uid)
	return
}


// получаем токен из строки
// временно просто маршалим из джейсона
func (j *jwt) parseValidateJWT(str string) (token token, err error) {
	err = json.Unmarshal([]byte(str), &token)
	return
}