package detector

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// signingSlashingDetector - checks for slashable events in case one(or more)
// validator signs multiple headers in the same round
type signingSlashingDetector struct {
	slashingCache    RoundDetectorCache
	headersCache     HeadersCache
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
	slashingCache RoundDetectorCache,
	headersCache HeadersCache,
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
	if check.IfNil(slashingCache) {
		return nil, process.ErrNilRoundDetectorCache
	}
	if check.IfNil(headersCache) {
		return nil, process.ErrNilRoundHeadersCache
	}

	baseDetector := baseSlashingDetector{roundHandler: roundHandler}

	return &signingSlashingDetector{
		slashingCache:        slashingCache,
		headersCache:         headersCache,
		nodesCoordinator:     nodesCoordinator,
		baseSlashingDetector: baseDetector,
		hasher:               hasher,
		marshaller:           marshaller,
	}, nil
}

// VerifyData - checks if an intercepted data represents a slashable event
func (ssd *signingSlashingDetector) VerifyData(interceptedData process.InterceptedData) (slash.SlashingProofHandler, error) {
	interceptedHeader, castOk := interceptedData.(*interceptedBlocks.InterceptedHeader)
	if !castOk {
		return nil, process.ErrCannotCastInterceptedDataToHeader
	}

	header := interceptedHeader.HeaderHandler()
	round := header.GetRound()
	if !ssd.isRoundRelevant(round) {
		return nil, process.ErrHeaderRoundNotRelevant
	}

	headerHash, err := ssd.computeHashWithoutSignatures(header)
	if err != nil {
		return nil, err
	}

	if ssd.headersCache.Contains(round, headerHash) {
		return nil, process.ErrHeadersNotDifferentHashes
	}

	err = ssd.cacheSigners(interceptedHeader)
	if err != nil {
		return nil, err
	}
	ssd.headersCache.Add(round, headerHash, header)

	slashingResult := ssd.getSlashingResult(round)
	if len(slashingResult) != 0 {
		return slash.NewMultipleSigningProof(slashingResult)
	}

	return nil, process.ErrNoSlashingEventDetected
}

func (ssd *signingSlashingDetector) computeHashWithoutSignatures(header data.HeaderHandler) ([]byte, error) {
	headerCopy := header.Clone()
	headerCopy.SetPubKeysBitmap(nil)
	headerCopy.SetSignature(nil)
	headerCopy.SetLeaderSignature(nil)

	return core.CalculateHash(ssd.marshaller, ssd.hasher, headerCopy)
}

func (ssd *signingSlashingDetector) cacheSigners(interceptedHeader *interceptedBlocks.InterceptedHeader) error {
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
		if slash.IsIndexSetInBitmap(uint32(idx), bitmap) {
			ssd.slashingCache.Add(header.GetRound(), validator.PubKey(), interceptedHeader)
		}
	}

	return nil
}

func (ssd *signingSlashingDetector) getSlashingResult(round uint64) map[string]slash.SlashingResult {
	slashingData := make(map[string]slash.SlashingResult)

	for _, validator := range ssd.slashingCache.GetPubKeys(round) {
		signedHeaders := ssd.slashingCache.GetData(round, validator)
		if len(signedHeaders) > 1 {
			slashingData[string(validator)] = slash.SlashingResult{
				SlashingLevel: computeSlashLevel(signedHeaders),
				Data:          signedHeaders,
			}
		}
	}

	return slashingData
}

// ValidateProof - validates the given proof
func (ssd *signingSlashingDetector) ValidateProof(proof slash.SlashingProofHandler) error {
	multipleSigningProof, castOk := proof.(slash.MultipleSigningProofHandler)
	if !castOk {
		return process.ErrCannotCastProofToMultipleSignedHeaders
	}
	if multipleSigningProof.GetType() != slash.MultipleSigning {
		return process.ErrInvalidSlashType
	}

	signers := multipleSigningProof.GetPubKeys()
	for _, signer := range signers {
		err := checkSlashLevel(multipleSigningProof.GetHeaders(signer), multipleSigningProof.GetLevel(signer))
		if err != nil {
			return err
		}

		err = ssd.checkSignedHeaders(signer, multipleSigningProof.GetHeaders(signer))
		if err != nil {
			return err
		}
	}

	return nil
}

func (ssd *signingSlashingDetector) checkSignedHeaders(pubKey []byte, headers []*interceptedBlocks.InterceptedHeader) error {
	if len(headers) < minSlashableNoOfHeaders {
		return process.ErrNotEnoughHeadersProvided
	}

	hashes := make(map[string]struct{})
	round := headers[0].HeaderHandler().GetRound()
	for _, header := range headers {
		if header.HeaderHandler().GetRound() != round {
			return process.ErrHeadersNotSameRound
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

func (ssd *signingSlashingDetector) checkHash(header data.HeaderHandler, hashes map[string]struct{}) (string, error) {
	hash, err := ssd.computeHashWithoutSignatures(header)
	if err != nil {
		return "", err
	}

	if _, exists := hashes[string(hash)]; exists {
		return "", process.ErrHeadersNotDifferentHashes
	}

	return string(hash), nil
}

func (ssd *signingSlashingDetector) signedHeader(pubKey []byte, header data.HeaderHandler) bool {
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
			slash.IsIndexSetInBitmap(uint32(idx), header.GetPubKeysBitmap()) {
			return true
		}
	}

	return false
}
