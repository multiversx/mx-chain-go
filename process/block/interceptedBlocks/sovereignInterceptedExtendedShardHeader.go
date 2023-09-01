package interceptedBlocks

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// InterceptedHeader represents the wrapper over HeaderWrapper struct.
// It implements Newer and Hashed interfaces
type SovereignInterceptedHeader struct {
	hdr               data.HeaderHandler
	sigVerifier       process.InterceptedHeaderSigVerifier
	integrityVerifier process.HeaderIntegrityVerifier
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	hash              []byte
	isForCurrentShard bool
	validityAttester  process.ValidityAttester
	epochStartTrigger process.EpochStartTriggerHandler
}

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewSovereignInterceptedHeader(arg *ArgInterceptedBlockHeader) (*InterceptedHeader, error) {
	err := checkBlockHeaderArgument(arg)
	if err != nil {
		return nil, err
	}

	hdr, err := UnmarshalExtendedShardHeader(arg.Marshalizer, arg.HdrBuff)
	if err != nil {
		return nil, err
	}

	inHdr := &InterceptedHeader{
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

func UnmarshalExtendedShardHeader(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.ShardHeaderHandler, error) {
	hdrV2 := &block.ShardHeaderExtended{}
	err := marshalizer.Unmarshal(hdrV2, hdrBuff)
	if err != nil {
		return nil, err
	}
	if check.IfNil(hdrV2.Header) {
		return nil, fmt.Errorf("ErrNilHeaderHandler while checking inner header")
	}

	return hdrV2, nil
}

func (inHdr *SovereignInterceptedHeader) processFields(txBuff []byte) {
	inHdr.hash = inHdr.hasher.Compute(string(txBuff))

	inHdr.isForCurrentShard = true
}

// CheckValidity checks if the received header is valid (not nil fields, valid sig and so on)
func (inHdr *SovereignInterceptedHeader) CheckValidity() error {
	return nil
}

// Hash gets the hash of this header
func (inHdr *SovereignInterceptedHeader) Hash() []byte {
	return inHdr.hash
}

// IsForCurrentShard returns true if this header is meant to be processed by the node from this shard
func (inHdr *SovereignInterceptedHeader) IsForCurrentShard() bool {
	return inHdr.isForCurrentShard
}

// Type returns the type of this intercepted data
func (inHdr *SovereignInterceptedHeader) Type() string {
	return "intercepted sovereign extended shard header"
}

// String returns the header's most important fields as string
func (inHdr *SovereignInterceptedHeader) String() string {
	return fmt.Sprintf("shardId=%d, metaEpoch=%d, shardEpoch=%d, round=%d, nonce=%d",
		inHdr.hdr.GetShardID(),
		inHdr.epochStartTrigger.Epoch(),
		inHdr.hdr.GetEpoch(),
		inHdr.hdr.GetRound(),
		inHdr.hdr.GetNonce(),
	)
}

// Identifiers returns the identifiers used in requests
func (inHdr *SovereignInterceptedHeader) Identifiers() [][]byte {
	keyNonce := []byte(fmt.Sprintf("%d-%d", inHdr.hdr.GetShardID(), inHdr.hdr.GetNonce()))
	keyEpoch := []byte(core.EpochStartIdentifier(inHdr.hdr.GetEpoch()))

	return [][]byte{inHdr.hash, keyNonce, keyEpoch}
}

// IsInterfaceNil returns true if there is no value under the interface
func (inHdr *SovereignInterceptedHeader) IsInterfaceNil() bool {
	return inHdr == nil
}
