package detector

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go/process"
)

// minSlashableNoOfHeaders represents the min number of headers required for a
// proof to be considered slashable
const minSlashableNoOfHeaders = 2

// MaxDeltaToCurrentRound represents the max delta from the current round to any
// other round from an intercepted data in order for a detector to process it and cache it
const MaxDeltaToCurrentRound = 3

type baseSlashingDetector struct {
	roundHandler process.RoundHandler
}

func (bsd *baseSlashingDetector) isRoundRelevant(round uint64) bool {
	currRound := uint64(bsd.roundHandler.Index())
	return absDiff(currRound, round) < MaxDeltaToCurrentRound
}

func absDiff(x, y uint64) uint64 {
	if x < y {
		return y - x
	}
	return x - y
}

func computeSlashLevelBasedOnHeadersCount(headers []data.HeaderInfoHandler) coreSlash.ThreatLevel {
	ret := coreSlash.Low

	if len(headers) == minSlashableNoOfHeaders {
		ret = coreSlash.Medium
	} else if len(headers) >= minSlashableNoOfHeaders+1 {
		ret = coreSlash.High
	}

	return ret
}

func checkSlashLevelBasedOnHeadersCount(headers []data.HeaderInfoHandler, level coreSlash.ThreatLevel) error {
	if level < coreSlash.Medium || level > coreSlash.High {
		return process.ErrInvalidSlashLevel
	}
	if len(headers) < minSlashableNoOfHeaders {
		return process.ErrNotEnoughHeadersProvided
	}
	if len(headers) == minSlashableNoOfHeaders && level != coreSlash.Medium {
		return process.ErrSlashLevelDoesNotMatchSlashType
	}
	if len(headers) > minSlashableNoOfHeaders && level != coreSlash.High {
		return process.ErrSlashLevelDoesNotMatchSlashType
	}

	return nil
}
