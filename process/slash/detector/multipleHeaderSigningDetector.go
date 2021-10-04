package detector

import (
	"errors"

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
		return nil, errors.New("dsa")
	}

	validatorSignedHeaders := make(map[string]slash.DataWithSlashingLevel)
	for _, currData := range ssd.headersCache.headers(round) {
		signers, _ := ssd.getValidatorsWhichSignedBothHeaders(currData.header, header.HeaderHandler())
		if len(signers) >= 1 {
			for _, signer := range signers {
				ssd.slashingCache.add(round, signer.PubKey(), header)

				validatorSignedHeaders[string(signer.PubKey())] = slash.DataWithSlashingLevel{
					SlashingLevel: 0,
					Data:          ssd.slashingCache.proposedData(round, signer.PubKey()),
				}
			}
		}
	}

	ssd.headersCache.add(round, headerHash, header.HeaderHandler())

	if len(validatorSignedHeaders) != 0 {
		return slash.NewMultipleHeaderSigningProof(validatorSignedHeaders)
	}

	// check another signature with the same round and proposer exists, but a different header exists
	// if yes a slashingDetectorResult is returned with a message and the two signatures
	return nil, process.ErrNoSlashingEventDetected
}

func (ssd *SigningSlashingDetector) computeHashWithoutSignatures(header data.HeaderHandler) ([]byte, error) {
	header.SetPubKeysBitmap(nil)
	header.SetSignature(nil)
	header.SetLeaderSignature(nil)

	headerBytes, err := ssd.marshaller.Marshal(header)
	if err != nil {
		return nil, err
	}

	return ssd.hasher.Compute(string(headerBytes)), nil
}

func (ssd *SigningSlashingDetector) getValidatorsWhichSignedBothHeaders(header1 data.HeaderHandler, header2 data.HeaderHandler) ([]sharding.Validator, error) {
	group1, err := ssd.nodesCoordinator.ComputeConsensusGroup(
		header1.GetPrevRandSeed(),
		header1.GetRound(),
		header1.GetShardID(),
		header1.GetEpoch())
	if err != nil {
		return nil, err
	}

	group2, err := ssd.nodesCoordinator.ComputeConsensusGroup(
		header2.GetPrevRandSeed(),
		header2.GetRound(),
		header2.GetShardID(),
		header2.GetEpoch())
	if err != nil {
		return nil, err
	}

	pubKeyBitmap1 := header1.GetPubKeysBitmap()
	pubKeyBitmap2 := header2.GetPubKeysBitmap()

	return doubleSigners(group1, group2, pubKeyBitmap1, pubKeyBitmap2), nil
}

func isIndexSetInBitmap(index uint32, bitmap []byte) bool {
	indexOutOfBounds := index >= uint32(len(bitmap))*8
	if indexOutOfBounds {
		return false
	}

	return bitmap[index/8]&(1<<uint8(index%8)) != 0
}

func doubleSigners(
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
		pubKeyIdxMap[string(validator.PubKey())] = uint32(idx)
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
