package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/iamthe1whoknocks/bft/models"
	"github.com/iamthe1whoknocks/bft/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	errNotEnoughVotes = errors.New("tx msg with certain number of votes was not found")
)

const (
	blockchainCol      = "Blockchain"
	MessagesPoolCol    = "MessagesPool"
	ConsensusPoolCol   = "ConsensusPool"
	BlockCandidatesCol = "BlockCandidates"
	ParametersCol      = "Parameters"
	RndMessagesPoolCol = "RNDPool"
	maxRoundNumber     = 7
	btcKeyFile         = "btc_keys.json"
)

type VmResponse struct {
	VMProcessed bool        `json:"vm_processed"`
	VMResult    bool        `json:"vm_result"`
	VmResponse  interface{} `json:"vm_response"`
}

// main process of blockchain
func (s *InternalService) Processing() {

	s.GlobalService.Logger.Debug("starting processing") //DEBUG

	// for tests
	//btcKeys1, _ := s.getBTCkeys("btc_keys3.json", SaiBTCaddress)
	// btcKeys2, _ := s.getBTCkeys("btc_keys2.json", SaiBTCaddress)
	// btcKeys3, _ := s.getBTCkeys("btc_keys1.json", SaiBTCaddress)
	//s.TrustedValidators = append(s.TrustedValidators, s.BTCkeys.Address)

	err := s.setValidators(s.GlobalService.GetConfig(SaiStorageToken, "").String())
	if err != nil {
		s.GlobalService.Logger.Fatal("process - check validator state", zap.Error(err))
	}
	s.GlobalService.Logger.Debug("get validators", zap.Strings("validators", s.Validators)) //DEBUG

	for _, validator := range s.Validators {
		if validator == s.BTCkeys.Address {
			s.IsValidator = true
			s.IsInitialized = true
		}
	}

	s.GlobalService.Logger.Debug("node mode", zap.Bool("is_validator", s.IsValidator)) //DEBUG

	for !s.IsValidator {
		select {
		case data := <-s.InitialSignalCh:
			s.Validators = append(s.Validators, data.(string))

			err, _ := s.Storage.Update(ParametersCol, bson.M{}, bson.M{"validators": s.Validators}, s.GlobalService.GetConfig(SaiStorageToken, "").String())
			if err != nil {
				Service.GlobalService.Logger.Error("listenFromSaiP2P - initial block consensus msg - put to storage", zap.Error(err))
				continue
			}

			s.IsValidator = true
			s.GlobalService.Logger.Debug("node was add as validator by incoming block consensus msg")
			break
		}
	}

	//TEST transaction &consensus messages
	//s.saveTestTx(s.GlobalService.GetConfig(SaiBTCaddress, "").String(), s.GlobalService.GetConfig(SaiStorageToken, "").String(), s.GlobalService.GetConfig(SaiP2pAddress, "").String())
	//s.saveTestTx2(s.GlobalService.GetConfig(SaiBTCaddress, "").String(), s.GlobalService.GetConfig(SaiStorageToken, "").String(), s.GlobalService.GetConfig(SaiP2pAddress, "").String())

	if s.GlobalService.GetConfig(SaiDuplicateStorageRequests, false).Bool() {
		go s.duplicateRequests()
	}

	for {
	startLoop:
		round := 0
		s.GlobalService.Logger.Debug("start loop,round = 0") // DEBUG

		// get last block from blockchain collection or create initial block
		block, err := s.getLastBlockFromBlockChain(s.GlobalService.GetConfig(SaiStorageToken, "").String(), s.GlobalService.GetConfig(SaiBTCaddress, "").String())
		if err != nil {
			continue
		}

		rnd, err := s.rndProcessing(s.GlobalService.GetConfig(SaiBTCaddress, "").String(), s.GlobalService.GetConfig(SaiP2pAddress, "").String(), s.GlobalService.GetConfig(SaiStorageToken, "").String(), block.Block.Number)
		if err != nil {
			s.GlobalService.Logger.Error("process - process rnd", zap.Error(err))
			continue
		}
		s.GlobalService.Logger.Debug("processing - got rnd", zap.Int("block_number", block.Block.Number), zap.Int64("rnd", rnd))

	checkRound:
		s.GlobalService.Logger.Sugar().Debugf("ROUND = %d", round) //DEBUG
		if round == 0 {
			// get messages with votes = 0
			transactions, err := s.getZeroVotedTransactions(s.GlobalService.GetConfig(SaiStorageToken, "").String(), block.Block.Number)
			if err != nil {
				s.GlobalService.Logger.Error("process - round == 0 - get zero-voted tx messages", zap.Error(err))
			}
			consensusMsg := &models.ConsensusMessage{
				Type:          models.ConsensusMsgType,
				SenderAddress: s.BTCkeys.Address,
				BlockNumber:   block.Block.Number,
				Round:         round,
			}
			// validate/execute each tx msg, update hash and votes
			if len(transactions) != 0 {
				for _, tx := range transactions {
					err = s.validateExecuteTransactionMsg(tx, rnd, block.Block.Number, s.GlobalService.GetConfig(SaiBTCaddress, "").String(), s.GlobalService.GetConfig(SaiVM1Address, "").String(), s.GlobalService.GetConfig(SaiStorageToken, "").String())
					if err != nil {
						continue
					}
					consensusMsg.Messages = append(consensusMsg.Messages, tx.ExecutedHash)
				}
			}

			consensusMsg.Round = round + 1

			err = consensusMsg.HashAndSign(s.GlobalService.GetConfig(SaiBTCaddress, "").String(), s.BTCkeys.Private)
			if err != nil {
				s.GlobalService.Logger.Error("process - round == 0 - hash and sign consensus msg", zap.Error(err))
				goto startLoop
			}

			err, _ = s.Storage.Put("ConsensusPool", consensusMsg, s.GlobalService.GetConfig(SaiStorageToken, "").String())
			if err != nil {
				s.GlobalService.Logger.Error("process - round == 0 - put consensus to ConsensusPool collection", zap.Error(err))
				goto startLoop
			}

			err = s.broadcastMsg(consensusMsg, s.GlobalService.GetConfig(SaiBTCaddress, "").String(), false)
			if err != nil {
				s.GlobalService.Logger.Error("process - round==0 - broadcast consensus message", zap.Error(err))
				goto startLoop
			}

			time.Sleep(time.Duration(s.GlobalService.GetConfig("sleep", 5).Int()) * time.Second)
			round++
			goto checkRound

		} else {

			var syncValue int
			//clean block candidate collection at round = 5
			if round == 5 {
				err := s.removeCandidates(s.GlobalService.GetConfig(SaiStorageToken, "").String())
				if err != nil {
					s.GlobalService.Logger.Error("process - clean blockCandidates collection", zap.Error(err))
					continue
				}
			}

			// get consensus messages for the round
			msgs, err := s.getConsensusMsgForTheRound(round, block.Block.Number, s.GlobalService.GetConfig(SaiStorageToken, "").String())
			if err != nil {
				goto startLoop
			}
			for _, msg := range msgs {
				// check if consensus message sender is from trusted validators list
				err = checkConsensusMsgSender(s.Validators, msg)
				if err != nil {
					s.GlobalService.Logger.Error("process - round != 0 - check consensus message sender", zap.Error(err))
					continue
				}

				//s.GlobalService.Logger.Debug("Consensus message transactions", zap.Strings("msgs", msg.Messages), zap.String("hash", msg.Hash)) //DEBUG

				// update votes for each tx message from consensusMsg
				for _, txMsgHash := range msg.Messages {
					err, result := s.Storage.Get("MessagesPool", bson.M{"executed_hash": txMsgHash, "block_hash": ""}, bson.M{}, s.GlobalService.GetConfig(SaiStorageToken, "").String())
					if err != nil {
						s.GlobalService.Logger.Error("process - get msg from consensus msg from storage", zap.Error(err))
						continue
					}

					if len(result) == 2 {
						continue
					}
					err = s.updateTxMsgVotes(txMsgHash, s.GlobalService.GetConfig(SaiStorageToken, "").String(), round)
					if err != nil {
						continue
					}
				}
			}

			// get messages with votes>=(roundNumber*10)%
			txMsgs, err := s.getTxMsgsWithCertainNumberOfVotes(s.GlobalService.GetConfig(SaiStorageToken, "").String(), round)
			if err != nil {
				goto startLoop
			}

			if round < maxRoundNumber-1 {
				newConsensusMsg := &models.ConsensusMessage{
					Type:          models.ConsensusMsgType,
					SenderAddress: Service.BTCkeys.Address,
					BlockNumber:   block.Block.Number,
				}

				for _, txMsg := range txMsgs {
					newConsensusMsg.Messages = append(newConsensusMsg.Messages, txMsg.ExecutedHash)
				}
				newConsensusMsg.Round = round + 1

				err = newConsensusMsg.HashAndSign(s.GlobalService.GetConfig(SaiBTCaddress, "").String(), s.BTCkeys.Private)
				if err != nil {
					s.GlobalService.Logger.Error("process - hash and sign consensus msg", zap.Error(err))
					goto startLoop
				}

				err, _ = s.Storage.Put("ConsensusPool", newConsensusMsg, s.GlobalService.GetConfig(SaiStorageToken, "").String())
				if err != nil {
					s.GlobalService.Logger.Error("process -  put consensus to ConsensusPool collection", zap.Error(err))
					goto startLoop
				}

				s.SyncConsensus.Mu.Lock()
				//s.SyncConsensusMap[string(newConsensusMsg.BlockNumber)+"+"+string(newConsensusMsg.Round)]++
				s.SyncConsensus.Storage[models.SyncConsensusKey{
					BlockNumber: newConsensusMsg.BlockNumber,
					Round:       newConsensusMsg.Round,
				}]++

				s.SyncConsensus.Mu.Unlock()

				syncValue = Service.syncSleep(newConsensusMsg)
				s.GlobalService.Logger.Debug("process - sync sleep", zap.Int("sync value", syncValue))

				err = s.broadcastMsg(newConsensusMsg, s.GlobalService.GetConfig(SaiP2pAddress, "").String(), false)
				if err != nil {
					goto startLoop
				}
			}
			s.GlobalService.Logger.Debug("process - sync sleep", zap.Int("round", round), zap.Int("sync value", syncValue))

			if round == 6 {
				syncValue = 1
			}
			round = round + syncValue
			//round++

			time.Sleep(time.Duration(s.GlobalService.GetConfig("sleep", 5).Int()) * time.Second)

			if round < maxRoundNumber {
				goto checkRound
			} else {
				s.GlobalService.Logger.Sugar().Debugf("ROUND = %d", round) //DEBUG

				s.clearSyncMap()

				newBlock, err := s.formAndSaveNewBlock(block, s.GlobalService.GetConfig(SaiBTCaddress, "").String(), s.GlobalService.GetConfig(SaiStorageToken, "").String(), txMsgs, rnd)
				if err != nil {
					goto startLoop
				}

				err = s.broadcastMsg(newBlock.Block, s.GlobalService.GetConfig(SaiP2pAddress, "").String(), false)
				if err != nil {
					goto startLoop
				}

				goto startLoop
			}
		}
	}
}

