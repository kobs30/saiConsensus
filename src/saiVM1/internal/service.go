package internal

import (
	"github.com/saiset-co/saiService"
	"github.com/saiset-co/saiVM1/utils"
)

var counter = 0

type InternalService struct {
	Context *saiService.Context
	Storage *utils.Database
}

func Service(context *saiService.Context) *InternalService {
	return &InternalService{
		Context: context,
		Storage: utils.Storage(
			context.GetConfig("service.storage.url", "").(string),
			context.GetConfig("service.storage.token", "").(string),
		),
	}
}
