package internal

import (
	"bytes"
	"log"
	"time"

	"github.com/iamthe1whoknocks/bft/utils"
	"go.uber.org/zap"
)

func NewDB(duplicateCh chan *bytes.Buffer) utils.Database {
	url, ok := Service.GlobalService.Configuration["storage_url"].(string)
	if !ok {
		log.Fatalf("configuration : invalid storage url provided, url : %s", Service.GlobalService.Configuration["storage_url"])
	}
	email, ok := Service.GlobalService.Configuration["storage_email"].(string)
	if !ok {
		log.Fatalf("configuration : invalid storage email provided, email : %s", Service.GlobalService.Configuration["storage_email"])
	}
	password, ok := Service.GlobalService.Configuration["storage_password"].(string)
	if !ok {
		log.Fatalf("configuration : invalid storage password provided, password : %s", Service.GlobalService.Configuration["storage_email"])
	}

	duplicateRequests, ok := Service.GlobalService.Configuration["duplicate_storage_requests"].(bool)
	if !ok {
		duplicateRequests = false
	}

	duplicateRequestsUrl, ok := Service.GlobalService.Configuration["duplicate_storage_requests_url"].(string)
	if !ok {
		duplicateRequestsUrl = ""
	}

	return utils.Storage(url, email, password, duplicateRequests, duplicateRequestsUrl, duplicateCh)
}

func (s *InternalService) duplicateRequests() {
	for {
		time.Sleep(time.Duration(s.GlobalService.GetConfig(SaiDuplicateStorageRequestsInterval, 3).Int()) * time.Second)
		buf := <-s.DuplicateStorageCh
		err, _ := utils.Send(s.GlobalService.GetConfig(SaiDuplicateStorageRequestsURL, "").String(), buf, "")
		if err != nil {
			s.GlobalService.Logger.Error("process - duplicate requests - send", zap.Error(err))
		}
	}
}
