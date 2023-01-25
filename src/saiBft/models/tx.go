package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	valid "github.com/asaskevich/govalidator"
)

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

// Validate transaction message
func (m *TransactionMessage) Validate() error {
	_, err := valid.ValidateStruct(m)
	return err
}

// Hashing tx
func (m *Tx) SetHash() error {
	b, err := json.Marshal(&Tx{
		SenderAddress: m.SenderAddress,
		Message:       m.Message,
		Nonce:         m.Nonce,
		Type:          m.Type,
	})
	if err != nil {
		return err
	}

	hash := sha256.Sum256(b)
	m.MessageHash = hex.EncodeToString(hash[:])
	return nil
}

// get executed hash of TransactionMessage
func (m *TransactionMessage) GetExecutedHash() error {
	b, err := json.Marshal(&TransactionMessage{
		Tx:         m.Tx,
		VmResult:   m.VmResult,
		VmResponse: m.VmResponse,
	})
	if err != nil {
		return err
	}

	hash := sha256.Sum256(b)
	m.ExecutedHash = hex.EncodeToString(hash[:])
	return nil
}

func CreateTxMsg(keys *BtcKeys, address string, argStr []string) (*TransactionMessage, error) {
	transactionMessage := &TransactionMessage{
		Tx: &Tx{
			Type:          TransactionMsgType,
			SenderAddress: keys.Address,
			Message:       argStr,
			Nonce:         int(time.Now().Unix()),
		},
	}

	err := transactionMessage.HashAndSign(address, keys.Private)
	if err != nil {
		return nil, err
	}

	return transactionMessage, nil
}

func (m *TransactionMessage) SignMessage(address, privateKey string) error {
	data, err := json.Marshal(&Tx{
		SenderAddress: m.Tx.SenderAddress,
		Message:       m.Tx.Message,
		Nonce:         m.Tx.Nonce,
		Type:          m.Tx.Type,
	})
	if err != nil {
		return err
	}
	preparedString := fmt.Sprintf("method=signMessage&p=%s&message=%s", privateKey, string(data))
	signature, err := getBTCResponse(preparedString, address)
	if err != nil {
		return err
	}
	m.Tx.SenderSignature = signature
	return nil
}

//fill hash and signature in tx message
func (m *TransactionMessage) HashAndSign(address, privateKey string) error {
	err := m.Tx.SetHash()
	if err != nil {
		return err
	}
	m.MessageHash = m.Tx.MessageHash
	err = m.SignMessage(address, privateKey)
	if err != nil {
		return err
	}
	return nil
}
