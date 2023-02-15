package internal

import (
	"bytes"
	"sync"

	"github.com/iamthe1whoknocks/bft/models"
	"github.com/iamthe1whoknocks/bft/utils"
	"github.com/iamthe1whoknocks/saiService"
	"go.uber.org/zap"
)

// here we add all implemented handlers, create name of service and register config
// moved from handlers to service because of initialization problems
func Init(svc *saiService.Service) {
	storage := NewDB(Service.DuplicateStorageCh)
	Service.Storage = storage

	btckeys, err := Service.GetBTCkeys("btc_keys.json", Service.GlobalService.Configuration["saiBTC_address"].(string))
	if err != nil {
		svc.Logger.Fatal("main - init - open btc keys", zap.Error(err))
	}
	Service.BTCkeys = btckeys
	svc.Logger.Debug("main - init", zap.Any("btc keys", btckeys)) //DEBUG

	Service.IpAddress = utils.GetOutboundIP()
	if Service.IpAddress == "" {
		svc.Logger.Fatal("Cannot detect outbound IP address of node")
	}
	svc.Logger.Debug("main - init", zap.String("ip address", Service.IpAddress)) //DEBUG

	Service.Handler[GetMissedBlocks.Name] = GetMissedBlocks
	Service.Handler[HandleTxFromCli.Name] = HandleTxFromCli
	Service.Handler[HandleMessage.Name] = HandleMessage
	Service.Handler[CreateBTCKeys.Name] = CreateBTCKeys
	Service.Handler[GetTx.Name] = GetTx
	Service.Handler[AddValidator.Name] = AddValidator

	// setting values to core context
	//Service.SetContext(btckeys)
	//svc.Logger.Sugar().Debugf("main - init - core context :[%+v]", Service.CoreCtx)

}

type InternalService struct {
	Handler              saiService.Handler  // handlers to define in this specified microservice
	GlobalService        *saiService.Service // saiService reference
	Validators           []string
	ConnectedSaiP2pNodes map[string]*models.SaiP2pNode
	BTCkeys              *models.BtcKeys
	MsgQueue             chan interface{}
	InitialSignalCh      chan interface{} // chan for notification, if initial block consensus msg was got already
	IsInitialized        bool             // if inital block consensus msg was got or timeout was passed
	Storage              utils.Database
	IpAddress            string        // outbound ip address
	MissedBlocksLinkCh   chan string   //chan to get link from p2pProxy handler
	TxHandlerSyncCh      chan struct{} // chan to handle tx via http/cli
	IsValidator          bool          //is node a validator
	//CoreCtx              context.Context
	DuplicateStorageCh chan *bytes.Buffer //chan for duplicate save/update/upsert requests to storage
	SyncConsensus      *SyncConsensus     // for consensus sync
	Sleep              int                // moved here to change dynamically
}

// global handler for registering handlers
var Service = &InternalService{
	Handler:              saiService.Handler{},
	ConnectedSaiP2pNodes: make(map[string]*models.SaiP2pNode),
	MsgQueue:             make(chan interface{}),
	InitialSignalCh:      make(chan interface{}),
	IsInitialized:        false,
	MissedBlocksLinkCh:   make(chan string),
	TxHandlerSyncCh:      make(chan struct{}),
	DuplicateStorageCh:   make(chan *bytes.Buffer, 100),
	SyncConsensus: &SyncConsensus{
		Mu:      new(sync.RWMutex),
		Storage: make(map[models.SyncConsensusKey]int),
	},
}

// struct for handling consensus sync
type SyncConsensus struct {
	Mu      *sync.RWMutex
	Storage map[models.SyncConsensusKey]int
}
