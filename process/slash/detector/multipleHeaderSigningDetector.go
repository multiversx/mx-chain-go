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
func NewMultipleHeaderSigningDetector(
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

	return &multipleHeaderSigningDetector{
		slashingCache:        slashingCache,
		headersCache:         headersCache,
		nodesCoordinator:     nodesCoordinator,
		baseSlashingDetector: baseDetector,
		hasher:               hasher,
		marshaller:           marshaller,
	}, nil
}

// VerifyData - checks if an intercepted data represents a slashable event
func (mhs *multipleHeaderSigningDetector) VerifyData(interceptedData process.InterceptedData) (slash.SlashingProofHandler, error) {
	interceptedHeader, castOk := interceptedData.(*interceptedBlocks.InterceptedHeader)
	if !castOk {
		return nil, process.ErrCannotCastInterceptedDataToHeader
	}

	header := interceptedHeader.HeaderHandler()
	round := header.GetRound()
	if !mhs.isRoundRelevant(round) {
		return nil, process.ErrHeaderRoundNotRelevant
	}

	headerHash, err := mhs.computeHashWithoutSignatures(header)
	if err != nil {
		return nil, err
	}

	if mhs.headersCache.Contains(round, headerHash) {
		return nil, process.ErrHeadersNotDifferentHashes
	}

	err = mhs.cacheSigners(interceptedHeader)
	if err != nil {
		return nil, err
	}
	mhs.headersCache.Add(round, headerHash, header)

	slashingResult := mhs.getSlashingResult(round)
	if len(slashingResult) != 0 {
		return slash.NewMultipleSigningProof(slashingResult)
	}

	return nil, process.ErrNoSlashingEventDetected
}

func (mhs *multipleHeaderSigningDetector) computeHashWithoutSignatures(header data.HeaderHandler) ([]byte, error) {
	headerCopy := header.Clone()
	headerCopy.SetPubKeysBitmap(nil)
	headerCopy.SetSignature(nil)
	headerCopy.SetLeaderSignature(nil)

	return core.CalculateHash(mhs.marshaller, mhs.hasher, headerCopy)
}

func (mhs *multipleHeaderSigningDetector) cacheSigners(interceptedHeader *interceptedBlocks.InterceptedHeader) error {
	header := interceptedHeader.HeaderHandler()
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
			mhs.slashingCache.Add(header.GetRound(), validator.PubKey(), interceptedHeader)
		}
	}

	return nil
}

func (mhs *multipleHeaderSigningDetector) getSlashingResult(round uint64) map[string]slash.SlashingResult {
	slashingData := make(map[string]slash.SlashingResult)

	for _, validator := range mhs.slashingCache.GetPubKeys(round) {
		signedHeaders := mhs.slashingCache.GetData(round, validator)
		if len(signedHeaders) > 1 {
			slashingData[string(validator)] = slash.SlashingResult{
				SlashingLevel: mhs.computeSlashLevel(signedHeaders),
				Data:          signedHeaders,
			}
		}
	}

	return slashingData
}

// TODO: Add different logic here once slashing threat levels are clearly defined
func (mhs *multipleHeaderSigningDetector) computeSlashLevel(data []process.InterceptedData) slash.ThreatLevel {
	return computeSlashLevelBasedOnHeadersCount(data)
}

// ValidateProof - validates the given proof
func (mhs *multipleHeaderSigningDetector) ValidateProof(proof slash.SlashingProofHandler) error {
	multipleSigningProof, castOk := proof.(slash.MultipleSigningProofHandler)
	if !castOk {
		return process.ErrCannotCastProofToMultipleSignedHeaders
	}
	if multipleSigningProof.GetType() != slash.MultipleSigning {
		return process.ErrInvalidSlashType
	}

	signers := multipleSigningProof.GetPubKeys()
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
func (mhs *multipleHeaderSigningDetector) checkSlashLevel(headers []*interceptedBlocks.InterceptedHeader, level slash.ThreatLevel) error {
	return checkSlashLevelBasedOnHeadersCount(headers, level)
}

func (mhs *multipleHeaderSigningDetector) checkSignedHeaders(pubKey []byte, headers []*interceptedBlocks.InterceptedHeader) error {
	if len(headers) < minSlashableNoOfHeaders {
		return process.ErrNotEnoughHeadersProvided
	}

	hashes := make(map[string]struct{})
	round := headers[0].HeaderHandler().GetRound()
	for _, header := range headers {
		if header.HeaderHandler().GetRound() != round {
			return process.ErrHeadersNotSameRound
		}

		hash, err := mhs.checkHash(header.HeaderHandler(), hashes)
		if err != nil {
			return err
		}

		if !mhs.signedHeader(pubKey, header.HeaderHandler()) {
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
