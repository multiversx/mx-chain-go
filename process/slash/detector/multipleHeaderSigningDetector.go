package detector

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("process/slash/detector/multipleHeaderSigning")

// MultipleHeaderSigningDetectorArgs is a a struct containing all arguments required to create a new multipleHeaderSigningDetector
type MultipleHeaderSigningDetectorArgs struct {
	NodesCoordinator           sharding.NodesCoordinator
	RoundHandler               process.RoundHandler
	Hasher                     hashing.Hasher
	Marshaller                 marshal.Marshalizer
	RoundHashCache             RoundHashCache
	RoundValidatorHeadersCache RoundValidatorHeadersCache
	HeaderSigVerifier          consensus.HeaderSigVerifier
}

// multipleHeaderSigningDetector - checks for slashable events in case one(or more)
// validator signs multiple headers in the same round
type multipleHeaderSigningDetector struct {
	slashingCache     RoundValidatorHeadersCache
	hashesCache       RoundHashCache
	nodesCoordinator  sharding.NodesCoordinator
	hasher            hashing.Hasher
	marshaller        marshal.Marshalizer
	cachesMutex       sync.RWMutex
	headerSigVerifier consensus.HeaderSigVerifier
	baseSlashingDetector
}

// NewMultipleHeaderSigningDetector - creates a new header slashing detector for multiple signatures
func NewMultipleHeaderSigningDetector(args *MultipleHeaderSigningDetectorArgs) (*multipleHeaderSigningDetector, error) {
	if args == nil {
		return nil, process.ErrNilMultipleHeaderSigningDetectorArgs
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
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
	if check.IfNil(args.RoundValidatorHeadersCache) {
		return nil, process.ErrNilRoundValidatorHeadersCache
	}
	if check.IfNil(args.RoundHashCache) {
		return nil, process.ErrNilRoundHeadersCache
	}
	if check.IfNil(args.HeaderSigVerifier) {
		return nil, process.ErrNilHeaderSigVerifier
	}

	baseDetector := baseSlashingDetector{roundHandler: args.RoundHandler}

	return &multipleHeaderSigningDetector{
		baseSlashingDetector: baseDetector,
		slashingCache:        args.RoundValidatorHeadersCache,
		hashesCache:          args.RoundHashCache,
		nodesCoordinator:     args.NodesCoordinator,
		hasher:               args.Hasher,
		marshaller:           args.Marshaller,
		headerSigVerifier:    args.HeaderSigVerifier,
		cachesMutex:          sync.RWMutex{},
	}, nil
}

// VerifyData - checks if an intercepted data represents a slashable event
func (mhs *multipleHeaderSigningDetector) VerifyData(interceptedData process.InterceptedData) (coreSlash.SlashingProofHandler, error) {
	header, err := getCheckedHeader(interceptedData)
	if err != nil {
		return nil, err
	}

	round := header.GetRound()
	if !mhs.isRoundRelevant(round) {
		return nil, process.ErrHeaderRoundNotRelevant
	}

	mhs.cachesMutex.Lock()
	defer mhs.cachesMutex.Unlock()

	headerHashWithoutSignatures, err := mhs.cacheHeaderHashWithoutSignatures(header)
	if err != nil {
		return nil, err
	}

	err = mhs.cacheSigners(header, interceptedData.Hash())
	if err != nil {
		mhs.hashesCache.Remove(round, headerHashWithoutSignatures)
		return nil, err
	}

	slashingResult := mhs.getSlashingResult(round)
	if len(slashingResult) != 0 {
		return coreSlash.NewMultipleSigningProof(slashingResult)
	}

	return nil, process.ErrNoSlashingEventDetected
}

func (mhs *multipleHeaderSigningDetector) cacheHeaderHashWithoutSignatures(header data.HeaderHandler) ([]byte, error) {
	headerHash, err := mhs.computeHashWithoutSignatures(header)
	if err != nil {
		return nil, err
	}

	err = mhs.hashesCache.Add(header.GetRound(), headerHash)
	if err != nil {
		return nil, err
	}

	return headerHash, nil
}

func (mhs *multipleHeaderSigningDetector) computeHashWithoutSignatures(header data.HeaderHandler) ([]byte, error) {
	headerCopy, err := process.CopyHeaderWithoutSig(header)
	if err != nil {
		return nil, err
	}

	return core.CalculateHash(mhs.marshaller, mhs.hasher, headerCopy)
}

func (mhs *multipleHeaderSigningDetector) cacheSigners(header data.HeaderHandler, hash []byte) error {
	group, err := mhs.nodesCoordinator.ComputeConsensusGroup(
		header.GetPrevRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		header.GetEpoch())
	if err != nil {
		return err
	}

	bitmap := header.GetPubKeysBitmap()
	headerInfo := &slash.HeaderInfo{Header: header, Hash: hash}
	for idx, validator := range group {
		if sliceUtil.IsIndexSetInBitmap(uint32(idx), bitmap) {
			// We could never have an error here, since an error could only happen if:
			// 1. Round is irrelevant = false, because it is checked before calling this func
			// 2. Header is already cached = false, because it is already checked before using mhs.hashesCache
			err = mhs.slashingCache.Add(header.GetRound(), validator.PubKey(), headerInfo)
			log.LogIfError(err)
		}
	}

	return nil
}

func (mhs *multipleHeaderSigningDetector) getSlashingResult(round uint64) map[string]coreSlash.SlashingResult {
	slashingData := make(map[string]coreSlash.SlashingResult)

	for _, validator := range mhs.slashingCache.GetPubKeys(round) {
		signedHeadersInfo := mhs.slashingCache.GetHeaders(round, validator)
		if len(signedHeadersInfo) > 1 {
			headerHandlers := getHeaderHandlers(signedHeadersInfo)
			slashingData[string(validator)] = coreSlash.SlashingResult{
				SlashingLevel: mhs.computeSlashLevel(headerHandlers),
				Headers:       signedHeadersInfo,
			}
		}
	}

	return slashingData
}

// TODO: Add different logic here once slashing threat levels are clearly defined
func (mhs *multipleHeaderSigningDetector) computeSlashLevel(headers slash.HeaderList) coreSlash.ThreatLevel {
	return computeSlashLevelBasedOnHeadersCount(headers)
}

// ValidateProof - validates the given proof
func (mhs *multipleHeaderSigningDetector) ValidateProof(proof coreSlash.SlashingProofHandler) error {
	err := checkProofType(proof, coreSlash.MultipleSigning)
	if err != nil {
		return err
	}
	multipleSigningProof, castOk := proof.(coreSlash.MultipleSigningProofHandler)
	if !castOk {
		return process.ErrCannotCastProofToMultipleSignedHeaders
	}

	signers := multipleSigningProof.GetPubKeys()
	if len(signers) == 0 {
		return process.ErrNotEnoughPublicKeysProvided
	}

	for _, signer := range signers {
		err := mhs.checkSlashLevel(multipleSigningProof.GetHeaders(signer), multipleSigningProof.GetLevel(signer))
		if err != nil {
			return err
		}

		err = mhs.checkSignedHeaders(signer, multipleSigningProof.GetHeaders(signer))
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Add different logic here once slashing threat levels are clearly defined
func (mhs *multipleHeaderSigningDetector) checkSlashLevel(headers slash.HeaderList, level coreSlash.ThreatLevel) error {
	return checkThreatLevelBasedOnHeadersCount(headers, level)
}

func (mhs *multipleHeaderSigningDetector) checkSignedHeaders(pubKey []byte, headers slash.HeaderList) error {
	if len(headers) < minSlashableNoOfHeaders {
		return process.ErrNotEnoughHeadersProvided
	}
	if check.IfNil(headers[0]) {
		return process.ErrNilHeaderHandler
	}

	round := headers[0].GetRound()
	hashes := make(map[string]struct{})
	for _, header := range headers {
		if check.IfNil(header) {
			return process.ErrNilHeaderHandler
		}
		if header.GetRound() != round {
			return process.ErrHeadersNotSameRound
		}

		hash, err := mhs.checkHashWithoutSigExists(hashes, header)
		if err != nil {
			return err
		}

		if !mhs.signedHeader(pubKey, header) {
			return process.ErrHeaderNotSignedByValidator
		}

		hashes[hash] = struct{}{}
	}

	return nil
}

func (mhs *multipleHeaderSigningDetector) checkHashWithoutSigExists(hashes map[string]struct{}, header data.HeaderHandler) (string, error) {
	hash, err := mhs.computeHashWithoutSignatures(header)
	if err != nil {
		return "", err
	}

	hashStr := string(hash)
	if _, exists := hashes[hashStr]; exists {
		return "", process.ErrHeadersNotDifferentHashes
	}

	return hashStr, nil
}

func (mhs *multipleHeaderSigningDetector) signedHeader(pubKey []byte, header data.HeaderHandler) bool {
	group, err := mhs.nodesCoordinator.ComputeConsensusGroup(
		header.GetPrevRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		header.GetEpoch())
	if err != nil {
		return false
	}

	bitmap := header.GetPubKeysBitmap()
	for idx, validator := range group {
		currPubKey := validator.PubKey()
		if bytes.Equal(currPubKey, pubKey) &&
			sliceUtil.IsIndexSetInBitmap(uint32(idx), bitmap) &&
			mhs.headerSigVerifier.VerifySignature(header) == nil {
			return true
		}
	}

	return false
}