// get last block from blockchain collection
func (s *InternalService) getLastBlockFromBlockChain(storageToken string, SaiBTCaddress string) (*models.BlockConsensusMessage, error) {
	opts := options.Find().SetSort(bson.M{"block.number": -1}).SetLimit(1)
	err, result := s.Storage.Get(blockchainCol, bson.M{}, opts, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("handlers - process - processing - get last block", zap.Error(err))
		return nil, err
	}

	// empty get response returns '{}' in storage get method -> new block should be created
	if len(result) == 2 {
		block, err := s.createInitialBlock(SaiBTCaddress)
		if err != nil {
			s.GlobalService.Logger.Error("process - create initial block", zap.Error(err))
			return nil, err
		}
		return block, nil
	} else {
		blocks := make([]*models.BlockConsensusMessage, 0)
		data, err := utils.ExtractResult(result)
		if err != nil {
			Service.GlobalService.Logger.Error("process - get last block from blockchain - extract data from response", zap.Error(err))
			return nil, err
		}
		//		s.GlobalService.Logger.Sugar().Debugf("get last block data : %s", string(data))
		err = json.Unmarshal(data, &blocks)
		if err != nil {
			s.GlobalService.Logger.Error("handlers - process - unmarshal result of last block from blockchain collection", zap.Error(err))
			return nil, err
		}
		block := blocks[0]
		block.Block.Number++
		s.GlobalService.Logger.Debug("Got last block from blockchain collection", zap.Int("block number", block.Block.Number)) //DEBUG

		return block, nil
	}
}

