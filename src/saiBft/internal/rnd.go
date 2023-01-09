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
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// process rnd before creating new block
func (s *InternalService) rndProcessing(saiBTCAddress, saiP2pAddress, storageToken string, blockNumber int) (*models.RND, error) {
	s.GlobalService.Logger.Debug("process - rnd processing started")
	// select transaction message candidates to block
	criteria := bson.M{"block_number": 0, "block_hash": ""}
	err, result := s.Storage.Get(MessagesPoolCol, criteria, bson.M{}, storageToken)
	if err != nil {
		return nil, err
	}
	txMsgs := make([]*models.TransactionMessage, 0)
	txMsgsHashes := make([]string, 0)
	if len(result) != 2 {
		data, err := utils.ExtractResult(result)
		if err != nil {
			Service.GlobalService.Logger.Error("process - rnd processing - extract data from response", zap.Error(err))
			return nil, err
		}
		err = json.Unmarshal(data, &txMsgs)
		if err != nil {
			s.GlobalService.Logger.Error("handlers - process - unmarshal result of last block from blockchain collection", zap.Error(err))
			return nil, err
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
				return nil, err
			}
		}
	}

	rndRound := 0

	s.GlobalService.Logger.Debug("process - rnd processing", zap.Any("tx messages candidates", txMsgs), zap.Int("round", rndRound))

	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd := source.Int63n(100)
	s.GlobalService.Logger.Debug("rnd generated", zap.Int64("rnd", rnd))

	rndMsg := &models.RND{
		Votes: 1,
		Message: &models.RNDMessage{
			Type:          models.RNDMessageType,
			SenderAddress: s.BTCkeys.Address,
			BlockNumber:   blockNumber,
			Round:         rndRound + 1,
			Rnd:           rnd,
			TxMsgHashes:   txMsgsHashes,
		},
	}

	hash, err := rndMsg.Message.GetHash()
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - get hash", zap.Error(err))
		return nil, err
	}
	rndMsg.Message.Hash = hash

	resp, err := models.SignMessage(rndMsg.Message, saiBTCAddress, s.BTCkeys.Private)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - sign message", zap.Error(err))
		return nil, err
	}

	rndMsg.Message.SenderSignature = resp.Signature

	err, _ = s.Storage.Put(RndMessagesPoolCol, rndMsg, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - put to db", zap.Error(err))
		return nil, err
	}
	s.GlobalService.Logger.Debug("process - rnd - put initial rnd msg", zap.Int("rnd", int(rndMsg.Message.Rnd)))

	err = s.broadcastMsg(rndMsg.Message, saiP2pAddress)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - broadcast msg", zap.Error(err))
		return nil, err
	}

	rndMsg, err = s.getValidatedRnd(storageToken, blockNumber, rndRound+1)
	if err != nil {
		goto getRndForSpecifiedRoundAndBlock
	}
	return rndMsg, nil

