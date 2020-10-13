// марштуры сервиса
package main

import "github.com/buildboxapp/services/yandex.pay/pkg/pay"

var routes = Routes{
	// системные ручки
	Route{"PIndex", "GET", "/", srv.PIndex},
	Route{"ProxyPing", "GET", "/ping",  srv.ProxyPing}, // функция ping-запросов

	// платежный модуль Yandex.Money
	Route{"GPayYandex", "GET", "/tools/pay", pay.GPayYandex},

	// адрес для уведомлений необходимо прописывать для каждого подключаемого сайта на стороне яндекс.касса
	// в https://kassa.yandex.ru/my/shop-settings -> URL для уведомлений
	// например: https://buildbox.app/buildbox/gui/tools/pay/confirmation
	Route{"GPayYandexСonfirmation", "POST", "/tools/pay/confirmation", pay.GPayYandexСonfirmation},
}