// create initial block
func (s *InternalService) createInitialBlock(address string) (block *models.BlockConsensusMessage, err error) {
	s.GlobalService.Logger.Sugar().Debugf("block not found, creating initial block") //DEBUG

	block = &models.BlockConsensusMessage{
		Block: &models.Block{
			Type:              models.BlockConsensusMsgType,
			Number:            1,
			SenderAddress:     s.BTCkeys.Address,
			PreviousBlockHash: "",
			Messages:          make([]*models.TransactionMessage, 0),
		},
	}

	err = block.Block.HashAndSign(address, s.BTCkeys.Private)
	if err != nil {
		return nil, err
	}

	block.BlockHash = block.Block.BlockHash

	return block, nil

}

// get messages with votes = 0
func (s *InternalService) getZeroVotedTransactions(storageToken string, blockNumber int) ([]*models.TransactionMessage, error) {
	err, result := s.Storage.Get(MessagesPoolCol, bson.M{"votes.0": 0, "block_number": blockNumber}, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - round = 0 - get messages with 0 votes", zap.Error(err))
		return nil, err
	}

	if len(result) == 2 {
		err = errors.New("no 0 voted messages found")
		s.GlobalService.Logger.Error("process - round = 0 - get messages with 0 votes", zap.Error(err))
		return nil, err
	}

	data, err := utils.ExtractResult(result)
	if err != nil {
		Service.GlobalService.Logger.Error("process - getZeroVotedTransactions - extract data from response", zap.String("data", string(result)), zap.Error(err))
		return nil, err
	}

	transactions := make([]*models.TransactionMessage, 0)

	err = json.Unmarshal(data, &transactions)
	if err != nil {
		s.GlobalService.Logger.Error("handlers - process - round = 0 - unmarshal result of messages with votes = 0", zap.Error(err))
		return nil, err
	}

	filteredTx := make([]*models.TransactionMessage, 0)

	for _, tx := range transactions {
		if tx.BlockHash == "" {
			filteredTx = append(filteredTx, tx)
		}
	}

	//sort tx messages by nonce, by hash
	sort.Slice(filteredTx, func(i, j int) bool {
		if filteredTx[i].Tx.Nonce == filteredTx[j].Tx.Nonce {
			return filteredTx[i].Tx.MessageHash < filteredTx[j].Tx.MessageHash
		}
		return filteredTx[i].Tx.Nonce < filteredTx[j].Tx.Nonce
	})

	s.GlobalService.Logger.Sugar().Debugf("Got transactions with votes = 0 : %+v", filteredTx) //DEBUG

	return filteredTx, nil
}

