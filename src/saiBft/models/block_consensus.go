package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	valid "github.com/asaskevich/govalidator"
)

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
