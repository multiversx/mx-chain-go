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
	slashRes, slashType, headers := hsd.getSlashingResult(currentHeader, currRound, proposer)

	hsd.cache.add(currRound, proposer, currentHeader)

	if slashRes == slash.MultipleProposal {
		// TODO: Maybe a linear interpolation to deduce severity?
		return slash.NewMultipleProposalProof(slashType, slashRes, headers)
	}
	// check another header with the same round and proposer exists, but a different hash
	// if yes a slashingDetectorResult is returned with a message and the two headers
	return slash.NewSlashingProof(slash.Level0, slash.None), nil
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
) (slash.SlashingType, slash.SlashingLevel, []process.InterceptedData) {
	headers := make([]process.InterceptedData, 0)
	slashType := slash.None
	slashLevel := slash.Level0
	proposedHeaders := hsd.cache.proposedData(currRound, proposerPubKey)

	if len(proposedHeaders) >= 1 {
		headers = hsd.getProposedHeadersWithDifferentHash(currHeader.Hash(), proposedHeaders)
		if len(headers) >= 1 {
			if len(headers) == 1 {
				slashLevel = slash.Level1
			} else {
				slashLevel = slash.Level2
			}
			slashType = slash.MultipleProposal
			headers = append(headers, currHeader)
		}
	}

	return slashType, slashLevel, headers
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

// ValidateProof - validates the given proof
func (hsd *HeaderSlashingDetector) ValidateProof(proof slash.SlashingProofHandler) error {
	switch proof.GetType() {
	case slash.None:
		return validateNoSlash(proof)
	case slash.MultipleProposal:
		return hsd.validateMultipleProposedHeaders(proof)
	default:
		return process.ErrInvalidSlashType
	}
}

func validateNoSlash(proof slash.SlashingProofHandler) error {
	if proof.GetLevel() != slash.Level0 {
		return process.ErrInvalidSlashLevel
	}
	return nil
}

func (hsd *HeaderSlashingDetector) validateMultipleProposedHeaders(proof slash.SlashingProofHandler) error {
	p, castOk := proof.(slash.MultipleProposalProofHandler)
	if !castOk {
		return process.ErrCannotCastProofToMultipleProposedHeaders
	}

	err := checkSlashTypeAndLevel(p)
	if err != nil {
		return err
	}

	return hsd.checkProposedHeaders(p.GetHeaders())
}

func checkSlashTypeAndLevel(proof slash.MultipleProposalProofHandler) error {
	headers := proof.GetHeaders()
	level := proof.GetLevel()

	if level < slash.Level1 || level > slash.Level2 {
		return process.ErrInvalidSlashLevel
	}
	if len(headers) < 2 {
		return process.ErrNotEnoughHeadersProvided
	}
	if len(headers) == 2 && level != slash.Level1 {
		return process.ErrSlashLevelDoesNotMatchSlashType
	}
	if len(headers) > 2 && level != slash.Level2 {
		return process.ErrSlashLevelDoesNotMatchSlashType
	}

	return nil
}

// Warning! This function should only be called after calling checkSlashTypeAndLevel
// Reason is that this function assumes that len(input headers) >= 2
func (hsd *HeaderSlashingDetector) checkProposedHeaders(headers []*interceptedBlocks.InterceptedHeader) error {
	hashes := make(map[string]struct{})
	round := headers[0].HeaderHandler().GetRound()
	proposer, err := hsd.getProposer(headers[0].HeaderHandler())
	if err != nil {
		return err
	}

	for _, header := range headers {
		hash := string(header.Hash())
		if _, exists := hashes[hash]; exists {
			return process.ErrProposedHeadersDoNotHaveDifferentHashes
		}

		err = hsd.checkHeaderHasSameProposerAndRound(header, round, proposer)
		if err != nil {
			return err
		}

		hashes[hash] = struct{}{}
	}

	return nil
}

func (hsd *HeaderSlashingDetector) checkHeaderHasSameProposerAndRound(
	header *interceptedBlocks.InterceptedHeader,
	round uint64,
	proposer []byte,
) error {
	if header.HeaderHandler().GetRound() != round {
		return process.ErrHeadersDoNotHaveSameRound
	}
	currProposer, err := hsd.getProposer(header.HeaderHandler())
	if err != nil {
		return err
	}

	if !bytes.Equal(proposer, currProposer) {
		return process.ErrHeadersDoNotHaveSameProposer
	}

	return nil
}