func (s *InternalService) callVM1(msg *models.TransactionMessage, rnd int64, block int, saiVM1Address string) *models.TransactionMessage {
	var parsed VmResponse
	response, ok := utils.SendHttpRequest(saiVM1Address, bson.M{
		"method": "execute",
		"data": bson.M{
			"block":   block,
			"rnd":     rnd,
			"tx":      msg.Tx,
			"message": msg.Tx.Message,
		},
	})

	if ok {
		err := json.Unmarshal(response.([]byte), &parsed)
		if err != nil {
			s.GlobalService.Logger.Error("process - callVM1 - parse vm response error 1", zap.Error(err))
			return msg
		}

		msg.VmProcessed = parsed.VMProcessed
		msg.VmResult = parsed.VMResult
		msg.VmResponse = parsed.VmResponse
	} else {

	}

	return msg
}

// validate/execute each message, update message and hash and vote for valid messages
func (s *InternalService) validateExecuteTransactionMsg(msg *models.TransactionMessage, rnd int64, block int, saiBTCaddress, saiVM1address, storageToken string) error {
	s.GlobalService.Logger.Sugar().Debugf("Handling transaction : %+v", msg) //DEBUG

	err := models.ValidateSignature(msg, saiBTCaddress, msg.Tx.SenderAddress, msg.Tx.SenderSignature)
	if err != nil {
		s.GlobalService.Logger.Error("process - ValidateExecuteTransactionMsg - validate tx msg signature", zap.Error(err))
		return err
	}

	msg = s.callVM1(msg, rnd, block, saiVM1address)

	err = msg.GetExecutedHash()
	if err != nil {
		s.GlobalService.Logger.Error("process - ValidateExecuteTransactionMsg - get executed hash", zap.Error(err))
		return err
	}

	msg.Votes[0]++
	filter := bson.M{"message_hash": msg.MessageHash}
	update := bson.M{"votes": msg.Votes, "vm_processed": msg.VmProcessed, "vm_result": msg.VmResult, "vm_response": msg.VmResponse, "executed_hash": msg.ExecutedHash}
	err, _ = s.Storage.Update("MessagesPool", filter, update, storageToken)
	if err != nil {
		Service.GlobalService.Logger.Error("process - ValidateExecuteTransactionMsg - update transactions in storage", zap.Error(err))
		return err
	}
	return nil

}

// check if consensus message sender is not from validators list
func checkConsensusMsgSender(validators []string, msg *models.ConsensusMessage) error {
	for _, validator := range validators {
		if msg.SenderAddress == validator {
			return nil
		}
	}
	return fmt.Errorf("Consensus message sender is not from validators list, validators : %s, sender : %s", validators, msg.SenderAddress)
}

