package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

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

// BlockConsensus message
type BlockConsensusMessage struct {
	BlockHash  string   `json:"block_hash" valid:",required"`
	Votes      int      `json:"votes"` // additional field, which was not added by Valeriy
	Block      *Block   `json:"block" valid:",required"`
	Count      int      `json:"-"` // extended value for consensus while get missed blocks from p2p services
	Signatures []string `json:"voted_signatures"`
}

type Block struct {
	Type              string                         `json:"type" valid:",required"`
	Number            int                            `json:"number" valid:",required"`
	PreviousBlockHash string                         `json:"prev_block_hash" valid:",required"`
	SenderAddress     string                         `json:"sender_address" valid:",required"`
	SenderSignature   string                         `json:"sender_signature,omitempty" valid:",required"`
	BlockHash         string                         `json:"block_hash"`
	Messages          map[string]*TransactionMessage `json:"messages"`
	BaseRND           int64                          `json:"base_rnd"`
}

// Validate block consensus message
func (m *Block) Validate() error {
	_, err := valid.ValidateStruct(m)
	return err
}

// Hashing block  message
func (m *Block) GetHash() (string, error) {
	b, err := json.Marshal(&Block{
		Number:            m.Number,
		PreviousBlockHash: m.PreviousBlockHash,
		Messages:          m.Messages,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
}

type GetBlockMsg struct {
	BCMessage        *BlockConsensusMessage `json:"block_consensus"`
	EqualHashesCount int
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
