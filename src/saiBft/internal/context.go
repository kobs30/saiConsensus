package internal

const (
	SaiBTCaddress                       = "saiBTC_address"
	SaiP2pAddress                       = "saiP2P_address"
	SaiVM1Address                       = "saiVM1_address"
	SaiStorageToken                     = "storage_token"
	SaiSleep                            = "sleep"
	SaiBTCKeys                          = "saiBTCKeys"
	SaiDuplicateStorageRequests         = "duplicate_storage_requests"
	SaiDuplicateStorageRequestsURL      = "duplicate_storage_requests_url"
	SaiDuplicateStorageRequestsInterval = "duplicate_storage_requests_interval"
	SaiP2pProxyAddress                  = "saiProxy_address"
)

// func (s *InternalService) SetContext(btcKeys *models.BtcKeys) {
// 	vm1Address, ok := s.GlobalService.Configuration["saiVM1_address"].(string)
// 	if !ok {
// 		s.GlobalService.Logger.Fatal("processing - wrong type of saiVM1 address value from config")
// 	}

// 	btcAddress, ok := s.GlobalService.Configuration["saiBTC_address"].(string)
// 	if !ok {
// 		s.GlobalService.Logger.Fatal("processing - wrong type of saiBTC address value from config")
// 	}

// 	p2pAddress, ok := s.GlobalService.Configuration["saiP2P_address"].(string)
// 	if !ok {
// 		s.GlobalService.Logger.Fatal("processing - wrong type of saiP2P address value from config")
// 	}

// 	storageToken, ok := s.GlobalService.Configuration["storage_token"].(string)
// 	if !ok {
// 		s.GlobalService.Logger.Fatal("handlers - processing - wrong type of storage token value from config")
// 	}

// 	sleepValue, ok := s.GlobalService.Configuration["sleep"].(int)
// 	if !ok {
// 		s.GlobalService.Logger.Sugar().Fatalf("handlers - processing - wrong type of sleep value from config, provided type : %s", reflect.TypeOf(sleepValue).String())
// 	}

// 	duplicateRequests, ok := Service.GlobalService.Configuration["duplicate_storage_requests"].(bool)
// 	if !ok {
// 		duplicateRequests = false
// 	}

// 	duplicateRequestsUrl, ok := Service.GlobalService.Configuration["duplicate_storage_requests_url"].(string)
// 	if !ok {
// 		duplicateRequestsUrl = ""
// 	}
// 	duplicateRequestsInterval, ok := Service.GlobalService.Configuration["duplicate_storage_requests_interval"].(int)
// 	if !ok {
// 		duplicateRequestsInterval = 0
// 	}

// 	p2pProxyAddress, ok := s.GlobalService.Configuration["saiProxy_address"].(string)
// 	if !ok {
// 		s.GlobalService.Logger.Fatal("processing - wrong type of saiProxy address value from config")
// 	}

// 	s.CoreCtx = context.WithValue(context.Background(), SaiBTCaddress, btcAddress)
// 	s.CoreCtx = context.WithValue(s.CoreCtx, SaiP2pAddress, p2pAddress)
// 	s.CoreCtx = context.WithValue(s.CoreCtx, SaiVM1Address, vm1Address)
// 	s.CoreCtx = context.WithValue(s.CoreCtx, SaiStorageToken, storageToken)
// 	s.CoreCtx = context.WithValue(s.CoreCtx, SaiSleep, sleepValue)
// 	s.CoreCtx = context.WithValue(s.CoreCtx, SaiBTCKeys, btcKeys)
// 	s.CoreCtx = context.WithValue(s.CoreCtx, SaiDuplicateStorageRequests, duplicateRequests)
// 	s.CoreCtx = context.WithValue(s.CoreCtx, SaiDuplicateStorageRequestsURL, duplicateRequestsUrl)
// 	s.CoreCtx = context.WithValue(s.CoreCtx, SaiDuplicateStorageRequestsInterval, duplicateRequestsInterval)
// 	s.CoreCtx = context.WithValue(s.CoreCtx, SaiP2pProxyAddress, p2pProxyAddress)
// }
