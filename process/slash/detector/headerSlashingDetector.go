package detector

import (
	"bytes"
	"errors"

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

// ValidateProof - validates the given proof
func (hsd *HeaderSlashingDetector) ValidateProof(proof slash.SlashingProofHandler) error {
	switch proof.GetType() {
	case slash.None:
		if proof.GetLevel() == slash.Level0 {
			return nil
		}
	case slash.MultipleProposal:
		return hsd.validateMultipleProposedHeaders(proof)
	default:
		return errors.New("not handled")
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

func (hsd *HeaderSlashingDetector) validateMultipleProposedHeaders(proof slash.SlashingProofHandler) error {
	p, castOk := proof.(slash.MultipleProposalProofHandler)
	if !castOk {
		return process.ErrCannotCastProofToMultipleProposedHeaders
	}

	if p.GetType() != slash.MultipleProposal {
		return process.ErrInvalidSlashType
	}

	if checkSlashLevel(p) != true {
		return process.ErrSlashLevelDoesNotMatchSlashType
	}

	return hsd.checkProposedHeaders(p.GetHeaders())
}

func checkSlashLevel(proof slash.MultipleProposalProofHandler) bool {
	headers := proof.GetHeaders()

	if len(headers) < 2 {
		return false
	}
	if len(headers) == 2 && proof.GetLevel() != slash.Level1 {
		return false
	}
	if len(headers) > 2 && proof.GetLevel() != slash.Level2 {
		return false
	}

	return true
}

// Warning! This function should only be called after calling checkSlashLevel
// Reason is that this function assumes that len(input headers) >= 1
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
	if header.HeaderHandler().GetRound() == round {
		currProposer, err := hsd.getProposer(header.HeaderHandler())
		if err != nil {
			return err
		}

		if !bytes.Equal(proposer, currProposer) {
			return process.ErrHeadersDoNotHaveSameProposer
		}
	}

	return nil
}
