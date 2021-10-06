package detector

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

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

func computeSlashLevel(data []process.InterceptedData) slash.SlashingLevel {
	ret := slash.Level0
	// TODO: Maybe a linear interpolation to deduce severity?
	if len(data) == 2 {
		ret = slash.Level1
	} else if len(data) >= 3 {
		ret = slash.Level2
	}

	return ret
}

func checkSlashLevel(headers []*interceptedBlocks.InterceptedHeader, level slash.SlashingLevel) error {
	if level < slash.Level1 || level > slash.Level2 {
		return process.ErrInvalidSlashLevel
	}
	if len(headers) < MinSlashableNoOfHeaders {
		return process.ErrNotEnoughHeadersProvided
	}
	if len(headers) == MinSlashableNoOfHeaders && level != slash.Level1 {
		return process.ErrSlashLevelDoesNotMatchSlashType
	}
	if len(headers) > MinSlashableNoOfHeaders && level != slash.Level2 {
		return process.ErrSlashLevelDoesNotMatchSlashType
	}

	return nil
}
