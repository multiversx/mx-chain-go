package detector

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// HeaderSlashingDetector - checks for slashable events for headers
type HeaderSlashingDetector struct {
	cache            *roundProposerDataCache
	nodesCoordinator sharding.NodesCoordinator
}

// NewHeaderSlashingDetector - creates a new header slashing detector for multiple propose
func NewHeaderSlashingDetector(nodesCoordinator sharding.NodesCoordinator) (slash.SlashingDetector, error) {
	if check.IfNil(nodesCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	//TODO: Use a number from config file
	cache := newRoundProposerDataCache(10)
	return &HeaderSlashingDetector{
		cache:            cache,
		nodesCoordinator: nodesCoordinator,
	}, nil
}

// VerifyData - checks if an intercepted data represents a slashable event
func (hsd *HeaderSlashingDetector) VerifyData(data process.InterceptedData) (slash.SlashingProofHandler, error) {
	currentHeader, ok := data.(*interceptedBlocks.InterceptedHeader)
	if !ok {
		return nil, process.ErrCannotCastInterceptedDataToHeader
	}

	proposer, err := hsd.getProposer(currentHeader.HeaderHandler())
	if err != nil {
		return nil, err
	}

	currRound := currentHeader.HeaderHandler().GetRound()
	message, data2 := hsd.getSlashingResult(currentHeader, currRound, proposer)

	hsd.cache.add(currRound, proposer, currentHeader)

	if message == slash.MultipleProposal {
		return slash.NewMultipleProposalProof("0", message, data2)
	}
	// check another header with the same round and proposer exists, but a different hash
	// if yes a slashingDetectorResult is returned with a message and the two headers

	return slash.NewSlashingProof("0", slash.None), nil
}

// ValidateProof - validates the given proof
func (hsd *HeaderSlashingDetector) ValidateProof(proof slash.SlashingProofHandler) error {
	switch proof.GetType() {
	case slash.None:
	case slash.MultipleProposal:
	default:

	}

	return nil
}

func (hsd *HeaderSlashingDetector) getProposer(header data.HeaderHandler) ([]byte, error) {
	validators, err := hsd.nodesCoordinator.ComputeConsensusGroup(
		header.GetRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		header.GetEpoch())

	if err != nil {
		return nil, err
	}
	return validators[0].PubKey(), nil
}

func (hsd *HeaderSlashingDetector) getSlashingResult(
	currHeader process.InterceptedData,
	currRound uint64,
	proposerPubKey []byte,
) (slash.SlashingType, []process.InterceptedData) {
	headers := make([]process.InterceptedData, 0)
	message := slash.None
	proposedHeaders := hsd.cache.proposedData(currRound, proposerPubKey)

	if len(proposedHeaders) >= 1 {
		headers = hsd.getProposedHeadersWithDifferentHash(currHeader.Hash(), proposedHeaders)
		if len(headers) >= 1 {
			message = slash.MultipleProposal
			headers = append(headers, currHeader)
		}
	}

	return message, headers
}

func (hsd *HeaderSlashingDetector) getProposedHeadersWithDifferentHash(currHash []byte, otherHeaders dataList) []process.InterceptedData {
	ret := make([]process.InterceptedData, 0)

	for _, currHeader := range otherHeaders {
		if !bytes.Equal(currHash, currHeader.Hash()) {
			ret = append(ret, currHeader)
		}
	}

	return ret
}
