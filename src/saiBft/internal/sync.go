package internal

import (
	"time"

	"github.com/iamthe1whoknocks/bft/models"
)

// sync node by fixing sleep value
// fromHandler set to true if this func was invoked from handler, set to false if from process
func (s *InternalService) syncSleep(msg *models.ConsensusMessage, fromHandler bool) {
	consensusSyncKey := models.SyncConsensusKey{
		BlockNumber: msg.BlockNumber,
		Round:       msg.Round,
	}

	Service.Mutex.RLock()

	if t, ok := s.SyncConsensusMap[consensusSyncKey]; ok {
		Service.Mutex.RUnlock()
		if t == time.Now().Unix() { // do not fix sleep
			return
		} else if t < time.Now().UnixNano() {
			if !fromHandler {
				delta := time.Now().UnixNano() - t
				s.Sleep = time.Duration(s.Sleep.Nanoseconds() - delta)
				return
			} else {
				return
			}

		}
	} else {
		Service.Mutex.RUnlock()
		Service.Mutex.Lock()
		Service.SyncConsensusMap[consensusSyncKey] = time.Now().Unix()
		Service.Mutex.Unlock()
	}
	return
}
