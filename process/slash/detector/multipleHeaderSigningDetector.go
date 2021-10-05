package detector

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// SigningSlashingDetector - checks for slashable events for headers
type headersCache interface {
	add(round uint64, hash []byte, header data.HeaderHandler)
	contains(round uint64, hash []byte) bool
	headers(round uint64) headerHashList
}

type SigningSlashingDetector struct {
	slashingCache    detectorCache
	headersCache     headersCache
	nodesCoordinator sharding.NodesCoordinator
	hasher           hashing.Hasher
	marshaller       marshal.Marshalizer
	baseSlashingDetector
}

// NewSigningSlashingDetector - creates a new header slashing detector for multiple signatures
func NewSigningSlashingDetector(
	nodesCoordinator sharding.NodesCoordinator,
	roundHandler process.RoundHandler,
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
	maxRoundCacheSize uint64,
) (slash.SlashingDetector, error) {
	if check.IfNil(nodesCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(roundHandler) {
		return nil, process.ErrNilRoundHandler
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}

	//TODO: Here, instead of CacheSize, use maxRoundCacheSize = from config file
	slashingCache := newRoundProposerDataCache(CacheSize)
	headersCache := newRoundHeadersCache(CacheSize)
	baseDetector := baseSlashingDetector{roundHandler: roundHandler}

	return &SigningSlashingDetector{
		slashingCache:        slashingCache,
		headersCache:         headersCache,
		nodesCoordinator:     nodesCoordinator,
		baseSlashingDetector: baseDetector,
		hasher:               hasher,
		marshaller:           marshaller,
	}, nil
}

// VerifyData - checks if an intercepted data represents a slashable event
func (ssd *SigningSlashingDetector) VerifyData(interceptedData process.InterceptedData) (slash.SlashingProofHandler, error) {
	header, castOk := interceptedData.(*interceptedBlocks.InterceptedHeader)
	if !castOk {
		return nil, process.ErrCannotCastInterceptedDataToHeader
	}

	round := header.HeaderHandler().GetRound()
	if !ssd.isRoundRelevant(round) {
		return nil, process.ErrHeaderRoundNotRelevant
	}

	headerHash, err := ssd.computeHashWithoutSignatures(header.HeaderHandler())
	if err != nil {
		return nil, err
	}

	if ssd.headersCache.contains(round, headerHash) {
		return nil, process.ErrHeadersShouldHaveDifferentHashes
	}

	slashingData := make(map[string]slash.SlashingData)
	/*
		for _, currData := range ssd.headersCache.headers(round) {
			doubleSigners, err := ssd.doubleSigningValidators(currData.header, header.HeaderHandler())
			if err != nil {
				return nil, err
			}
			if len(doubleSigners) >= 1 {
				for _, doubleSigner := range doubleSigners {
					ssd.slashingCache.add(round, doubleSigner.PubKey(), header)
					headers := ssd.slashingCache.data(round, doubleSigner.PubKey())

					slashingData[string(doubleSigner.PubKey())] = slash.SlashingData{
						SlashingLevel: computeSlashLevel(headers),
						Data:          headers,
					}
				}
			}
		}
	*/

	group, err := ssd.nodesCoordinator.ComputeConsensusGroup(
		header.HeaderHandler().GetPrevRandSeed(),
		header.HeaderHandler().GetRound(),
		header.HeaderHandler().GetShardID(),
		header.HeaderHandler().GetEpoch())
	if err != nil {
		return nil, err
	}
	bitmap := header.HeaderHandler().GetPubKeysBitmap()

	for idx, validator := range group {
		ssd.checkIfValidatorSignedAndCacheHim(round, validator.PubKey(), header, uint32(idx), bitmap)
	}

	for _, validator := range ssd.slashingCache.validators(round) {
		signedHeaders := ssd.slashingCache.data(round, validator)
		if len(signedHeaders) > 1 {
			slashingData[string(validator)] = slash.SlashingData{
				SlashingLevel: computeSlashLevel(signedHeaders),
				Data:          signedHeaders,
			}
		}
	}

	ssd.headersCache.add(round, headerHash, header.HeaderHandler())
	if len(slashingData) != 0 {
		return slash.NewMultipleSigningProof(slashingData)
	}

	// check another signature with the same round and proposer exists, but a different header exists
	// if yes a slashingDetectorResult is returned with a message and the two signatures
	return nil, process.ErrNoSlashingEventDetected
}

func (ssd *SigningSlashingDetector) computeHashWithoutSignatures(header data.HeaderHandler) ([]byte, error) {
	headerCopy := header.Clone()
	headerCopy.SetPubKeysBitmap(nil)
	headerCopy.SetSignature(nil)
	headerCopy.SetLeaderSignature(nil)

	headerBytes, err := ssd.marshaller.Marshal(headerCopy)
	if err != nil {
		return nil, err
	}

	return ssd.hasher.Compute(string(headerBytes)), nil
}

func (ssd *SigningSlashingDetector) doubleSigningValidators(
	round uint64,
	header1 *interceptedBlocks.InterceptedHeader,
	header2 *interceptedBlocks.InterceptedHeader,
) ([]sharding.Validator, error) {
	group1, err := ssd.nodesCoordinator.ComputeConsensusGroup(
		header1.HeaderHandler().GetPrevRandSeed(),
		header1.HeaderHandler().GetRound(),
		header1.HeaderHandler().GetShardID(),
		header1.HeaderHandler().GetEpoch())
	if err != nil {
		return nil, err
	}

	group2, err := ssd.nodesCoordinator.ComputeConsensusGroup(
		header2.HeaderHandler().GetPrevRandSeed(),
		header2.HeaderHandler().GetRound(),
		header2.HeaderHandler().GetShardID(),
		header2.HeaderHandler().GetEpoch())
	if err != nil {
		return nil, err
	}

	bitmap1 := header1.HeaderHandler().GetPubKeysBitmap()
	bitmap2 := header2.HeaderHandler().GetPubKeysBitmap()

	intersectionSet := make([]sharding.Validator, 0)
	pubKeyIdxMap := make(map[string]uint32, 0)
	for idx, validator := range group1 {
		if ssd.checkIfValidatorSignedAndCacheHim(round, validator.PubKey(), header1, uint32(idx), bitmap1) {
			pubKeyIdxMap[string(validator.PubKey())] = uint32(idx)
		}
	}

	for idxGroup2, validatorGroup2 := range group2 {
		if _, found := pubKeyIdxMap[string(validatorGroup2.PubKey())]; found {
			if ssd.checkIfValidatorSignedAndCacheHim(round, validatorGroup2.PubKey(), header2, uint32(idxGroup2), bitmap2) {
				intersectionSet = append(intersectionSet, validatorGroup2)
			}
		}
	}
	return intersectionSet, nil

}

func isIndexSetInBitmap(index uint32, bitmap []byte) bool {
	indexOutOfBounds := index >= uint32(len(bitmap))*8
	if indexOutOfBounds {
		return false
	}

	return bitmap[index/8]&(1<<uint8(index%8)) != 0
}

func (ssd *SigningSlashingDetector) checkIfValidatorSignedAndCacheHim(
	round uint64,
	pubKey []byte,
	header *interceptedBlocks.InterceptedHeader,
	idx uint32,
	bitmap []byte,
) bool {
	if isIndexSetInBitmap(idx, bitmap) {
		ssd.slashingCache.add(round, pubKey, header)
		return true
	}
	return false
}

func (ssd *SigningSlashingDetector) doubleSigners(
	round uint64,
	group1 []sharding.Validator,
	group2 []sharding.Validator,
	bitmap1 []byte,
	bitmap2 []byte,
) []sharding.Validator {
	intersectionSet := make([]sharding.Validator, 0)
	if len(group1) == 0 || len(group2) == 0 {
		return intersectionSet
	}

	pubKeyIdxMap := make(map[string]uint32, 0)
	for idx, validator := range group1 {
		if isIndexSetInBitmap(uint32(idx), bitmap1) {
			pubKeyIdxMap[string(validator.PubKey())] = uint32(idx)
		}
	}

	for idxGroup2, validatorGroup2 := range group2 {
		if idxGroup1, found := pubKeyIdxMap[string(validatorGroup2.PubKey())]; found {
			if isIndexSetInBitmap(idxGroup1, bitmap1) &&
				isIndexSetInBitmap(uint32(idxGroup2), bitmap2) {
				intersectionSet = append(intersectionSet, validatorGroup2)
			}
		}
	}
	return intersectionSet
}

// ValidateProof - validates the given proof
func (ssd *SigningSlashingDetector) ValidateProof(proof slash.SlashingProofHandler) error {
	return nil
}
