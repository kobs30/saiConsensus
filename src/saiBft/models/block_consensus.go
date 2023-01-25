package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

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
	PreviousBlockHash string                `json:"prev_block_hash"`
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
func (m *Block) SetHash() error {
	b, err := json.Marshal(&Block{
		Number:            m.Number,
		PreviousBlockHash: m.PreviousBlockHash,
		Messages:          m.Messages,
	})
	if err != nil {
		return err
	}

	hash := sha256.Sum256(b)
	m.BlockHash = hex.EncodeToString(hash[:])
	return nil
}

func (m *Block) SignMessage(address, privateKey string) error {
	data, err := json.Marshal(&Block{
		Number:            m.Number,
		PreviousBlockHash: m.PreviousBlockHash,
		Messages:          m.Messages,
		SenderAddress:     m.SenderAddress,
	})
	if err != nil {
		return err
	}
	preparedString := fmt.Sprintf("method=signMessage&p=%s&message=%s", privateKey, string(data))
	signature, err := getBTCResponse(preparedString, address)
	if err != nil {
		return err
	}
	m.SenderSignature = signature
	return nil
}

func (m *Block) HashAndSign(address, privateKey string) error {
	err := m.SetHash()
	if err != nil {
		return err
	}
	err = m.SignMessage(address, privateKey)
	if err != nil {
		return err
	}
	return nil
}
