package detector

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type HeaderSlashingResult string

const (
	None             HeaderSlashingResult = "none"
	DoubleProposal   HeaderSlashingResult = "header double proposal"
	MultipleProposal HeaderSlashingResult = "header multiple proposal"
)

// HeaderSlashingDetector - checks for slashable events for headers
type HeaderSlashingDetector struct {
	cache            *roundProposerDataCache
	nodesCoordinator sharding.NodesCoordinator
}

// NewHeaderSlashingDetector - creates a new header slashing detector for multiple propose
func NewHeaderSlashingDetector(nodesCoordinator sharding.NodesCoordinator) slash.SlashingDetector {
	if check.IfNil(nodesCoordinator) {
		return nil // TODO: ,process.ErrNilShardCoordinator
	}

	//TODO: Use a number from config file
	cache := newRoundProposerDataCache(10)
	return &HeaderSlashingDetector{
		cache:            cache,
		nodesCoordinator: nodesCoordinator,
	}
}

//TODO: Return error

// VerifyData - checks if an intercepted data represents a slashable event
func (hsd *HeaderSlashingDetector) VerifyData(data process.InterceptedData) slash.SlashingDetectorResultHandler {
	currentHeader, ok := data.(*interceptedBlocks.InterceptedHeader)
	if !ok {
		return nil
	}

	proposer, err := hsd.getProposer(currentHeader.HeaderHandler())
	if err != nil {
		return nil
	}

	currRound := currentHeader.HeaderHandler().GetRound()
	message, data2 := hsd.getSlashingResult(currentHeader, currRound, proposer)

	hsd.cache.addProposerData(currRound, proposer, currentHeader)

	// check another header with the same round and proposer exists, but a different hash
	// if yes a slashingDetectorResult is returned with a message and the two headers
	return slash.NewSlashingDetectorResult(string(message), currentHeader, data2)
}

// GenerateProof - creates the SlashingProofHandler for the DetectorResult to be added to the Tx Data Field
func (hsd *HeaderSlashingDetector) GenerateProof(result slash.SlashingDetectorResultHandler) slash.SlashingProofHandler {
	return slash.NewSlashingProof("level", result.GetType(), result.GetData1(), result.GetData2())
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

func (hsd *HeaderSlashingDetector) getSlashingResult(currHeader process.InterceptedData, currRound uint64, proposerPubKey []byte) (HeaderSlashingResult, process.InterceptedData) {
	data2 := process.InterceptedData(nil)
	message := None
	proposedHeaders := hsd.cache.getProposedHeaders(currRound, proposerPubKey)

	if len(proposedHeaders) == 1 && bytes.Equal(currHeader.Hash(), proposedHeaders[0].Hash()) {
		data2 = proposedHeaders[0]
		message = DoubleProposal
	} else if len(proposedHeaders) >= 2 {
		data2 = hsd.getFirstProposedHeaderWithDifferentHash(currHeader.Hash(), proposedHeaders)
		if data2 != nil {
			message = MultipleProposal
		}
	}

	return message, data2
}

func (hsd *HeaderSlashingDetector) getFirstProposedHeaderWithDifferentHash(currHash []byte, otherHeaders headerList) process.InterceptedData {
	for _, currHeader := range otherHeaders {
		if bytes.Equal(currHash, currHeader.Hash()) {
			return currHeader
		}
	}

	return nil
}
