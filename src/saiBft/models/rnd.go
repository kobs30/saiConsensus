package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	valid "github.com/asaskevich/govalidator"
)

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
func (m *RNDMessage) SetHash() error {
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
		return err
	}

	hash := sha256.Sum256(b)
	m.Hash = hex.EncodeToString(hash[:])
	return nil
}

func (m *RNDMessage) SignMessage(address, privateKey string) error {
	data, err := json.Marshal(&RNDMessage{
		SenderAddress: m.SenderAddress,
		BlockNumber:   m.BlockNumber,
		Round:         m.Round,
		Rnd:           m.Rnd,
		TxMsgHashes:   m.TxMsgHashes,
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

func (m *RNDMessage) HashAndSign(address, privateKey string) error {
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