getRndForSpecifiedRoundAndBlock:
	rndRound++

	s.GlobalService.Logger.Debug("process - rnd processing -  new round", zap.Int("round", rndRound), zap.Int("rnd", int(rnd)))

	// get rnd messages for the round and for block
	err, result = s.Storage.Get(RndMessagesPoolCol, bson.M{"message.block_number": blockNumber, "message.round": rndRound, "message.sender_address": bson.M{"$ne": s.BTCkeys.Address}}, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("processing - rnd processing - get rnd for specified round/block", zap.Error(err))
		return nil, err
	}
	//s.GlobalService.Logger.Debug("process - rnd processing - got msg for block and round", zap.Int("block", blockNumber), zap.Int("round", rndRound), zap.String("rnd msgs", string(result)))

	//if specified messages was found
	if len(result) != 2 {
		rndMsgs := make([]*models.RND, 0)
		data, err := utils.ExtractResult(result)
		if err != nil {
			Service.GlobalService.Logger.Error("processing - rnd processing - extract data from response", zap.Error(err))
			return nil, err
		}
		err = json.Unmarshal(data, &rndMsgs)
		if err != nil {
			s.GlobalService.Logger.Error("processing - rnd processing - unmarshal", zap.Error(err))
			return nil, err
		}

		filteredRndMsgs := make([]*models.RND, 0)

		// filter messages which is not from validator list
		for _, msg := range rndMsgs {
			for _, validator := range s.Validators {
				if validator == msg.Message.SenderAddress {
					filteredRndMsgs = append(filteredRndMsgs, msg)
				}
			}
		}

		s.GlobalService.Logger.Debug("process - rnd processing - rnd msgs after filtration", zap.Any("filtered msgs", filteredRndMsgs))

		var newRndMsg *models.RND
		var _rnd = rnd

		for _, msg := range filteredRndMsgs {
			newRndMsg = msg
			if msg.Message.Rnd == _rnd {
				s.GlobalService.Logger.Debug("process - rnd - found rnd msg with the same rnd", zap.Int64("rnd", msg.Message.Rnd), zap.Int("round", rndRound))
				criteria := bson.M{"message.hash": msg.Message.Hash}
				update := bson.M{"$inc": bson.M{"votes": 1}}
				err, _ := s.Storage.Upsert(RndMessagesPoolCol, criteria, update, storageToken)
				if err != nil {
					s.GlobalService.Logger.Error("handlers - process - round != 0 - get messages for specified round", zap.Error(err))
					return nil, err
				}
			} else {
				_rnd += msg.Message.Rnd
			}
		}

		if _rnd > 0 {
			rnd = _rnd
			newRndMsg.Votes = 1
			newRndMsg.Message.Rnd = rnd
			newRndMsg.Message.Round = rndRound + 1

			hash, err := newRndMsg.Message.GetHash()
			if err != nil {
				s.GlobalService.Logger.Error("process - rnd processing - get hash", zap.Error(err))
				return nil, err
			}
			newRndMsg.Message.Hash = hash

			resp, err := models.SignMessage(newRndMsg.Message, saiBTCAddress, s.BTCkeys.Private)
			if err != nil {
				s.GlobalService.Logger.Error("process - rnd processing - sign message", zap.Error(err))
				return nil, err
			}

			newRndMsg.Message.SenderSignature = resp.Signature

			err, _ = s.Storage.Put(RndMessagesPoolCol, newRndMsg, storageToken)
			if err != nil {
				s.GlobalService.Logger.Error("process - rnd processing - put to db", zap.Error(err))
				return nil, err
			}

			err = s.broadcastMsg(newRndMsg.Message, saiP2pAddress)
			if err != nil {
				s.GlobalService.Logger.Error("processing - rnd processing - broadcast msg", zap.Error(err))
				return nil, err
			}
		}
	}

	rndMsg, err = s.getValidatedRnd(storageToken, blockNumber, rndRound)
	if err != nil {
		goto getRndForSpecifiedRoundAndBlock
	}
	return rndMsg, nil
}

// get message with the most votes
func (s *InternalService) getRndMsgWithMostVotes(storageToken string, blockNumber, rndRound int) (*models.RND, error) {
	opts := options.Find().SetSort(bson.M{"votes": -1}).SetLimit(1)
	err, rndResult := s.Storage.Get(RndMessagesPoolCol, bson.M{"message.block_number": blockNumber, "message.round": rndRound}, opts, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("processing - rnd processing - get rnd with max votes", zap.Error(err))
		return nil, err
	}
	if len(rndResult) == 2 {
		return nil, errNoRndMsgsFound
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
	RNDMsg := RNDMsgs[0]
	return RNDMsg, nil
}

// get validated rnd to put in block
func (s *InternalService) getValidatedRnd(storageToken string, blockNumber, round int) (*models.RND, error) {
	time.Sleep(time.Duration(s.GlobalService.Configuration["sleep"].(int)) * time.Second)
	rndMsg, err := s.getRndMsgWithMostVotes(storageToken, blockNumber, round)
	if err != nil {
		s.GlobalService.Logger.Error("processing - rnd processing - get validated rnd - get rnd with most votes", zap.Error(err))
		return nil, err
	}

	requiredVotes := math.Ceil(float64(len(s.Validators)) * 7 / 10)
	if float64(rndMsg.Votes) >= requiredVotes {
		s.GlobalService.Logger.Debug("processing - rnd processing - get validated rnd -  found rnd message to add to block", zap.Float64("required votes", requiredVotes), zap.Any("rnd message", rndMsg))
		return rndMsg, nil
	} else {
		s.GlobalService.Logger.Debug("processing - rnd processing - rnd message to add to block was not found", zap.Float64("required votes", requiredVotes), zap.Any("rnd message", rndMsg))
		return nil, errNotEnoughVotes
	}
}

func (s *InternalService) getRndMsgByRnd(storageToken string, blockNumber, rnd int) (*models.RND, error) {
	err, rndResult := s.Storage.Get(RndMessagesPoolCol, bson.M{"message.block_number": blockNumber, "message.rnd": rnd}, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("processing - rnd processing - get rnd with max votes", zap.Error(err))
		return nil, err
	}
	if len(rndResult) == 2 {
		return nil, errNoRndMsgsFound
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
	RNDMsg := RNDMsgs[0]
	return RNDMsg, nil
}
