package detector

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type headersCache interface {
	add(round uint64, hash []byte, header data.HeaderHandler)
	contains(round uint64, hash []byte) bool
	headers(round uint64) headerHashList
}

// SigningSlashingDetector - checks for slashable events in case one(or more)
// validator signs multiple headers in the same round
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

	err = ssd.cacheSigners(header)
	if err != nil {
		return nil, err
	}
	ssd.headersCache.add(round, headerHash, header.HeaderHandler())

	slashingData := ssd.getSlashingResult(round)
	if len(slashingData) != 0 {
		return slash.NewMultipleSigningProof(slashingData)
	}

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

func (ssd *SigningSlashingDetector) cacheSigners(interceptedHeader *interceptedBlocks.InterceptedHeader) error {
	header := interceptedHeader.HeaderHandler()
	group, err := ssd.nodesCoordinator.ComputeConsensusGroup(
		header.GetPrevRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		header.GetEpoch())
	if err != nil {
		return err
	}

	bitmap := header.GetPubKeysBitmap()
	for idx, validator := range group {
		if isIndexSetInBitmap(uint32(idx), bitmap) {
			ssd.slashingCache.add(header.GetRound(), validator.PubKey(), interceptedHeader)
		}
	}

	return nil
}

func (ssd *SigningSlashingDetector) getSlashingResult(round uint64) map[string]slash.SlashingData {
	slashingData := make(map[string]slash.SlashingData)

	for _, validator := range ssd.slashingCache.validators(round) {
		signedHeaders := ssd.slashingCache.data(round, validator)
		if len(signedHeaders) > 1 {
			slashingData[string(validator)] = slash.SlashingData{
				SlashingLevel: computeSlashLevel(signedHeaders),
				Data:          signedHeaders,
			}
		}
	}

	return slashingData
}

func isIndexSetInBitmap(index uint32, bitmap []byte) bool {
	indexOutOfBounds := index >= uint32(len(bitmap))*8
	if indexOutOfBounds {
		return false
	}

	return bitmap[index/8]&(1<<uint8(index%8)) != 0
}

// ValidateProof - validates the given proof
func (ssd *SigningSlashingDetector) ValidateProof(proof slash.SlashingProofHandler) error {
	multipleSigningProof, castOk := proof.(slash.MultipleSigningProofHandler)
	if !castOk {
		return process.ErrCannotCastProofToMultipleSignedHeaders
	}
	if multipleSigningProof.GetType() != slash.MultipleSigning {
		return process.ErrInvalidSlashType
	}

	signers := multipleSigningProof.GetPubKeys()
	for _, signer := range signers {
		err := ssd.checkSignedHeaders(signer, multipleSigningProof.GetHeaders(signer), multipleSigningProof.GetLevel(signer))
		if err != nil {
			return err
		}
	}

	return nil
}

func (ssd *SigningSlashingDetector) checkSignedHeaders(
	pubKey []byte,
	headers []*interceptedBlocks.InterceptedHeader,
	level slash.SlashingLevel,
) error {
	if len(headers) < MinSlashableNoOfHeaders {
		return process.ErrNotEnoughHeadersProvided
	}

	hashes := make(map[string]struct{})
	round := headers[0].HeaderHandler().GetRound()
	for _, header := range headers {
		if header.HeaderHandler().GetRound() != round {
			return process.ErrHeadersDoNotHaveSameRound
		}

		err := checkSlashLevel(headers, level)
		if err != nil {
			return err
		}

		hash, err := ssd.checkHash(header.HeaderHandler(), hashes)
		if err != nil {
			return err
		}

		if !ssd.signedHeader(pubKey, header.HeaderHandler()) {
			return process.ErrHeaderNotSignedByValidator
		}
		hashes[hash] = struct{}{}
	}

	return nil
}

func (ssd *SigningSlashingDetector) checkHash(header data.HeaderHandler, hashes map[string]struct{}) (string, error) {
	hash, err := ssd.computeHashWithoutSignatures(header)
	if err != nil {
		return "", err
	}

	if _, exists := hashes[string(hash)]; exists {
		return "", process.ErrHeadersShouldHaveDifferentHashes
	}

	return string(hash), nil
}

func (ssd *SigningSlashingDetector) signedHeader(pubKey []byte, header data.HeaderHandler) bool {
	group, err := ssd.nodesCoordinator.ComputeConsensusGroup(
		header.GetPrevRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		header.GetEpoch())
	if err != nil {
		return false
	}

	for idx, validator := range group {
		if bytes.Equal(validator.PubKey(), pubKey) &&
			isIndexSetInBitmap(uint32(idx), header.GetPubKeysBitmap()) {
			return true
		}
	}

	return false
}
