package utils

import (
	"github.com/buildboxapp/service-yandex-money/pkg/config"
	"github.com/buildboxapp/service-yandex-money/pkg/model"
	"github.com/buildboxapp/lib/log"
)

type utils struct {
	cfg config.Config
	logger log.Log
}

type Utils interface {
	AddressProxy() (port string)
	Curl(method, urlc, bodyJSON string, response interface{}, headers map[string]string) (result interface{}, err error)
	RemoveElementFromData(p *model.ResponseData, i int) bool
}


func New(cfg config.Config, logger log.Log) Utils {
	return &utils{
		cfg,
		logger,
	}
}

/////////////////////////////////////////////////////
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
/////////////////////////////////////////////////////

// удаляем элемент из слайса
func (u *utils) RemoveElementFromData(p *model.ResponseData, i int) bool {

	if (i < len(p.Data)){
		p.Data = append(p.Data[:i], p.Data[i+1:]...)
	} else {
		//log.Warning("Error! Position invalid (", i, ")")
		return false
	}

	return true
}