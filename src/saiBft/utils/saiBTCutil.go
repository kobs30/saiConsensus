package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/iamthe1whoknocks/bft/models"
)

// send and handle response from saiBTC service
func sendRequest(address string, requestBody io.Reader) ([]byte, error) {
	keysRequest, err := http.NewRequest("POST", address, requestBody)
	if err != nil {
		return nil, err
	}
	keysRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(keysRequest)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// get btc keys
func GetBtcKeys(address string) (*models.BtcKeys, []byte, error) {
	requestBody := strings.NewReader("method=generateBTC")

	body, err := sendRequest(address, requestBody)
	if err != nil {
		return nil, nil, err
	}

	keys := models.BtcKeys{}

	err = json.Unmarshal(body, &keys)
	if err != nil {
		return nil, nil, err
	}

	return &keys, body, nil
}

// validate message signature
func ValidateSignature(msg interface{}, address, SenderAddress, signature string) (err error) {
	b := make([]byte, 0)
	switch msg.(type) {
	case *models.Block:
		BCMsg := msg.(*models.Block)
		b, err = json.Marshal(&models.Block{
			Number:            BCMsg.Number,
			PreviousBlockHash: BCMsg.PreviousBlockHash,
			Messages:          BCMsg.Messages,
			SenderAddress:     BCMsg.SenderAddress,
		})
		if err != nil {
			return fmt.Errorf(" marshal blockConsensusMessage : %w", err)
		}
	case *models.ConsensusMessage:
		cMsg := msg.(*models.ConsensusMessage)
		b, err = json.Marshal(&models.ConsensusMessage{
			SenderAddress: cMsg.SenderAddress,
			BlockNumber:   cMsg.BlockNumber,
			Round:         cMsg.Round,
			Messages:      cMsg.Messages,
		})
		if err != nil {
			return fmt.Errorf("marshal ConsensusMessage : %w", err)
		}
	case *models.TransactionMessage:
		txMsg := msg.(*models.TransactionMessage)
		b, err = json.Marshal(&models.Tx{
			SenderAddress: txMsg.Tx.SenderAddress,
			Message:       txMsg.Tx.Message,
		})
		if err != nil {
			return fmt.Errorf("marshal TransactionMessage : %w", err)
		}
	case *models.RNDMessage:
		rndMsg := msg.(*models.RNDMessage)
		b, err = json.Marshal(&models.RNDMessage{
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

	body, err := sendRequest(address, requestBody)
	if err != nil {
		return fmt.Errorf("sendRequest to saiBTC : %w", err)
	}

	resp := models.ValidateSignatureResponse{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return fmt.Errorf("unmarshal response from saiBTC  : %w", err)
	}
	if resp.Signature == "valid" {
		return nil
	}
	return errors.New("Signature is not valid\n")
}

// saiBTC sign message method
func SignMessage(msg interface{}, address, privateKey string) (resp *models.SignMessageResponse, err error) {
	var preparedString string
	switch msg.(type) {
	case *models.Block:
		BCMsg := msg.(*models.Block)
		data, err := json.Marshal(&models.Block{
			Number:            BCMsg.Number,
			PreviousBlockHash: BCMsg.PreviousBlockHash,
			Messages:          BCMsg.Messages,
			SenderAddress:     BCMsg.SenderAddress,
		})
		if err != nil {
			return nil, err
		}
		preparedString = fmt.Sprintf("method=signMessage&p=%s&message=%s", privateKey, string(data))
	case *models.ConsensusMessage:
		cMsg := msg.(*models.ConsensusMessage)
		data, err := json.Marshal(&models.ConsensusMessage{
			SenderAddress: cMsg.SenderAddress,
			BlockNumber:   cMsg.BlockNumber,
			Round:         cMsg.Round,
			Messages:      cMsg.Messages,
		})
		if err != nil {
			return nil, err
		}
		preparedString = fmt.Sprintf("method=signMessage&p=%s&message=%s", privateKey, string(data))
	case *models.TransactionMessage:
		TxMsg := msg.(*models.TransactionMessage)
		data, err := json.Marshal(&models.Tx{
			SenderAddress: TxMsg.Tx.SenderAddress,
			Message:       TxMsg.Tx.Message,
		})
		if err != nil {
			return nil, err
		}
		preparedString = fmt.Sprintf("method=signMessage&p=%s&message=%s", privateKey, string(data))
	case *models.Tx:
		TxMsg := msg.(*models.Tx)
		data, err := json.Marshal(&models.Tx{
			SenderAddress: TxMsg.SenderAddress,
			Message:       TxMsg.Message,
		})
		if err != nil {
			return nil, err
		}
		preparedString = fmt.Sprintf("method=signMessage&p=%s&message=%s", privateKey, string(data))
	case *models.RNDMessage:
		rndMsg := msg.(*models.RNDMessage)
		data, err := json.Marshal(&models.RNDMessage{
			SenderAddress: rndMsg.SenderAddress,
			BlockNumber:   rndMsg.BlockNumber,
			Round:         rndMsg.Round,
			Rnd:           rndMsg.Rnd,
			TxMsgHashes:   rndMsg.TxMsgHashes,
		})
		if err != nil {
			return nil, err
		}
		preparedString = fmt.Sprintf("method=signMessage&p=%s&message=%s", privateKey, string(data))
	default:
		return nil, errors.New("unknown type of message")
	}
	requestBody := strings.NewReader(preparedString)

	body, err := sendRequest(address, requestBody)
	if err != nil {
		return nil, fmt.Errorf("sendRequest to saiBTC : %w", err)
	}

	response := models.SignMessageResponse{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response from saiBTC  : %w\n response body : %s", err, string(body))
	}

	return &response, nil
}

func GetBTCkeys(fileStr, saiBTCaddress string) (*models.BtcKeys, error) {
	file, err := os.OpenFile(fileStr, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("processing - open key btc file - %w", err)
	}

	data, err := ioutil.ReadFile(fileStr)
	if err != nil {
		fmt.Errorf("processing - read key btc file - %w", err)
	}
	btcKeys := models.BtcKeys{}

	err = json.Unmarshal(data, &btcKeys)
	if err != nil {
		fmt.Errorf("get btc keys - error unmarshal from file - %w", err)
		btcKeys, body, err := GetBtcKeys(saiBTCaddress)
		if err != nil {
			fmt.Errorf("get btc keys - get btc keys - %w", err)
		}
		_, err = file.Write(body)
		if err != nil {
			fmt.Errorf("get btc keys - write btc keys to file - %w", err)
		}
		return btcKeys, nil
	} else {
		err = btcKeys.Validate()
		if err != nil {
			btcKeys, body, err := GetBtcKeys(saiBTCaddress)
			if err != nil {
				fmt.Errorf("get btc keys - get btc keys - %w", err)
			}
			_, err = file.Write(body)
			if err != nil {
				fmt.Errorf("get btc keys - write btc keys to file - %w", err)
				return nil, err
			}
			return btcKeys, nil
		}
		return &btcKeys, nil
	}
}
