package internal

import (
	"github.com/iamthe1whoknocks/bft/models"
	"go.uber.org/zap"
)

// sync node by fixing sleep value
func (s *InternalService) syncSleep(msg *models.ConsensusMessage) int {
	s.GlobalService.Logger.Debug("process - sync sleep")

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// Debug
	for k, v := range s.SyncConsensusMap {
		s.GlobalService.Logger.Debug("-----", zap.Int("block_number", k.BlockNumber), zap.Int("round", k.Round), zap.Int("count", v))
	}
	// Debug

	currentKey := &models.SyncConsensusKey{
		BlockNumber: msg.BlockNumber,
		Round:       msg.Round,
	}

	currentCount := s.SyncConsensusMap[currentKey]

	for k := range s.SyncConsensusMap {
		k.Processed = true
	}
	defer s.clearSyncMap()

	totalKeysLen := len(s.SyncConsensusMap)
	if totalKeysLen == 0 {
		return 1
	}
	s.GlobalService.Logger.Debug("process - sync sleep - map", zap.Int("block_number", msg.BlockNumber), zap.Int("round", msg.Round))

	countForLagging := 0
	for k := range s.SyncConsensusMap {
		if k.BlockNumber == msg.BlockNumber && k.Round-msg.Round > 1 {
			countForLagging++
		}
	}
	if countForLagging == totalKeysLen { // если мы отстаем более чем на 1 раунд от любой ноды, независимо от веса -> round = round + 2
		return 2
	}

	for k, v := range s.SyncConsensusMap {
		if k.BlockNumber == msg.BlockNumber && v >= currentCount && msg.Round < k.Round { // если мы отстаем от любого раунда, у которого вес больше или равен нашему -> round = round + 2
			return 2
		} else if k.BlockNumber == msg.BlockNumber && msg.Round > k.Round && v > currentCount { // если мы торопимся и имеем меньший раунд с большим весом чем у нас, т.е. 3й раунд с весом 3, то round = round + 0

			return 0
		}
	}

	return 1

}

func (s *InternalService) clearSyncMap() {
	for k := range s.SyncConsensusMap {
		if k.Processed {
			delete(s.SyncConsensusMap, k)
		}
	}
}
