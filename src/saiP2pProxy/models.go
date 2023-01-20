package main

import (
	"saiP2p/utils"

	valid "github.com/asaskevich/govalidator"
	"go.uber.org/zap"
)

const (
	SyncRequestType  = "sync_request"
	SyncResponseType = "sync_response"
)

type Proxy struct {
	Config  *Config
	Storage *utils.Database
	Logger  *zap.Logger
}

type Config struct {
	Host            string `yaml:"proxy_host"`
	Port            string `yaml:"proxy_port"`
	ProxyEndpoint   string `yaml:"proxy_endpoint"`
	BftHost         string `yaml:"bft_http_host"`
	BftPort         string `yaml:"bft_http_port"`
	P2pHost         string `yaml:"p2p_host"`
	P2pPort         string `yaml:"p2p_port"`
	StorageURL      string `yaml:"storage_url"`
	StorageEmail    string `yaml:"storage_email"`
	StoragePassword string `yaml:"storage_password"`
	StorageToken    string `yaml:"storage_token"`
}

type SyncRequest struct {
	Type    string `json:"type"`
	From    int    `json:"block_number_from"`
	To      int    `json:"block_number_to"`
	Address string `json:"address"`
}

type SyncResponse struct {
	Type  string `json:"type"`
	Error error  `json:"error"`
	Link  string `json:"link"`
}

type jsonRequestType struct {
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}

// BlockConsensus message
type BlockConsensusMessage struct {
	BlockHash  string   `json:"block_hash" valid:",required"`
	Votes      int      `json:"votes"` // additional field, which was not added by Valeriy
	Block      *Block   `json:"block" valid:",required"`
	Count      int      `json:"-"` // extended value for consensus while get missed blocks from p2p services
	Signatures []string `json:"voted_signatures"`
}

type Block struct {
	Type              string                `json:"type" valid:",required"`
	Number            int                   `json:"number" valid:",required"`
	PreviousBlockHash string                `json:"prev_block_hash" valid:",required"`
	SenderAddress     string                `json:"sender_address" valid:",required"`
	SenderSignature   string                `json:"sender_signature,omitempty" valid:",required"`
	BlockHash         string                `json:"block_hash"`
	Messages          []*TransactionMessage `json:"messages"`
	BaseRND           int64                 `json:"base_rnd"`
}

// Validate block consensus message
func (m *BlockConsensusMessage) Validate() error {
	_, err := valid.ValidateStruct(m)
	return err
}

// Transaction message
type TransactionMessage struct {
	MessageHash  string      `json:"message_hash" valid:",required"`
	Tx           *Tx         `json:"message" valid:",required"`
	Votes        [7]uint64   `json:"votes"`
	VmProcessed  bool        `json:"vm_processed"`
	VmResult     bool        `json:"vm_result"`
	VmResponse   interface{} `json:"vm_response"`
	BlockHash    string      `json:"block_hash"`
	BlockNumber  int         `json:"block_number"`
	ExecutedHash string      `json:"executed_hash"`
}

// transaction struct
type Tx struct {
	Type            string      `json:"type" valid:",required"`
	SenderAddress   string      `json:"sender_address" valid:",required"`
	Message         interface{} `json:"message"`
	SenderSignature string      `json:"sender_signature" valid:",required"`
	MessageHash     string      `json:"message_hash" valid:",required"`
	Nonce           int         `json:"nonce"`
}

type IP struct {
	Query string `json:"query"`
}
