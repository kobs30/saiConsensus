package internal

import (
	"context"
	"reflect"

	"github.com/iamthe1whoknocks/bft/models"
)

const (
	SaiBTCaddress   = "saiBTC_address"
	SaiP2pAddress   = "saiP2P_address"
	SaiStorageToken = "storage_token"
	SaiSleep        = "sleep"
	SaiBTCKeys      = "saiBTCKeys"
)

func (s *InternalService) SetContext(btcKeys *models.BtcKeys) {
	btcAddress, ok := s.GlobalService.Configuration["saiBTC_address"].(string)
	if !ok {
		s.GlobalService.Logger.Fatal("processing - wrong type of saiBTC address value from config")
	}
	p2pAddress, ok := s.GlobalService.Configuration["saiP2P_address"].(string)
	if !ok {
		s.GlobalService.Logger.Fatal("processing - wrong type of saiP2P address value from config")
	}

	storageToken, ok := s.GlobalService.Configuration["storage_token"].(string)
	if !ok {
		s.GlobalService.Logger.Fatal("handlers - processing - wrong type of storage token value from config")
	}

	sleepValue, ok := s.GlobalService.Configuration["sleep"].(int)
	if !ok {
		s.GlobalService.Logger.Sugar().Fatalf("handlers - processing - wrong type of sleep value from config, provided type : %s", reflect.TypeOf(sleepValue).String())
	}

	s.CoreCtx = context.WithValue(context.Background(), SaiBTCaddress, btcAddress)
	s.CoreCtx = context.WithValue(s.CoreCtx, SaiP2pAddress, p2pAddress)
	s.CoreCtx = context.WithValue(s.CoreCtx, SaiStorageToken, storageToken)
	s.CoreCtx = context.WithValue(s.CoreCtx, SaiSleep, sleepValue)
	s.CoreCtx = context.WithValue(s.CoreCtx, SaiBTCKeys, btcKeys)
}