// get consensus messages for the round
func (s *InternalService) getConsensusMsgForTheRound(round, blockNumber int, storageToken string) ([]*models.ConsensusMessage, error) {
	err, result := s.Storage.Get(ConsensusPoolCol, bson.M{"round": round, "block_number": blockNumber}, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - round != 0 - get messages for specified round", zap.Error(err))
		return nil, err
	}

	if len(result) == 2 {
		err = fmt.Errorf("no consensusMsg found for round : %d", round)
		s.GlobalService.Logger.Error("process - get consensusMsg for round", zap.Int("round", round), zap.Error(err))
		return nil, err
	}

	data, err := utils.ExtractResult(result)
	if err != nil {
		Service.GlobalService.Logger.Error("process - getConsensusMsgForTheRound - extract data from response", zap.Error(err))
		return nil, err
	}

	msgs := make([]*models.ConsensusMessage, 0)

	err = json.Unmarshal(data, &msgs)
	if err != nil {
		s.GlobalService.Logger.Error("process - round != 0 - unmarshal result of consensus messages for specified round", zap.Error(err))
		return nil, err
	}

	return msgs, nil
}

// broadcast messages to connected nodes
func (s *InternalService) broadcastMsg(msg interface{}, SaiP2pAddress string, force bool) error {
	s.GlobalService.Logger.Debug("process - broadcastMsg - starting broadcast", zap.Any("msg", msg))

	if !s.IsValidator && !force {
		s.GlobalService.Logger.Debug("process - broadcastMsg - not a validator")
		return nil
	}

	param := url.Values{}
	data, err := json.Marshal(msg)

	if err != nil {
		s.GlobalService.Logger.Error("process - round != 0 - broadcastMsg - marshal msg", zap.Error(err))
		return err
	}

	param.Add("message", string(data))
	postRequest, err := http.NewRequest("POST", SaiP2pAddress+"/Send_message", bytes.NewBufferString(param.Encode()))
	if err != nil {
		s.GlobalService.Logger.Error("process - round != 0 - broadcastMsg - create post request", zap.Error(err))
		return err
	}

	postRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(postRequest)
	if err != nil {
		s.GlobalService.Logger.Error("process - round != 0 - broadcastMsg - send post request ", zap.String("host", SaiP2pAddress), zap.Error(err))
		return err
	}

	defer resp.Body.Close()

	// if resp.StatusCode != 200 {
	// 	s.GlobalService.Logger.Sugar().Errorf("process - round != 0 - broadcastMsg - send post request to host : %s wrong status code : %d", SaiP2pAddress, resp.StatusCode)
	// 	return fmt.Errorf("Wrong status code : %d", resp.StatusCode)
	// }

	//s.GlobalService.Logger.Sugar().Debugf("broadcasting - success, message : %+v", msg) // DEBUG                                                       //DEBUG
	return nil
}

