package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	valid "github.com/asaskevich/govalidator"
)

const (
	BlockConsensusMsgType = "blockConsensus"
	ConsensusMsgType      = "consensus"
	TransactionMsgType    = "message"
	RNDMessageType        = "rnd"
)

// main message, which comes from saiP2P service
type P2pMsg struct {
	Signature string      `json:"signature"`
	Data      interface{} `json:"data"` // can be different type of messages here
}

// Consensus message
type ConsensusMessage struct {
	Type          string   `json:"type" valid:",required"`
	SenderAddress string   `json:"sender_address" valid:",required"`
	BlockNumber   int      `json:"block_number" valid:",required"`
	Round         int      `json:"round" valid:",required"`
	Messages      []string `json:"messages"`
	Signature     string   `json:"signature" valid:",required"`
	Hash          string   `json:"hash" valid:",required"`
}

// Validate consensus message
func (m *ConsensusMessage) Validate() error {
	_, err := valid.ValidateStruct(m)
	return err
}

// Hashing consensus message
func (m *ConsensusMessage) GetHash() (string, error) {
	b, err := json.Marshal(&ConsensusMessage{
		Type:          m.Type,
		SenderAddress: m.SenderAddress,
		BlockNumber:   m.BlockNumber,
		Round:         m.Round,
		Messages:      m.Messages,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
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

// RND message
type RND struct {
	Votes   int         `json:"votes"`
	Message *RNDMessage `json:"message"`
}

// RND
type RNDMessage struct {
	Type            string   `json:"type" valid:",required"`
	SenderAddress   string   `json:"sender_address" valid:",required"`
	BlockNumber     int      `json:"block_number" valid:",required"`
	Round           int      `json:"round"`
	Rnd             int64    `json:"rnd"`
	Hash            string   `json:"hash" valid:",required"`
	TxMsgHashes     []string `json:"tx_hashes"`
	SenderSignature string   `json:"sender_signature" valid:",required"`
}

// Validate RND message
func (m *RNDMessage) Validate() error {
	_, err := valid.ValidateStruct(m)
	return err
}

// Hashing RND  message
func (m *RNDMessage) GetHash() (string, error) {
	b, err := json.Marshal(&RND{
		Message: &RNDMessage{
			SenderAddress: m.SenderAddress,
			BlockNumber:   m.BlockNumber,
			Round:         m.Round,
			Rnd:           m.Rnd,
			TxMsgHashes:   m.TxMsgHashes,
		},
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
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
