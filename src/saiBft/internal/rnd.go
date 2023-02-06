package internal

import (
	"encoding/json"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/iamthe1whoknocks/bft/models"
	"github.com/iamthe1whoknocks/bft/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

const (
	maxRndRound = 7
)

// process rnd before creating new block
func (s *InternalService) rndProcessing(saiBTCAddress, saiP2pAddress, storageToken string, blockNumber int) (int64, error) {
	s.GlobalService.Logger.Debug("process - rnd processing started")
	// select transaction message candidates to block
	criteria := bson.M{"block_number": 0, "block_hash": ""}
	err, result := s.Storage.Get(MessagesPoolCol, criteria, bson.M{}, storageToken)
	if err != nil {
		return 0, err
	}
	txMsgs := make([]*models.TransactionMessage, 0)
	txMsgsHashes := make([]string, 0)
	if len(result) != 2 {
		data, err := utils.ExtractResult(result)
		if err != nil {
			Service.GlobalService.Logger.Error("process - rnd processing - extract data from response", zap.Error(err))
			return 0, err
		}
		err = json.Unmarshal(data, &txMsgs)
		if err != nil {
			s.GlobalService.Logger.Error("handlers - process - unmarshal result of last block from blockchain collection", zap.Error(err))
			return 0, err
		}

		//sort tx messages by nonce, by hash
		sort.Slice(txMsgs, func(i, j int) bool {
			if txMsgs[i].Tx.Nonce == txMsgs[j].Tx.Nonce {
				return txMsgs[i].Tx.MessageHash < txMsgs[j].Tx.MessageHash
			}
			return txMsgs[i].Tx.Nonce < txMsgs[j].Tx.Nonce
		})

		for _, tx := range txMsgs {
			txMsgsHashes = append(txMsgsHashes, tx.Tx.MessageHash)
			err, _ := s.Storage.Update(MessagesPoolCol, bson.M{"message_hash": tx.MessageHash}, bson.M{"block_number": blockNumber}, storageToken)
			if err != nil {
				s.GlobalService.Logger.Error("process - round == 7 - form and save new block - update tx blockhash", zap.Error(err))
				return 0, err
			}
		}
	}

	rndRound := 0

	err, _ = s.Storage.Remove(RndMessagesPoolCol, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - round == 7 - form and save new block - clear rnd error", zap.Error(err))
		return 0, err
	}

	s.GlobalService.Logger.Debug("process - rnd processing", zap.Any("tx messages candidates", txMsgs), zap.Int("round", rndRound))

	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd := source.Int63n(100)
	s.GlobalService.Logger.Debug("rnd generated", zap.Int64("rnd", rnd))
	baseRnd := int64(0)

nextRound:
	s.GlobalService.Logger.Debug("process - rnd processing -  new round", zap.Int("round", rndRound), zap.Int("rnd", int(rnd)))

	rndMsg := &models.RND{
		Votes: 0,
		Message: &models.RNDMessage{
			Type:          models.RNDMessageType,
			SenderAddress: s.BTCkeys.Address,
			BlockNumber:   blockNumber,
			Round:         rndRound + 1,
			Rnd:           rnd,
			TxMsgHashes:   txMsgsHashes,
		},
	}

	err = rndMsg.Message.HashAndSign(saiBTCAddress, s.BTCkeys.Private)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - hash and sign message", zap.Error(err))
		return 0, err
	}

	s.GlobalService.Logger.Debug("process - rnd - put msg to storage", zap.Int("block_number", blockNumber), zap.Int("round", rndRound), zap.Any("rndMsg", rndMsg))

	err, _ = s.Storage.Put(RndMessagesPoolCol, rndMsg, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - put to db", zap.Error(err))
		return 0, err
	}

	requiredVotes := math.Ceil(float64(len(s.Validators)) * 7 / 10)

	err = s.broadcastMsg(rndMsg.Message, saiP2pAddress, false)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - broadcast msg", zap.Error(err))
		return 0, err
	}

	time.Sleep(time.Duration(time.Duration(s.Sleep) * time.Second))

	resultMap, err := s.GetResultRoundMap(blockNumber, rndRound+1)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing -  get result map", zap.Error(err))
		goto nextRound
	}

	s.GlobalService.Logger.Debug("process - rnd - got result map", zap.Any("map", resultMap), zap.Float64("required votes", requiredVotes), zap.Int("round", rndRound))

	for k, v := range resultMap {
		if math.Ceil(float64(v)) >= requiredVotes {
			rnd = k
			baseRnd = k
		}
	}

	if baseRnd == 0 {
		var _rnd int64 = 0
		for k := range resultMap {
			_rnd += k
			s.GlobalService.Logger.Debug("process - rnd - new rnd after sum", zap.Float64("required votes", requiredVotes), zap.Int("round", rndRound), zap.Int("new rnd", int(rnd)))
		}
		rnd = _rnd
	}

	rndRound++

	if !(rndRound >= 7) {
		goto nextRound
	}

	if baseRnd > 0 {
		return baseRnd, nil
	} else {
		// rnd, err := getRndFromDirectMsg()
		// if err != nil {
		// 	s.GlobalService.Logger.Error("process - rnd processing -  get rnd form direct connection", zap.Error(err))
		// 	return 0, err
		// }
	}
	return 0, nil
}

func (s *InternalService) GetResultRoundMap(blockNumber, round int) (map[int64]int, error) {
	s.GlobalService.Logger.Debug("rnd - request for rndMsg from db", zap.Int("block_number", blockNumber), zap.Int("round", round))
	err, rndResult := s.Storage.Get(RndMessagesPoolCol, bson.M{"message.block_number": blockNumber, "message.round": round}, bson.M{}, s.GlobalService.GetConfig(SaiStorageToken, "").String())
	if err != nil {
		s.GlobalService.Logger.Error("processing - rnd processing - get rnd msg for round and block", zap.Error(err))
		return nil, err
	}
	if len(rndResult) == 2 {
		return nil, nil
	}

	RNDMsgs := make([]*models.RND, 0)
	data, err := utils.ExtractResult(rndResult)
	if err != nil {
		s.GlobalService.Logger.Error("processing - rnd processing - extract data from response", zap.Error(err))
		return nil, err
	}
	err = json.Unmarshal(data, &RNDMsgs)
	if err != nil {
		s.GlobalService.Logger.Error("processing - rnd processing - unmarshal", zap.Error(err))
		return nil, err
	}

	resultMap := make(map[int64]int)

	for _, rndMsg := range RNDMsgs {
		resultMap[rndMsg.Message.Rnd]++
	}
	return resultMap, nil
}