// form and save new block
func (s *InternalService) formAndSaveNewBlock(previousBlock *models.BlockConsensusMessage, SaiBTCaddress, storageToken string, txMsgs []*models.TransactionMessage, rnd int64) (*models.BlockConsensusMessage, error) {
	newBlock := &models.BlockConsensusMessage{
		Block: &models.Block{
			Type:              models.BlockConsensusMsgType,
			BaseRND:           rnd,
			Number:            previousBlock.Block.Number,
			PreviousBlockHash: previousBlock.BlockHash,
			SenderAddress:     s.BTCkeys.Address,
			Messages:          make([]*models.TransactionMessage, 0),
		},
	}

	if newBlock.Block.Number == 1 {
		newBlock.Block.PreviousBlockHash = ""
	}

	newBlock.BlockHash = newBlock.Block.BlockHash

	for _, tx := range txMsgs {
		err, _ := s.Storage.Update(MessagesPoolCol, bson.M{"executed_hash": tx.ExecutedHash}, bson.M{"block_hash": newBlock.BlockHash, "block_number": newBlock.Block.Number}, storageToken)
		if err != nil {
			s.GlobalService.Logger.Error("process - round == 7 - form and save new block - update tx blockhash", zap.Error(err))
			return nil, err
		}

		tx.Votes = [7]uint64{}
		tx.BlockHash = "todo: circular reference"
		tx.BlockNumber = newBlock.Block.Number
		newBlock.Block.Messages = append(newBlock.Block.Messages, tx)
	}

	err := newBlock.Block.HashAndSign(SaiBTCaddress, s.BTCkeys.Private)
	if err != nil {
		s.GlobalService.Logger.Error("process - round == 7 - form and save new block - hash and sign block", zap.Error(err))
		return nil, err
	}

	newBlock.BlockHash = newBlock.Block.BlockHash

	newBlock.Votes = +1
	newBlock.Signatures = append(newBlock.Signatures, newBlock.Block.SenderSignature)

	if len(txMsgs) == 0 {
		s.GlobalService.Logger.Debug("process - formAndSaveNewBlock - clear consensus")
		rErr, _ := s.Storage.Remove(ConsensusPoolCol, bson.M{}, storageToken)

		if rErr != nil {
			s.GlobalService.Logger.Error("process - round == 7 - form and save new block - clear Consensus Pool", zap.Error(rErr))
			return nil, rErr
		}

		err := s.updateTxMsgZeroVotes(storageToken)

		if err != nil {
			s.GlobalService.Logger.Error("process - round == 7 - form and save new block - update tx msgs zero votes", zap.Error(err))
			return nil, err
		}
		s.GlobalService.Logger.Error("process - round == 7 - form and save new block - no tx messages found")
		return nil, errNoTxToFromBlock
	}
	s.GlobalService.Logger.Debug("process - formAndSaveNewBlock - formed block before count votes", zap.Int("block_number", newBlock.Block.Number), zap.Int("votes", newBlock.Votes), zap.String("hash", newBlock.BlockHash))

	requiredVotes := math.Ceil(float64(len(s.Validators)) * 7 / 10)

	// check if we have already such block candidate
	blockCandidate, err := s.getBlockCandidate(newBlock.BlockHash, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - round == 7 - form and save new block - get block candidate", zap.Error(err))
		return nil, err
	}

	// if there is no such blockCandidate, save block to BlockCandidate collection
	if blockCandidate == nil {
		if float64(newBlock.Votes) >= requiredVotes {
			err, _ := s.Storage.Put(blockchainCol, newBlock, storageToken)
			if err != nil {
				s.GlobalService.Logger.Error("process - formAndSaveNewBlock - blockCandidate was not found - add block to blockchain", zap.Error(err))
				return nil, err
			}
			s.GlobalService.Logger.Debug("process - formAndSaveNewBlock - found blockCandidate put to blockchain", zap.Int("block_number", newBlock.Block.Number), zap.Int("votes", newBlock.Votes), zap.String("hash", newBlock.BlockHash))
		} else {
			s.GlobalService.Logger.Debug("process - formSaveNewBlock  - blockCandidate not found - put to candidates", zap.Int("block_number", newBlock.Block.Number), zap.String("hash", newBlock.BlockHash), zap.Strings("signatures", newBlock.Signatures)) // DEBUG
			err, _ := s.Storage.Put(BlockCandidatesCol, newBlock, storageToken)
			if err != nil {
				s.GlobalService.Logger.Error("process - round == 7 - form and save new block - put to BlockCandidate collection", zap.Error(err))
				return nil, err
			}
			s.GlobalService.Logger.Debug("process - formAndSaveNewBlock - put to candidates", zap.Int("block_number", newBlock.Block.Number), zap.Int("votes", newBlock.Votes), zap.String("hash", newBlock.BlockHash), zap.Strings("signatures", newBlock.Signatures))
		}

	} else { // else, add vote and signature and save to blockchain
		s.GlobalService.Logger.Debug("process - formAndSaveNewBlock - found candidates with hash", zap.Int("block_number", blockCandidate.Block.Number), zap.Int("votes", blockCandidate.Votes), zap.String("hash", blockCandidate.BlockHash), zap.Strings("signatures", blockCandidate.Signatures))
		blockCandidate.Votes = newBlock.Votes + blockCandidate.Votes
		blockCandidate.Signatures = append(blockCandidate.Signatures, newBlock.Signatures...)
		s.GlobalService.Logger.Debug("process - formAndSaveNewBlock - blockCandidate after voting", zap.Int("block_number", blockCandidate.Block.Number), zap.Int("votes", blockCandidate.Votes), zap.String("hash", blockCandidate.BlockHash))

		if float64(blockCandidate.Votes) >= requiredVotes {
			err, _ := s.Storage.Put(blockchainCol, blockCandidate, storageToken)
			if err != nil {
				s.GlobalService.Logger.Error("process - formAndSaveNewBlock - found blockCandidate - insert block to BlockCandidates collection", zap.Error(err))
				return nil, err
			}
			s.GlobalService.Logger.Debug("process - formAndSaveNewBlock - found blockCandidate put to blockchain", zap.Int("block_number", blockCandidate.Block.Number), zap.Int("votes", blockCandidate.Votes), zap.String("hash", blockCandidate.BlockHash))

			newBlock = blockCandidate
		} else {
			filter := bson.M{"block_hash": blockCandidate.BlockHash}
			update := bson.M{"votes": blockCandidate.Votes, "voted_signatures": blockCandidate.Signatures}

			s.GlobalService.Logger.Debug("process - formSaveNewBlock  - blockCandidate  found - - not enough votes -  put to candidates", zap.Int("block_number", newBlock.Block.Number), zap.String("hash", newBlock.BlockHash)) // DEBUG
			err, _ := s.Storage.Update(BlockCandidatesCol, filter, update, storageToken)
			if err != nil {
				s.GlobalService.Logger.Error("process - round == 7 - form and save new block - put to BlockCandidate collection", zap.Error(err))
				return nil, err
			}

			newBlock = blockCandidate
		}

	}

	err = s.updateTxMsgZeroVotes(storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - round == 7  - form and save new block - clear messages", zap.Error(err))
		return nil, err
	}

	s.GlobalService.Logger.Sugar().Debugf("formed new block to save: %+v\n", newBlock) //DEBUG

	return newBlock, nil

}

