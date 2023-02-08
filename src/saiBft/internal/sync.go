package internal

import (
	"math"

	"github.com/iamthe1whoknocks/bft/models"
	"go.uber.org/zap"
)

// sync node by fixing sleep value
func (s *InternalService) syncSleep(msg *models.ConsensusMessage, localTime int64) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	s.GlobalService.Logger.Debug("process - sync", zap.Duration("sleep before syncing", s.Sleep))

	if len(s.SyncConsensusSlice) == 0 {
		s.GlobalService.Logger.Debug("process - sync - sync slice is empty")
		return
	}

	//find max
	var maxTime int64
	for _, k := range s.SyncConsensusSlice {
		if maxTime < k.Time {
			maxTime = k.Time
		}
	}

	//find min
	var minTime int64
	for _, k := range s.SyncConsensusSlice {
		if minTime > k.Time && minTime != 0 {
			minTime = k.Time
		} else {
			minTime = k.Time
		}
	}

	res1 := maxTime - localTime

	res2 := localTime - minTime

	if math.Abs(float64(res1)) > math.Abs(float64(res2)) {
		s.Sleep = s.Sleep + s.Sleep/10
	} else if math.Abs(float64(res1)) < math.Abs(float64(res2)) {
		s.Sleep = s.Sleep - s.Sleep/10
	} else {
	}
	s.GlobalService.Logger.Debug("process - sync", zap.Duration("sleep before after", s.Sleep))
	return
}
