package internal

import (
	"encoding/json"
	"math"
	"math/rand"
	"time"

	"github.com/iamthe1whoknocks/bft/models"
	"github.com/iamthe1whoknocks/bft/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// process rnd before creating new block
func (s *InternalService) rndProcessing(saiBTCAddress, saiP2pAddress, storageToken string, blockNumber int) (*models.RNDMessage, error) {
	s.GlobalService.Logger.Debug("process - rnd processing started")
	// select transaction message candidates to block
	criteria := bson.M{"rnd_processed": false}
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
		for _, tx := range txMsgs {
			txMsgsHashes = append(txMsgsHashes, tx.Tx.MessageHash)
			err, _ := s.Storage.Update(MessagesPoolCol, bson.M{"message_hash": tx.MessageHash}, bson.M{"rnd_processed": true}, storageToken)
			if err != nil {
				s.GlobalService.Logger.Error("process - round == 7 - form and save new block - update tx blockhash", zap.Error(err))
				return nil, err
			}
		}
	}

	txMsgHashes := make([]string, 0)

	rndRound := 0

	s.GlobalService.Logger.Debug("process - rnd processing", zap.Any("tx messages candidates", txMsgs), zap.Int("round", rndRound))

	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd := source.Int63n(1000)
	s.GlobalService.Logger.Debug("rnd generated", zap.Int64("rnd", rnd))

	rndMsg := &models.RNDMessage{
		Votes: 1,
		Type:  models.RNDMessageType,
		RND: &models.RND{
			SenderAddress: s.BTCkeys.Address,
			BlockNumber:   blockNumber,
			Round:         rndRound,
			Rnd:           rnd,
			TxMsgHashes:   txMsgHashes,
		},
	}

	hash, err := rndMsg.GetHash()
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - get hash", zap.Error(err))
		return nil, err
	}
	rndMsg.RND.Hash = hash

	resp, err := utils.SignMessage(rndMsg, saiBTCAddress, s.BTCkeys.Private)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - sign message", zap.Error(err))
		return nil, err
	}

	rndMsg.RND.SenderSignature = resp.Signature

	err, _ = s.Storage.Put(RndMessagesPoolCol, rndMsg, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - put to db", zap.Error(err))
		return nil, err
	}

	err = s.broadcastMsg(rndMsg, saiP2pAddress)
	if err != nil {
		s.GlobalService.Logger.Error("process - rnd processing - broadcast msg", zap.Error(err))
		return nil, err
	}

	rndMsg, err = s.getValidatedRnd(storageToken, blockNumber)
	if err != nil {
		goto getRndForSpecifiedRoundAndBlock
	}
	return rndMsg, nil

getRndForSpecifiedRoundAndBlock:
	rndRound++

	s.GlobalService.Logger.Debug("process - rnd processing -  new round", zap.Int("round", rndRound))

	// get rnd messages for the round and for block
	err, result = s.Storage.Get(RndMessagesPoolCol, bson.M{"message.block_number": blockNumber}, bson.M{}, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("processing - rnd processing - get rnd for specified round/block", zap.Error(err))
		return nil, err
	}
	//s.GlobalService.Logger.Debug("process - rnd processing - got msg for block and round", zap.Int("block", blockNumber), zap.Int("round", rndRound), zap.String("rnd msgs", string(result)))

	//if specified messages was found
	if len(result) != 2 {
		rndMsgs := make([]*models.RNDMessage, 0)
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

		filteredRndMsgs := make([]*models.RNDMessage, 0)

		// filter messages which is not from validator list
		for _, msg := range rndMsgs {
			if msg.RND.SenderAddress == s.BTCkeys.Address { //do not vote for own message
				continue
			}
			for _, validator := range s.Validators {
				if validator == msg.RND.SenderAddress {
					filteredRndMsgs = append(filteredRndMsgs, msg)
				}
			}
		}

		//s.GlobalService.Logger.Debug("process - rnd processing - rnd msgs after filtration", zap.Any("filtered msgs", filteredRndMsgs))

		for _, msg := range filteredRndMsgs {
			s.GlobalService.Logger.Sugar().Debugf("ROUND : %d", rndRound)
			s.GlobalService.Logger.Sugar().Debugf("HANDLING MSG : %+v", msg.RND)
			if msg.RND.Rnd == rnd {
				msg.Votes++
				criteria := bson.M{"message.hash": msg.RND.Hash}
				update := bson.M{"$inc": bson.M{"votes": 1}}
				err, _ := s.Storage.Upsert(RndMessagesPoolCol, criteria, update, storageToken)
				if err != nil {
					s.GlobalService.Logger.Error("handlers - process - round != 0 - get messages for specified round", zap.Error(err))
					return nil, err
				}
			} else {
				s.GlobalService.Logger.Sugar().Debugf("OLD MSG : %+v", msg.RND)
				msg.RND.Rnd = msg.RND.Rnd + rnd

				hash, err := msg.GetHash()
				if err != nil {
					s.GlobalService.Logger.Error("process - rnd processing - get hash", zap.Error(err))
					return nil, err
				}
				msg.RND.Hash = hash

				resp, err := utils.SignMessage(msg, saiBTCAddress, s.BTCkeys.Private)
				if err != nil {
					s.GlobalService.Logger.Error("process - rnd processing - sign message", zap.Error(err))
					return nil, err
				}

				msg.RND.SenderSignature = resp.Signature
				msg.RND.Round = rndRound
				s.GlobalService.Logger.Sugar().Debugf("NEW MSG : %+v", msg.RND)

				err, _ = s.Storage.Put(RndMessagesPoolCol, msg, storageToken)
				if err != nil {
					s.GlobalService.Logger.Error("process - rnd processing - put to db", zap.Error(err))
					return nil, err
				}
			}
			err = s.broadcastMsg(msg, saiP2pAddress)
			if err != nil {
				s.GlobalService.Logger.Error("processing - rnd processing - broadcast msg", zap.Error(err))
				return nil, err
			}
		}

		rndMsg, err := s.getValidatedRnd(storageToken, blockNumber)
		if err != nil {
			goto getRndForSpecifiedRoundAndBlock
		}
		return rndMsg, nil

	}
	rndMsg, err = s.getValidatedRnd(storageToken, blockNumber)
	if err != nil {
		goto getRndForSpecifiedRoundAndBlock
	}
	return rndMsg, nil

}

// get message with the most votes
func (s *InternalService) getRndMsgWithMostVotes(storageToken string, blockNumber int) (*models.RNDMessage, error) {
	opts := options.Find().SetSort(bson.M{"votes": -1}).SetLimit(1)
	err, rndResult := s.Storage.Get(RndMessagesPoolCol, bson.M{"message.block_number": blockNumber}, opts, storageToken)
	if err != nil {
		s.GlobalService.Logger.Error("processing - rnd processing - get rnd with max votes", zap.Error(err))
		return nil, err
	}
	if len(rndResult) == 2 {
		return nil, errNoRndMsgsFound
	}

	RNDMsgs := make([]*models.RNDMessage, 0)
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
func (s *InternalService) getValidatedRnd(storageToken string, blockNumber int) (*models.RNDMessage, error) {
	time.Sleep(time.Duration(s.GlobalService.Configuration["sleep"].(int)) * time.Second)
	rndMsg, err := s.getRndMsgWithMostVotes(storageToken, blockNumber)
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
