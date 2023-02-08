package internal

import (
	"github.com/iamthe1whoknocks/bft/models"
)

// sync node by fixing sleep value
func (s *InternalService) syncSleep(msg *models.ConsensusMessage) int {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	currentKey := &models.SyncConsensusKey{
		BlockNumber: msg.BlockNumber,
		Round:       msg.Round,
	}

	currentCount := s.SyncConsensusMap[currentKey]

	delete(s.SyncConsensusMap, currentKey)

	totalKeysLen := len(s.SyncConsensusMap)
	countForLagging := 0
	for k := range s.SyncConsensusMap {
		if k.Round-msg.Round > 1 {
			countForLagging++
		}
	}
	if countForLagging == totalKeysLen { // если мы отстаем более чем на 1 раунд от любой ноды, независимо от веса -> round = round + 2
		return 2
	}

	for k, v := range s.SyncConsensusMap {
		if v >= currentCount && msg.Round < k.Round { // если мы отстаем от любого раунда, у которого вес больше или равен нашему -> round = round + 2

			return 2
		} else if msg.Round > k.Round && v > currentCount { // если мы торопимся и имеем меньший раунд с большим весом чем у нас, т.е. 3й раунд с весом 3, то round = round + 0

			return 0
		}
	}

	return 1

}
