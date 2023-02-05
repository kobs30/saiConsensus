package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	valid "github.com/asaskevich/govalidator"
	"github.com/iamthe1whoknocks/bft/utils"
)

// btc keys got from saiBTC
type BtcKeys struct {
	Private string `json:"Private" valid:",required"`
	Public  string `json:"Public" valid:",required"`
	Address string `json:"Address" valid:",required"`
}

// Validate block consensus message
func (m *BtcKeys) Validate() error {
	_, err := valid.ValidateStruct(m)
	return err
}

// validate signature response from saiBTC
type ValidateSignatureResponse struct {
	Address   string `json:"address"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// sign signature response from saiBTC
type SignMessageResponse struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// validate message signature
func ValidateSignature(msg interface{}, address, SenderAddress, signature string) (err error) {
	b := make([]byte, 0)
	switch msg.(type) {
	case *Block:
		BCMsg := msg.(*Block)
		b, err = json.Marshal(&Block{
			Number:            BCMsg.Number,
			PreviousBlockHash: BCMsg.PreviousBlockHash,
			Messages:          BCMsg.Messages,
			SenderAddress:     BCMsg.SenderAddress,
		})
		if err != nil {
			return fmt.Errorf(" marshal blockConsensusMessage : %w", err)
		}
	case *ConsensusMessage:
		cMsg := msg.(*ConsensusMessage)
		b, err = json.Marshal(&ConsensusMessage{
			SenderAddress: cMsg.SenderAddress,
			BlockNumber:   cMsg.BlockNumber,
			Round:         cMsg.Round,
			Messages:      cMsg.Messages,
		})
		if err != nil {
			return fmt.Errorf("marshal ConsensusMessage : %w", err)
		}
	case *TransactionMessage:
		txMsg := msg.(*TransactionMessage)
		b, err = json.Marshal(&Tx{
			SenderAddress: txMsg.Tx.SenderAddress,
			Message:       txMsg.Tx.Message,
			Nonce:         txMsg.Tx.Nonce,
			Type:          txMsg.Tx.Type,
		})
		if err != nil {
			return fmt.Errorf("marshal TransactionMessage : %w", err)
		}
	case *RNDMessage:
		rndMsg := msg.(*RNDMessage)
		b, err = json.Marshal(&RNDMessage{
			SenderAddress: rndMsg.SenderAddress,
			BlockNumber:   rndMsg.BlockNumber,
			Round:         rndMsg.Round,
			Rnd:           rndMsg.Rnd,
			TxMsgHashes:   rndMsg.TxMsgHashes,
		})
	default:
		return fmt.Errorf("unknown type of message, incoming type : %+v\n", reflect.TypeOf(msg))
	}
	preparedString := fmt.Sprintf("method=validateSignature&a=%s&signature=%s&message=%s", SenderAddress, signature, string(b))
	requestBody := strings.NewReader(preparedString)

	body, err := utils.SendRequest(address, requestBody)
	if err != nil {
		return fmt.Errorf("sendRequest to saiBTC : %w", err)
	}

	resp := ValidateSignatureResponse{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return fmt.Errorf("unmarshal response from saiBTC  : %w", err)
	}
	if resp.Signature == "valid" {
		return nil
	}
	return errors.New("Signature is not valid\n")
}

func getBTCResponse(preparedString string, address string) (signature string, err error) {
	requestBody := strings.NewReader(preparedString)

	body, err := utils.SendRequest(address, requestBody)
	if err != nil {
		return "", fmt.Errorf("sendRequest to saiBTC : %w", err)
	}

	response := SignMessageResponse{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("unmarshal response from saiBTC  : %w\n response body : %s", err, string(body))
	}

	return response.Signature, nil
}

// get btc keys
func GetBtcKeys(address string) (*BtcKeys, []byte, error) {
	requestBody := strings.NewReader("method=generateBTC")

	body, err := utils.SendRequest(address, requestBody)
	if err != nil {
		return nil, nil, err
	}

	keys := BtcKeys{}

	err = json.Unmarshal(body, &keys)
	if err != nil {
		return nil, nil, err
	}

	return &keys, body, nil
}

func GetBTCkeys(fileStr, saiBTCaddress string) (*BtcKeys, error) {
	file, err := os.OpenFile(fileStr, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("processing - open key btc file - %w", err)
	}

	data, err := ioutil.ReadFile(fileStr)
	if err != nil {
		err = fmt.Errorf("processing - read key btc file - %w", err)
		return nil, err
	}
	btcKeys := BtcKeys{}

	err = json.Unmarshal(data, &btcKeys)
	if err != nil {
		btcKeys, body, err := GetBtcKeys(saiBTCaddress)
		if err != nil {
			err = fmt.Errorf("get btc keys - get btc keys - %w", err)
			return nil, err
		}
		_, err = file.Write(body)
		if err != nil {
			err = fmt.Errorf("get btc keys - write btc keys to file - %w", err)
			return nil, err
		}
		return btcKeys, nil
	} else {
		err = btcKeys.Validate()
		if err != nil {
			btcKeys, body, err := GetBtcKeys(saiBTCaddress)
			if err != nil {
				err = fmt.Errorf("get btc keys - get btc keys - %w", err)
				return nil, err
			}
			_, err = file.Write(body)
			if err != nil {
				err = fmt.Errorf("get btc keys - write btc keys to file - %w", err)
				return nil, err
			}
			return btcKeys, nil
		}
		return &btcKeys, nil
	}
}
