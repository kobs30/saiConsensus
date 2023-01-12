package models

import (
	"context"
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
func (m *Tx) GetHash() (string, error) {
	b, err := json.Marshal(&Tx{
		SenderAddress: m.SenderAddress,
		Message:       m.Message,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
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

func CreateTxMsg(ctx context.Context, argStr string) (*TransactionMessage, error) {
	transactionMessage := &TransactionMessage{
		Tx: &Tx{
			Type:          TransactionMsgType,
			SenderAddress: ctx.Value("saiBTCKeys").(*BtcKeys).Address,
			Message:       argStr,
			Nonce:         int(time.Now().Unix()),
		},
	}

	hash, err := transactionMessage.Tx.GetHash()
	if err != nil {
		return nil, fmt.Errorf("handlers  - createTx - count tx message hash: %w", err)
	}
	transactionMessage.Tx.MessageHash = hash

	btcResp, err := SignMessage(transactionMessage, ctx.Value("saiBTC_address").(string), ctx.Value("saiBTCKeys").(*BtcKeys).Private)
	if err != nil {
		return nil, fmt.Errorf("handlers  - createTx - sign tx message: %w", err)
	}
	transactionMessage.Tx.SenderSignature = btcResp.Signature
	transactionMessage.MessageHash = hash

	return transactionMessage, nil
}
