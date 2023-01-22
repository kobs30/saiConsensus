package internal

import (
	"github.com/iamthe1whoknocks/bft/models"
	"go.uber.org/zap"
)

// unput data for testing purposes

// save test tx (for testing purposes)
func (s *InternalService) saveTestTx(saiBtcAddress, storageToken, saiP2PAddress string) {
	testTxMsg := &models.TransactionMessage{
		Votes: [7]uint64{},
		Tx: &models.Tx{
			Type:          models.TransactionMsgType,
			SenderAddress: s.BTCkeys.Address,
			Message:       "test tx message",
			Nonce:         10,
		},
	}

	err := testTxMsg.HashAndSign(saiBtcAddress, s.BTCkeys.Private)
	if err != nil {
		s.GlobalService.Logger.Fatal("processing - hash and sign test tx", zap.Error(err))
	}

	err, _ = s.Storage.Put("MessagesPool", testTxMsg, storageToken)
	if err != nil {
		s.GlobalService.Logger.Fatal("processing - put test tx msg", zap.Error(err))
	}

	bcErr := s.broadcastMsg(testTxMsg.Tx, saiP2PAddress, false)
	if bcErr != nil {
		s.GlobalService.Logger.Fatal("processing - broadcast test tx msg", zap.Error(err))
	}
	s.GlobalService.Logger.Sugar().Debugf("test tx message saved") //DEBUG
}

// save test consensusMsg (for testing purposes)
func (s *InternalService) saveTestConsensusMsg(saiBtcAddress, storageToken, senderAddress string) {
	testConsensusMsg := &models.ConsensusMessage{
		Type:          models.ConsensusMsgType,
		SenderAddress: senderAddress,
		BlockNumber:   3,
		Round:         7,
		Messages:      []string{"0060ee497708e7d9a8428802a6651b93847dca9a0217d05ad67a5a1be7d49223"},
	}

	err := testConsensusMsg.HashAndSign(saiBtcAddress, s.BTCkeys.Private)
	if err != nil {
		s.GlobalService.Logger.Fatal("processing - hash and sign test consensus msg", zap.Error(err))
	}

	err, _ = s.Storage.Put("ConsensusPool", testConsensusMsg, storageToken)
	if err != nil {
		s.GlobalService.Logger.Fatal("processing - put test consensus msg", zap.Error(err))
	}

	s.GlobalService.Logger.Sugar().Debugf("test consensus message saved") //DEBUG

}

// save test tx (for testing purposes)
func (s *InternalService) saveTestTx2(saiBtcAddress, storageToken, saiP2PAddress string) {
	testTxMsg := &models.TransactionMessage{
		Votes: [7]uint64{},
		Tx: &models.Tx{
			Type:          models.TransactionMsgType,
			SenderAddress: s.BTCkeys.Address,
			Message:       "test tx message 2",
			Nonce:         4,
		},
	}

	err := testTxMsg.HashAndSign(saiBtcAddress, s.BTCkeys.Private)
	if err != nil {
		s.GlobalService.Logger.Fatal("processing - hash and sign test tx2", zap.Error(err))
	}

	err, _ = s.Storage.Put("MessagesPool", testTxMsg, storageToken)
	if err != nil {
		s.GlobalService.Logger.Fatal("processing - put test tx msg", zap.Error(err))
	}

	bcErr := s.broadcastMsg(testTxMsg.Tx, saiP2PAddress, false)
	if bcErr != nil {
		s.GlobalService.Logger.Fatal("processing - broadcast test tx msg", zap.Error(err))
	}
	s.GlobalService.Logger.Sugar().Debugf("test tx message saved") //DEBUG
}
