package detector

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// MultipleHeaderSingingDetectorArgs is a a struct containing all arguments required to create a new multipleHeaderSigningDetector
type MultipleHeaderSingingDetectorArgs struct {
	NodesCoordinator sharding.NodesCoordinator
	RoundHandler     process.RoundHandler
	Hasher           hashing.Hasher
	Marshaller       marshal.Marshalizer
	SlashingCache    RoundDetectorCache
	HeadersCache     HeadersCache
}

// multipleHeaderSigningDetector - checks for slashable events in case one(or more)
// validator signs multiple headers in the same round
type multipleHeaderSigningDetector struct {
	slashingCache    RoundDetectorCache
	headersCache     HeadersCache
	nodesCoordinator sharding.NodesCoordinator
	hasher           hashing.Hasher
	marshaller       marshal.Marshalizer
	baseSlashingDetector
}

// NewMultipleHeaderSigningDetector - creates a new header slashing detector for multiple signatures
func NewMultipleHeaderSigningDetector(args *MultipleHeaderSingingDetectorArgs) (slash.SlashingDetector, error) {
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
	if check.IfNil(args.SlashingCache) {
		return nil, process.ErrNilRoundDetectorCache
	}
	if check.IfNil(args.HeadersCache) {
		return nil, process.ErrNilRoundHeadersCache
	}

	baseDetector := baseSlashingDetector{roundHandler: args.RoundHandler}

	return &multipleHeaderSigningDetector{
		baseSlashingDetector: baseDetector,
		slashingCache:        args.SlashingCache,
		headersCache:         args.HeadersCache,
		nodesCoordinator:     args.NodesCoordinator,
		hasher:               args.Hasher,
		marshaller:           args.Marshaller,
	}, nil
}

// VerifyData - checks if an intercepted data represents a slashable event
func (mhs *multipleHeaderSigningDetector) VerifyData(interceptedData process.InterceptedData) (coreSlash.SlashingProofHandler, error) {
	interceptedHeader, castOk := interceptedData.(*interceptedBlocks.InterceptedHeader)
	if !castOk {
		return nil, process.ErrCannotCastInterceptedDataToHeader
	}

	header := interceptedHeader.HeaderHandler()
	round := header.GetRound()
	if !mhs.isRoundRelevant(round) {
		return nil, process.ErrHeaderRoundNotRelevant
	}

	err := mhs.cacheHeaderWithoutSignatures(header)
	if err != nil {
		return nil, err
	}

	err = mhs.cacheSigners(header, interceptedHeader.Hash())
	if err != nil {
		return nil, err
	}

	slashingResult := mhs.getSlashingResult(round)
	if len(slashingResult) != 0 {
		return coreSlash.NewMultipleSigningProof(slashingResult)
	}

	return nil, process.ErrNoSlashingEventDetected
}

func (mhs *multipleHeaderSigningDetector) cacheHeaderWithoutSignatures(header data.HeaderHandler) error {
	headerHash, err := mhs.computeHashWithoutSignatures(header)
	if err != nil {
		return err
	}
	headerInfo := &slash.HeaderInfo{Header: header, Hash: headerHash}

	return mhs.headersCache.Add(header.GetRound(), headerInfo)
}

func (mhs *multipleHeaderSigningDetector) computeHashWithoutSignatures(header data.HeaderHandler) ([]byte, error) {
	headerCopy := header.ShallowClone()
	err := headerCopy.SetPubKeysBitmap(nil)
	if err != nil {
		return nil, err
	}

	err = headerCopy.SetSignature(nil)
	if err != nil {
		return nil, err
	}

	err = headerCopy.SetLeaderSignature(nil)
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
	for idx, validator := range group {
		if slash.IsIndexSetInBitmap(uint32(idx), bitmap) {
			headerInfo := &slash.HeaderInfo{Header: header, Hash: hash}

			err = mhs.slashingCache.Add(header.GetRound(), validator.PubKey(), headerInfo)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (mhs *multipleHeaderSigningDetector) getSlashingResult(round uint64) map[string]coreSlash.SlashingResult {
	slashingData := make(map[string]coreSlash.SlashingResult)

	for _, validator := range mhs.slashingCache.GetPubKeys(round) {
		signedHeaders := mhs.slashingCache.GetHeaders(round, validator)
		if len(signedHeaders) > 1 {
			slashingData[string(validator)] = coreSlash.SlashingResult{
				SlashingLevel: mhs.computeSlashLevel(signedHeaders),
				Headers:       signedHeaders,
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
	multipleSigningProof, castOk := proof.(coreSlash.MultipleSigningProofHandler)
	if !castOk {
		return process.ErrCannotCastProofToMultipleSignedHeaders
	}
	if multipleSigningProof.GetType() != coreSlash.MultipleSigning {
		return process.ErrInvalidSlashType
	}

	signers := multipleSigningProof.GetPubKeys()
	if len(signers) == 0 {
		return process.ErrNotEnoughPubKeysProvided
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
	return checkSlashLevelBasedOnHeadersCount(headers, level)
}

func (mhs *multipleHeaderSigningDetector) checkSignedHeaders(pubKey []byte, headers slash.HeaderList) error {
	if len(headers) < minSlashableNoOfHeaders {
		return process.ErrNotEnoughHeadersProvided
	}

	hashes := make(map[string]struct{})
	round := headers[0].GetRound()
	for _, header := range headers {
		if header.GetRound() != round {
			return process.ErrHeadersNotSameRound
		}

		hash, err := mhs.checkHash(header, hashes)
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

func (mhs *multipleHeaderSigningDetector) checkHash(header data.HeaderHandler, hashes map[string]struct{}) (string, error) {
	hash, err := mhs.computeHashWithoutSignatures(header)
	if err != nil {
		return "", err
	}

	if _, exists := hashes[string(hash)]; exists {
		return "", process.ErrHeadersNotDifferentHashes
	}

	return string(hash), nil
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

	for idx, validator := range group {
		if bytes.Equal(validator.PubKey(), pubKey) &&
			slash.IsIndexSetInBitmap(uint32(idx), header.GetPubKeysBitmap()) {
			return true
		}
	}

	return false
}
