package detector

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// MultipleHeaderProposalDetectorArgs is a a struct containing all arguments required to create a new multipleHeaderProposalsDetector
type MultipleHeaderProposalDetectorArgs struct {
	NodesCoordinator  sharding.NodesCoordinator
	RoundHandler      process.RoundHandler
	Cache             RoundValidatorHeadersCache
	Hasher            hashing.Hasher
	Marshaller        marshal.Marshalizer
	HeaderSigVerifier consensus.HeaderSigVerifier
}

// multipleHeaderProposalsDetector - checks slashable events in case a validator proposes multiple(possibly) malicious headers.
type multipleHeaderProposalsDetector struct {
	cache             RoundValidatorHeadersCache
	nodesCoordinator  sharding.NodesCoordinator
	hasher            hashing.Hasher
	marshaller        marshal.Marshalizer
	headerSigVerifier consensus.HeaderSigVerifier
	baseSlashingDetector
}

// NewMultipleHeaderProposalsDetector - creates a new multipleHeaderProposalsDetector for multiple headers
// proposal detection or multiple headers proposal proof verification
func NewMultipleHeaderProposalsDetector(args *MultipleHeaderProposalDetectorArgs) (slash.SlashingDetector, error) {
	if args == nil {
		return nil, process.ErrNilMultipleHeaderProposalDetectorArgs
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.RoundHandler) {
		return nil, process.ErrNilRoundHandler
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Cache) {
		return nil, process.ErrNilRoundDetectorCache
	}
	if check.IfNil(args.HeaderSigVerifier) {
		return nil, process.ErrNilHeaderSigVerifier
	}

	baseDetector := baseSlashingDetector{roundHandler: args.RoundHandler}

	return &multipleHeaderProposalsDetector{
		cache:                args.Cache,
		nodesCoordinator:     args.NodesCoordinator,
		hasher:               args.Hasher,
		marshaller:           args.Marshaller,
		headerSigVerifier:    args.HeaderSigVerifier,
		baseSlashingDetector: baseDetector,
	}, nil
}

// VerifyData - checks if an intercepted data(which should be a header) represents a slashable event.
// If another header with the same round and proposer exists, but a different hash, then a proof of type
// slash.MultipleProposal is provided, otherwise a nil proof, along with an error is provided indicating that
// no slashing event has been detected or an error occurred verifying the data.
func (mhp *multipleHeaderProposalsDetector) VerifyData(interceptedData process.InterceptedData) (coreSlash.SlashingProofHandler, error) {
	header, err := checkAndGetHeader(interceptedData)
	if err != nil {
		return nil, err
	}

	round := header.GetRound()
	if !mhp.isRoundRelevant(round) {
		return nil, process.ErrHeaderRoundNotRelevant
	}

	proposer, err := mhp.getProposerPubKey(header)
	if err != nil {
		return nil, err
	}

	err = mhp.cache.Add(round, proposer, &slash.HeaderInfo{Header: header, Hash: interceptedData.Hash()})
	if err != nil {
		return nil, err
	}

	slashingResult := mhp.getSlashingResult(round, proposer)
	if slashingResult != nil {
		return coreSlash.NewMultipleProposalProof(slashingResult)
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

func (mhp *multipleHeaderProposalsDetector) getSlashingResult(currRound uint64, proposerPubKey []byte) *coreSlash.SlashingResult {
	proposedHeadersInfo := mhp.cache.GetHeaders(currRound, proposerPubKey)
	if len(proposedHeadersInfo) >= minSlashableNoOfHeaders {
		headerHandlers := getHeaderHandlers(proposedHeadersInfo)
		return &coreSlash.SlashingResult{
			SlashingLevel: mhp.computeSlashLevel(headerHandlers),
			Headers:       proposedHeadersInfo,
		}
	}

	return nil
}

// TODO: Add different logic here once slashing threat levels are clearly defined
func (mhp *multipleHeaderProposalsDetector) computeSlashLevel(headers []data.HeaderHandler) coreSlash.ThreatLevel {
	return computeSlashLevelBasedOnHeadersCount(headers)
}

// ValidateProof - validates if the given proof is valid.
// For a proof of type slash.MultipleProposal to be valid, it should:
//  - Be of either level slash.Medium (with 2 proposed headers) OR slash.High (with >2 proposed headers)
//  - Have all proposed headers with the same round and proposer, but different hashes
func (mhp *multipleHeaderProposalsDetector) ValidateProof(proof coreSlash.SlashingProofHandler) error {
	err := checkProofType(proof, coreSlash.MultipleProposal)
	if err != nil {
		return err
	}

	multipleProposalProof, castOk := proof.(coreSlash.MultipleProposalProofHandler)
	if !castOk {
		return process.ErrCannotCastProofToMultipleProposedHeaders
	}

	err = mhp.checkThreatLevel(multipleProposalProof.GetHeaders(), multipleProposalProof.GetLevel())
	if err != nil {
		return err
	}

	return mhp.checkProposedHeaders(multipleProposalProof.GetHeaders())
}

// TODO: Add different logic here once slashing threat levels are clearly defined
func (mhp *multipleHeaderProposalsDetector) checkThreatLevel(headers []data.HeaderHandler, level coreSlash.ThreatLevel) error {
	return checkThreatLevelBasedOnHeadersCount(headers, level)
}

func (mhp *multipleHeaderProposalsDetector) checkProposedHeaders(headers []data.HeaderHandler) error {
	if len(headers) < minSlashableNoOfHeaders {
		return process.ErrNotEnoughHeadersProvided
	}
	if check.IfNil(headers[0]) {
		return process.ErrNilHeaderHandler
	}

	round := headers[0].GetRound()
	proposer, err := mhp.getProposerPubKey(headers[0])
	if err != nil {
		return err
	}

	hashes := make(map[string]struct{})
	for _, header := range headers {
		if check.IfNil(header) {
			return process.ErrNilHeaderHandler
		}
		hash, err := mhp.checkHash(header, hashes)
		if err != nil {
			return err
		}
		err = mhp.checkHeaderHasSameProposerAndRound(header, round, proposer)
		if err != nil {
			return err
		}

		hashes[hash] = struct{}{}
	}

	return nil
}

func (mhp *multipleHeaderProposalsDetector) checkHash(header data.HeaderHandler, hashes map[string]struct{}) (string, error) {
	hash, err := core.CalculateHash(mhp.marshaller, mhp.hasher, header)
	if err != nil {
		return "", err
	}

	if _, exists := hashes[string(hash)]; exists {
		return "", process.ErrHeadersNotDifferentHashes
	}

	return string(hash), nil
}

func (mhp *multipleHeaderProposalsDetector) checkHeaderHasSameProposerAndRound(
	header data.HeaderHandler,
	round uint64,
	proposer []byte,
) error {
	if header.GetRound() != round {
		return process.ErrHeadersNotSameRound
	}

	currProposer, err := mhp.getProposerPubKey(header)
	if err != nil {
		return err
	}

	if !bytes.Equal(proposer, currProposer) {
		return process.ErrHeadersNotSameProposer
	}

	return mhp.headerSigVerifier.VerifyLeaderSignature(header)
}
