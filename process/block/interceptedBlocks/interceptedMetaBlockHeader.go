package interceptedBlocks

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.HdrValidatorHandler = (*InterceptedMetaHeader)(nil)
var _ process.InterceptedData = (*InterceptedMetaHeader)(nil)

// InterceptedMetaHeader represents the wrapper over the meta block header struct
type InterceptedMetaHeader struct {
	hdr               *block.MetaBlock
	sigVerifier       process.InterceptedHeaderSigVerifier
	integrityVerifier process.HeaderIntegrityVerifier
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	hash              []byte
	validityAttester  process.ValidityAttester
	epochStartTrigger process.EpochStartTriggerHandler
}

// NewInterceptedMetaHeader creates a new instance of InterceptedMetaHeader struct
func NewInterceptedMetaHeader(arg *ArgInterceptedBlockHeader) (*InterceptedMetaHeader, error) {
	err := checkBlockHeaderArgument(arg)
	if err != nil {
		return nil, err
	}

	hdr, err := createMetaHdr(arg.Marshalizer, arg.HdrBuff)
	if err != nil {
		return nil, err
	}

	inHdr := &InterceptedMetaHeader{
		hdr:               hdr,
		hasher:            arg.Hasher,
		sigVerifier:       arg.HeaderSigVerifier,
		integrityVerifier: arg.HeaderIntegrityVerifier,
		shardCoordinator:  arg.ShardCoordinator,
		validityAttester:  arg.ValidityAttester,
		epochStartTrigger: arg.EpochStartTrigger,
	}
	inHdr.processFields(arg.HdrBuff)

	return inHdr, nil
}

func createMetaHdr(marshalizer marshal.Marshalizer, hdrBuff []byte) (*block.MetaBlock, error) {
	hdr := &block.MetaBlock{
		ShardInfo: make([]block.ShardData, 0),
	}
	err := marshalizer.Unmarshal(hdr, hdrBuff)
	if err != nil {
		return nil, err
	}

	return hdr, nil
}

func (imh *InterceptedMetaHeader) processFields(txBuff []byte) {
	imh.hash = imh.hasher.Compute(string(txBuff))
}

// Hash gets the hash of this header
func (imh *InterceptedMetaHeader) Hash() []byte {
	return imh.hash
}

// HeaderHandler returns the MetaBlock pointer that holds the data
func (imh *InterceptedMetaHeader) HeaderHandler() data.HeaderHandler {
	return imh.hdr
}

// CheckValidity checks if the received meta header is valid (not nil fields, valid sig and so on)
func (imh *InterceptedMetaHeader) CheckValidity() error {
	err := imh.integrity()
	if err != nil {
		return err
	}

	if !imh.validityAttester.CheckBlockAgainstWhitelist(imh) {
		err = imh.validityAttester.CheckBlockAgainstFinal(imh.HeaderHandler())
		if err != nil {
			return err
		}
	}

	err = imh.validityAttester.CheckBlockAgainstRounder(imh.HeaderHandler())
	if err != nil {
		return err
	}

	err = imh.sigVerifier.VerifyRandSeedAndLeaderSignature(imh.hdr)
	if err != nil {
		return err
	}

	err = imh.sigVerifier.VerifySignature(imh.hdr)
	if err != nil {
		return err
	}

	return imh.integrityVerifier.Verify(imh.hdr)
}

// integrity checks the integrity of the meta header block wrapper
func (imh *InterceptedMetaHeader) integrity() error {
	err := checkHeaderHandler(imh.HeaderHandler())
	if err != nil {
		return err
	}

	err = checkMetaShardInfo(imh.hdr.ShardInfo, imh.shardCoordinator)
	if err != nil {
		return err
	}

	return nil
}

// IsForCurrentShard always returns true
func (imh *InterceptedMetaHeader) IsForCurrentShard() bool {
	return true
}

// Type returns the type of this intercepted data
func (imh *InterceptedMetaHeader) Type() string {
	return "intercepted meta header"
}

// String returns the meta header's most important fields as string
func (imh *InterceptedMetaHeader) String() string {
	return fmt.Sprintf("epoch=%d, round=%d, nonce=%d",
		imh.hdr.Epoch,
		imh.hdr.Round,
		imh.hdr.Nonce,
	)
}

// Identifiers returns the identifiers used in requests
func (imh *InterceptedMetaHeader) Identifiers() [][]byte {
	keyNonce := []byte(fmt.Sprintf("%d-%d", core.MetachainShardId, imh.hdr.Nonce))
	keyEpoch := []byte(core.EpochStartIdentifier(imh.hdr.Epoch))

	return [][]byte{imh.hash, keyNonce, keyEpoch}
}

// IsInterfaceNil returns true if there is no value under the interface
func (imh *InterceptedMetaHeader) IsInterfaceNil() bool {
	return imh == nil
}
