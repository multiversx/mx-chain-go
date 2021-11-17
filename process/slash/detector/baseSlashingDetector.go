package detector

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
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

func checkAndGetHeader(interceptedData process.InterceptedData) (data.HeaderHandler, error) {
	if check.IfNil(interceptedData) {
		return nil, process.ErrNilInterceptedData
	}

	interceptedHeader, castOk := interceptedData.(process.InterceptedHeader)
	if !castOk {
		return nil, process.ErrCannotCastInterceptedDataToHeader
	}

	header := interceptedHeader.HeaderHandler()
	if check.IfNil(header) {
		return nil, process.ErrNilHeaderHandler
	}

	return header, nil
}

func checkProofType(proof coreSlash.SlashingProofHandler, expectedType coreSlash.SlashingType) error {
	if proof == nil {
		return process.ErrNilProof
	}
	if proof.GetType() != expectedType {
		return process.ErrInvalidSlashType
	}

	return nil
}

func getHeaderHandlers(headersInfo []data.HeaderInfoHandler) []data.HeaderHandler {
	headers := make([]data.HeaderHandler, 0, len(headersInfo))
	for _, headerInfo := range headersInfo {
		headers = append(headers, headerInfo.GetHeaderHandler())
	}

	return headers
}

func computeSlashLevelBasedOnHeadersCount(headers []data.HeaderHandler) coreSlash.ThreatLevel {
	ret := coreSlash.Zero

	if len(headers) == minSlashableNoOfHeaders {
		ret = coreSlash.Medium
	} else if len(headers) >= minSlashableNoOfHeaders+1 {
		ret = coreSlash.High
	}

	return ret
}

func checkThreatLevelBasedOnHeadersCount(headers []data.HeaderHandler, level coreSlash.ThreatLevel) error {
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
