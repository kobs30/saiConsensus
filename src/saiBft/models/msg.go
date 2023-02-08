package models

import (
	"errors"
)

const (
	BlockConsensusMsgType = "blockConsensus"
	ConsensusMsgType      = "consensus"
	TransactionMsgType    = "message"
	RNDMessageType        = "rnd"
)

type Msger interface {
	SetHash() error
	SignMessage(string, string) error
	HashAndSign(string, string) error
}

// main message, which comes from saiP2P service
type P2pMsg struct {
	Signature string      `json:"signature"`
	Data      interface{} `json:"data"` // can be different type of messages here
}

// struct to handle tx from handler
type TxFromHandler struct {
	Tx        *Tx
	IsFromCli bool
}

// Parameters of blockchain
type Parameters struct {
	Validators     []string `json:"validators"`
	IsBoorstrapped bool     `json:"is_bootstrapped"`
}

// detect message type from saiP2p data input
func DetectMsgTypeFromMap(m map[string]interface{}) (string, error) {
	if _, ok := m["block_number"]; ok {
		return ConsensusMsgType, nil
	} else if _, ok := m["block_hash"]; ok {
		return BlockConsensusMsgType, nil
	} else if _, ok := m["message"]; ok {
		return TransactionMsgType, nil
	} else {
		return "", errors.New("unknown msg type")
	}
}

// struct for sync sleep value via consensus msg
type SyncConsensusKey struct {
	BlockNumber int
	Round       int
	Time        int64
}
