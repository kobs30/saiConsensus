package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	valid "github.com/asaskevich/govalidator"
)

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
func (m *ConsensusMessage) SetHash() error {
	b, err := json.Marshal(&ConsensusMessage{
		Type:          m.Type,
		SenderAddress: m.SenderAddress,
		BlockNumber:   m.BlockNumber,
		Round:         m.Round,
		Messages:      m.Messages,
	})
	if err != nil {
		return err
	}

	hash := sha256.Sum256(b)
	m.Hash = hex.EncodeToString(hash[:])
	return nil
}

func (m *ConsensusMessage) SignMessage(address, privateKey string) error {
	data, err := json.Marshal(&ConsensusMessage{
		SenderAddress: m.SenderAddress,
		BlockNumber:   m.BlockNumber,
		Round:         m.Round,
		Messages:      m.Messages,
	})
	if err != nil {
		return err
	}
	preparedString := fmt.Sprintf("method=signMessage&p=%s&message=%s", privateKey, string(data))
	signature, err := getBTCResponse(preparedString, address)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *ConsensusMessage) HashAndSign(address, privateKey string) error {
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
