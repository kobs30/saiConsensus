package internal

import (
	"github.com/iamthe1whoknocks/bft/models"
	"go.uber.org/zap"
)

// sync node by fixing sleep value
func (s *InternalService) syncSleep(msg *models.ConsensusMessage) int {
	s.GlobalService.Logger.Debug("process - sync sleep")

	s.SyncConsensus.Mu.Lock()
	defer s.SyncConsensus.Mu.Unlock()

	// Debug
	for k, v := range s.SyncConsensus.Storage {
		s.GlobalService.Logger.Debug("-----", zap.Any("key", k), zap.Int("count", v))
	}
	// Debug

	//currentKey := string(msg.BlockNumber) + "+" + string(msg.Round)
	currentKey := models.SyncConsensusKey{
		BlockNumber: msg.BlockNumber,
		Round:       msg.Round,
	}

	currentCount := s.SyncConsensus.Storage[currentKey]

	totalKeysLen := len(s.SyncConsensus.Storage)
	if totalKeysLen == 0 {
		return 1
	}
	s.GlobalService.Logger.Debug("process - sync sleep - map", zap.Int("block_number", msg.BlockNumber), zap.Int("round", msg.Round))

	countForLagging := 0
	for k := range s.SyncConsensus.Storage {
		if k.BlockNumber == msg.BlockNumber && k.Round-msg.Round > 1 {
			countForLagging++
		}
	}
	if countForLagging == totalKeysLen { // если мы отстаем более чем на 1 раунд от любой ноды, независимо от веса -> round = round + 2
		s.GlobalService.Logger.Debug("process -sync - lagging at all nodes")
		return 2
	}

	for k, v := range s.SyncConsensus.Storage {
		if k.BlockNumber == msg.BlockNumber && v >= currentCount && msg.Round < k.Round { // если мы отстаем от любого раунда, у которого вес больше или равен нашему -> round = round + 2
			s.GlobalService.Logger.Debug("process -sync - lagging at special node")
			return 2
		} else if k.BlockNumber == msg.BlockNumber && msg.Round > k.Round && v > currentCount { // если мы торопимся и имеем меньший раунд с большим весом чем у нас, т.е. 3й раунд с весом 3, то round = round + 0
			s.GlobalService.Logger.Debug("process - sync - we are hurrying")
			return 0
		}
	}

	return 1

}

func (s *InternalService) clearSyncMap() {
	s.SyncConsensus.Storage = map[models.SyncConsensusKey]int{}
}
