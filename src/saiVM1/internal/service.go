package internal

import (
	"github.com/saiset-co/saiService"
	"github.com/saiset-co/saiVM1/utils"
)

type InternalService struct {
	Context *saiService.Context
	Storage *utils.Database
}

func Service(context *saiService.Context) InternalService {
	return InternalService{
		Context: context,
		Storage: new(utils.Database),
	}
}

var counter = 0

func (is InternalService) Init() {
	is.Storage = utils.Storage(is.Context.GetConfig("service.storage.url", "").(string), is.Context.GetConfig("service.storage.token", "").(string))
}
