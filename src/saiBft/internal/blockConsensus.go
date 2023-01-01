package internal

import (
	"github.com/iamthe1whoknocks/bft/models"
	"go.uber.org/zap"
)

// merge msg and block from db (filter unique addressed, signatures to not to vote to the same block few times)
func (s *InternalService) getFilteredBlockConsensus(msg, block *models.BlockConsensusMessage) *models.BlockConsensusMessage {
	addrMap := make(map[string]int)

	for _, addr := range msg.VotedAddresses {
		addrMap[addr]++
	}

	for _, blockAddr := range block.VotedAddresses {
		addrMap[blockAddr]--
	}

	addrToAdd := make([]string, 0)
	for k, v := range addrMap {
		if v == 1 {
			addrToAdd = append(addrToAdd, k)
		}
	}

	if len(addrToAdd) == 0 {
		s.GlobalService.Logger.Debug("chain - handleBlockConsensusMsg - unique addreses not found")
		return nil
	}

	indexes := make([]int, 0)
	for i, msgAddr := range msg.VotedAddresses {
		for _, addr := range addrToAdd {
			if msgAddr == addr {
				indexes = append(indexes, i)
			}
		}
	}

	s.GlobalService.Logger.Debug("chain - handleBlockConsensusMsg - indexes", zap.Ints("indexes", indexes), zap.Strings("unique addresses", addrToAdd), zap.Strings("msg addresses", msg.VotedAddresses))

	uniqueSignatures := make([]string, 0)
	for _, v := range indexes {
		uniqueSignatures = append(uniqueSignatures, msg.Signatures[v])
	}

	block.Votes = block.Votes + len(addrToAdd)
	block.VotedAddresses = append(block.VotedAddresses, addrToAdd...)
	block.Signatures = append(block.Signatures, uniqueSignatures...)
	return block
}
