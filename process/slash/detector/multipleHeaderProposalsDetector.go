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

// multipleHeaderProposalsDetector - checks slashable events in case a validator proposes multiple(possibly) malicious headers.
type multipleHeaderProposalsDetector struct {
	cache            RoundDetectorCache
	nodesCoordinator sharding.NodesCoordinator
	baseSlashingDetector
}

// NewMultipleHeaderProposalsDetector - creates a new multipleHeaderProposalsDetector for multiple headers
// proposal detection or multiple headers proposal proof verification
func NewMultipleHeaderProposalsDetector(
	nodesCoordinator sharding.NodesCoordinator,
	roundHandler process.RoundHandler,
	cache RoundDetectorCache,
) (slash.SlashingDetector, error) {
	if check.IfNil(nodesCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(roundHandler) {
		return nil, process.ErrNilRoundHandler
	}
	if check.IfNil(cache) {
		return nil, process.ErrNilRoundDetectorCache
	}

	baseDetector := baseSlashingDetector{roundHandler: roundHandler}

	return &multipleHeaderProposalsDetector{
		cache:                cache,
		nodesCoordinator:     nodesCoordinator,
		baseSlashingDetector: baseDetector,
	}, nil
}

// VerifyData - checks if an intercepted data(which should be a header) represents a slashable event.
// If another header with the same round and proposer exists, but a different hash, then a proof of type
// slash.MultipleProposal is provided, otherwise a nil proof, along with an error is provided indicating that
// no slashing event has been detected or an error occurred verifying the data.
func (mhp *multipleHeaderProposalsDetector) VerifyData(data process.InterceptedData) (slash.SlashingProofHandler, error) {
	interceptedHeader, castOk := data.(*interceptedBlocks.InterceptedHeader)
	if !castOk {
		return nil, process.ErrCannotCastInterceptedDataToHeader
	}

	header := interceptedHeader.HeaderHandler()
	if check.IfNil(header) {
		return nil, process.ErrNilHeaderHandler
	}

	round := header.GetRound()
	if !mhp.isRoundRelevant(round) {
		return nil, process.ErrHeaderRoundNotRelevant
	}

	proposer, err := mhp.getProposerPubKey(header)
	if err != nil {
		return nil, err
	}

	err = mhp.cache.Add(round, proposer, slash.HeaderInfo{Header: header, Hash: interceptedHeader.Hash()})
	if err != nil {
		return nil, err
	}

	slashingResult := mhp.getSlashingResult(round, proposer)

	if slashingResult != nil {
		return slash.NewMultipleProposalProof(slashingResult)
	}
	return nil, process.ErrNoSlashingEventDetected
}

func (mhp *multipleHeaderProposalsDetector) getProposerPubKey(header data.HeaderHandler) ([]byte, error) {
	validators, err := mhp.nodesCoordinator.ComputeConsensusGroup(
		header.GetPrevRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		header.GetEpoch())

	if err != nil {
		return nil, err
	}
	if len(validators) == 0 {
		return nil, process.ErrEmptyConsensusGroup
	}

	return validators[0].PubKey(), nil
}

func (mhp *multipleHeaderProposalsDetector) getSlashingResult(currRound uint64, proposerPubKey []byte) *slash.SlashingResult {
	proposedHeaders := mhp.cache.GetData(currRound, proposerPubKey)
	if len(proposedHeaders) >= 2 {
		return &slash.SlashingResult{
			SlashingLevel: mhp.computeSlashLevel(proposedHeaders),
			Data:          proposedHeaders,
		}
	}

	return nil
}

// TODO: Add different logic here once slashing threat levels are clearly defined
func (mhp *multipleHeaderProposalsDetector) computeSlashLevel(headers slash.HeaderInfoList) slash.ThreatLevel {
	return computeSlashLevelBasedOnHeadersCount(headers)
}

// ValidateProof - validates if the given proof is valid.
// For a proof of type slash.MultipleProposal to be valid, it should:
//  - Be of either level slash.Medium (with 2 proposed headers) OR slash.High (with >2 proposed headers)
//  - Have all proposed headers with the same round and proposer, but different hashes
func (mhp *multipleHeaderProposalsDetector) ValidateProof(proof slash.SlashingProofHandler) error {
	multipleProposalProof, castOk := proof.(slash.MultipleProposalProofHandler)
	if !castOk {
		return process.ErrCannotCastProofToMultipleProposedHeaders
	}
	if proof.GetType() != slash.MultipleProposal {
		return process.ErrInvalidSlashType
	}

	err := mhp.checkSlashLevel(multipleProposalProof.GetHeaders(), multipleProposalProof.GetLevel())
	if err != nil {
		return err
	}

	return mhp.checkProposedHeaders(multipleProposalProof.GetHeaders())
}

// TODO: Add different logic here once slashing threat levels are clearly defined
func (mhp *multipleHeaderProposalsDetector) checkSlashLevel(headers slash.HeaderInfoList, level slash.ThreatLevel) error {
	return checkSlashLevelBasedOnHeadersCount(headers, level)
}

func (mhp *multipleHeaderProposalsDetector) checkProposedHeaders(headers slash.HeaderInfoList) error {
	if len(headers) < minSlashableNoOfHeaders {
		return process.ErrNotEnoughHeadersProvided
	}

	hashes := make(map[string]struct{})
	round := headers[0].Header.GetRound()
	proposer, err := mhp.getProposerPubKey(headers[0].Header)
	if err != nil {
		return err
	}

	for _, header := range headers {
		hash := string(header.Hash)
		if _, exists := hashes[hash]; exists {
			return process.ErrHeadersNotDifferentHashes
		}

		err = mhp.checkHeaderHasSameProposerAndRound(header, round, proposer)
		if err != nil {
			return err
		}

		hashes[hash] = struct{}{}
	}

	return nil
}

func (mhp *multipleHeaderProposalsDetector) checkHeaderHasSameProposerAndRound(
	headerInfo slash.HeaderInfo,
	round uint64,
	proposer []byte,
) error {
	if headerInfo.Header.GetRound() != round {
		return process.ErrHeadersNotSameRound
	}

	currProposer, err := mhp.getProposerPubKey(headerInfo.Header)
	if err != nil {
		return err
	}

	if !bytes.Equal(proposer, currProposer) {
		return process.ErrHeadersNotSameProposer
	}

	return nil
}