// update votes to zero for transaction message
func (s *InternalService) updateTxMsgZeroVotes(storageToken string) error {
	criteria := bson.M{"votes.0": bson.M{"$gte": 1}, "block_hash": ""}
	update := bson.M{"votes": bson.A{0, 0, 0, 0, 0, 0, 0}}

	//_, result := s.Storage.Get("MessagesPool", criteria, bson.M{}, storageToken)
	//	s.GlobalService.Logger.Sugar().Debugf("BEFORE UPDATING ON CLEAR RESULT : %s", string(result))

	err, _ := s.Storage.Update("MessagesPool", criteria, update, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("handlers - process - round != 0 - get messages for specified round", zap.Error(err))
		return err
	}

	//criteria2 := bson.M{"block_hash": ""}
	//_, result2 := s.Storage.Get("MessagesPool", criteria2, bson.M{}, storageToken)
	//	s.GlobalService.Logger.Sugar().Debugf("AFTER UPDATING ON CLEAR RESULT : %s", string(result2))

	return nil
}

// update votes for transaction message
func (s *InternalService) updateTxMsgVotes(hash, storageToken string, round int) error {
	// DEBUG votes updating
	// criteria := bson.M{"message_hash": hash}
	// _, result := s.Storage.Get("MessagesPool", criteria, bson.M{}, storageToken)
	// s.GlobalService.Logger.Sugar().Debugf("BEFORE UPDATING ON ROUND : %d, RESULT : %s", round, string(result))
	/////

	criteria := bson.M{"executed_hash": hash}
	update := bson.M{"$inc": bson.M{"votes." + strconv.Itoa(round): 1}}
	err, _ := s.Storage.Upsert("MessagesPool", criteria, update, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("handlers - process - round != 0 - get messages for specified round", zap.Error(err))
		return err
	}

	// DEBUG votes updating
	// criteria = bson.M{"message_hash": hash}
	// _, result = s.Storage.Get("MessagesPool", criteria, bson.M{}, storageToken)
	// s.GlobalService.Logger.Sugar().Debugf("AFTER UPDATING ON ROUND : %d, RESULT : %s", round, string(result))
	/////
	return nil
}

