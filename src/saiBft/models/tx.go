package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

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
	Type            string `json:"type" valid:",required"`
	SenderAddress   string `json:"sender_address" valid:",required"`
	Message         string `json:"message" valid:",required"`
	SenderSignature string `json:"sender_signature" valid:",required"`
	MessageHash     string `json:"message_hash" valid:",required"`
	Nonce           int    `json:"nonce"`
}

// tx message struct
type TxMessage struct {
	Method string   `json:"method" valid:",required"`
	Params []string `json:"params" valid:",required"`
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

// func (m *TransactionMessage) CreateTxMsg(txMsg []byte) (*TransactionMessage, error) {
// 	saiBtcAddress, ok := Service.GlobalService.Configuration["saiBTC_address"].(string)
// 	if !ok {
// 		Service.GlobalService.Logger.Fatal("wrong type of saiBTC_address value in config")
// 	}

// 	btcKeys := utils.GetBTCkeys
// 	Service.BTCkeys = btckeys

// 	txMsgBytes, err := json.Marshal(txMsg)
// 	if err != nil {
// 		return nil, fmt.Errorf("handlers - tx  -  marshal tx msg: %w", err)
// 	}
// 	transactionMessage := &TransactionMessage{
// 		Tx: &Tx{
// 			Type:          TransactionMsgType,
// 			SenderAddress: Service.BTCkeys.Address,
// 			Message:       string(txMsgBytes),
// 		},
// 	}

// 	hash, err := transactionMessage.Tx.GetHash()
// 	if err != nil {
// 		Service.GlobalService.Logger.Error("handlers  - tx - count tx message hash", zap.Error(err))
// 		return nil, fmt.Errorf("handlers  - tx - count tx message hash: %w", err)
// 	}
// 	transactionMessage.Tx.MessageHash = hash

// 	btcResp, err := utils.SignMessage(transactionMessage, saiBtcAddress, Service.BTCkeys.Private)
// 	if err != nil {
// 		Service.GlobalService.Logger.Error("handlers  - tx - sign tx message", zap.Error(err))
// 		return nil, fmt.Errorf("handlers  - tx - sign tx message: %w", err)
// 	}
// 	transactionMessage.Tx.SenderSignature = btcResp.Signature

// 	saiP2Paddress, ok := Service.GlobalService.Configuration["saiP2P_address"].(string)
// 	if !ok {
// 		Service.GlobalService.Logger.Fatal("processing - wrong type of saiP2P address value from config")
// 	}
// }