// get messages with certain number of votes
func (s *InternalService) getTxMsgsWithCertainNumberOfVotes(storageToken string, round int) ([]*models.TransactionMessage, error) {
	requiredVotes := math.Ceil(float64(len(s.Validators)) * float64(round) * 10 / 100)
	filterGte := bson.M{"votes." + strconv.Itoa(round): bson.M{"$gte": requiredVotes}}
	filteredTx := make([]*models.TransactionMessage, 0)
	txMsgs := make([]*models.TransactionMessage, 0)

	err, result := s.Storage.Get(MessagesPoolCol, filterGte, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("handlers - process - round != 0 - get tx messages with specified votes count", zap.Float64("votes count", requiredVotes), zap.Error(err))
		return nil, err
	}

	if len(result) == 2 {
		s.GlobalService.Logger.Error("process - get tx msgs with certain number of votes - emtpy result", zap.String("required votes", strconv.Itoa(int(requiredVotes))))
		return filteredTx, nil
	}

	data, err := utils.ExtractResult(result)
	if err != nil {
		Service.GlobalService.Logger.Error("process - getZeroVotedTransactions - extract data from response", zap.String("result", string(result)), zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(data, &txMsgs)
	if err != nil {
		s.GlobalService.Logger.Error("process - round != 0 - unmarshal result of consensus messages for specified round", zap.Error(err))
		return nil, err
	}

	for _, tx := range txMsgs {
		if tx.BlockHash == "" {
			filteredTx = append(filteredTx, tx)
		}
	}

	sort.Slice(filteredTx, func(i, j int) bool {
		if filteredTx[i].Tx.Nonce == filteredTx[j].Tx.Nonce {
			return filteredTx[i].Tx.MessageHash < filteredTx[j].Tx.MessageHash
		}
		return filteredTx[i].Tx.Nonce < filteredTx[j].Tx.Nonce
	})

	return filteredTx, nil
}

func (s *InternalService) GetBTCkeys(fileStr, SaiBTCaddress string) (*models.BtcKeys, error) {
	file, err := os.OpenFile(fileStr, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		s.GlobalService.Logger.Error("processing - open key btc file", zap.Error(err))
		return nil, err
	}

	data, err := ioutil.ReadFile(fileStr)
	if err != nil {
		s.GlobalService.Logger.Error("processing - read key btc file", zap.Error(err))
		return nil, err
	}
	btcKeys := models.BtcKeys{}

	err = json.Unmarshal(data, &btcKeys)
	if err != nil {
		s.GlobalService.Logger.Error("get btc keys - error unmarshal from file", zap.Error(err))
		btcKeys, body, err := models.GetBtcKeys(SaiBTCaddress)
		if err != nil {
			s.GlobalService.Logger.Error("processing - get btc keys", zap.Error(err))
			return nil, err
		}
		_, err = file.Write(body)
		if err != nil {
			s.GlobalService.Logger.Error("processing - write btc keys to file", zap.Error(err))
			return nil, err
		}
		return btcKeys, nil
	} else {
		err = btcKeys.Validate()
		if err != nil {
			btcKeys, body, err := models.GetBtcKeys(SaiBTCaddress)
			if err != nil {
				s.GlobalService.Logger.Fatal("processing - get btc keys", zap.Error(err))
				return nil, err
			}
			_, err = file.Write(body)
			if err != nil {
				s.GlobalService.Logger.Fatal("processing - write btc keys to file", zap.Error(err))
				return nil, err
			}
			return btcKeys, nil
		}
		return &btcKeys, nil
	}
}

func (s *InternalService) removeCandidates(storageToken string) error {
	err, result := s.Storage.Get(BlockCandidatesCol, bson.M{}, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - remove candidates - get all candidates", zap.Error(err))
		return err
	}

	if len(result) == 2 {
		return nil
	}

	data, err := utils.ExtractResult(result)
	if err != nil {
		Service.GlobalService.Logger.Error("process - remove candidates - extract data from response", zap.Error(err))
		return err
	}

	blockCandidates := make([]models.BlockConsensusMessage, 0)
	err = json.Unmarshal(data, &blockCandidates)
	if err != nil {
		s.GlobalService.Logger.Error("process - remove candidates - unmarshal", zap.Error(err))
		return err
	}

	if len(blockCandidates) == 0 {
		return nil
	}

	for _, blockCandidate := range blockCandidates {
		if blockCandidate.Block.Messages == nil {
			continue
		}
		for _, tx := range blockCandidate.Block.Messages {
			err, _ := s.Storage.Update(MessagesPoolCol, bson.M{"message_hash": tx.MessageHash}, bson.M{"block_number": 0, "block_hash": ""}, storageToken)
			if err != nil {
				s.GlobalService.Logger.Error("process - remove candidates - update tx", zap.Error(err))
				continue
			}
			err = s.updateTxMsgZeroVotes(storageToken)
			if err != nil {
				s.GlobalService.Logger.Error("process - remove candidates - update tx msgs zero votes", zap.Error(err))
				continue
			}
		}
	}

	err, _ = s.Storage.Remove(BlockCandidatesCol, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - remove candidates - remove all candidates", zap.Error(err))
		return err
	}
	return nil
}

// get block candidate by block number, add vote if exists
func (s *InternalService) getBlockCandidateByNumber(blockNumber int, storageToken string) ([]models.BlockConsensusMessage, error) {
	err, result := s.Storage.Get(BlockCandidatesCol, bson.M{"block.number": blockNumber}, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("handleBlockConsensusMsg - blockHash != msgBlockHash - get block candidate by msg block number", zap.Error(err))
		return nil, err
	}

	if len(result) == 2 {
		return nil, nil
	}

	data, err := utils.ExtractResult(result)
	if err != nil {
		Service.GlobalService.Logger.Error("process - get last block from blockchain - extract data from response", zap.Error(err))
		return nil, err
	}

	blockCandidates := make([]models.BlockConsensusMessage, 0)
	err = json.Unmarshal(data, &blockCandidates)
	if err != nil {
		s.GlobalService.Logger.Error("handleBlockConsensusMsg - blockCandidateHash = msgBlockHash - unmarshal", zap.Error(err))
		return nil, err
	}

	return blockCandidates, nil

}
