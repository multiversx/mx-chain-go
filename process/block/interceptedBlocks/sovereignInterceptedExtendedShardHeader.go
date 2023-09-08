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
)

type ArgsSovereignInterceptedHeader struct {
	Marshaller  marshal.Marshalizer
	Hasher      hashing.Hasher
	HeaderBytes []byte
}

type SovereignInterceptedHeader struct {
	hdr    data.ShardHeaderExtendedHandler
	hasher hashing.Hasher
	hash   []byte
}

// NewSovereignInterceptedHeader creates a new instance sovereign extended header interceptor
func NewSovereignInterceptedHeader(args ArgsSovereignInterceptedHeader) (*SovereignInterceptedHeader, error) {
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}

	hdr, err := unmarshalExtendedShardHeader(args.Marshaller, args.HeaderBytes)
	if err != nil {
		return nil, err
	}

	inHdr := &SovereignInterceptedHeader{
		hdr:    hdr,
		hasher: args.Hasher,
	}

	inHdr.hash = inHdr.hasher.Compute(string(args.HeaderBytes))

	return inHdr, nil
}

func unmarshalExtendedShardHeader(marshaller marshal.Marshalizer, headerBytes []byte) (data.ShardHeaderExtendedHandler, error) {
	extendedHeader := &block.ShardHeaderExtended{}
	err := marshaller.Unmarshal(extendedHeader, headerBytes)
	if err != nil {
		return nil, err
	}

	if check.IfNil(extendedHeader.Header) {
		return nil, fmt.Errorf("ErrNilHeaderHandler while checking inner header")
	}

	return extendedHeader, nil
}

// CheckValidity checks if the received header is valid (not nil fields, valid sig and so on)
func (inHdr *SovereignInterceptedHeader) CheckValidity() error {
	return nil
}

// Hash gets the hash of this header
func (inHdr *SovereignInterceptedHeader) Hash() []byte {
	return inHdr.hash
}

// GetExtendedHeader returns intercepted extended shard header
func (inHdr *SovereignInterceptedHeader) GetExtendedHeader() data.ShardHeaderExtendedHandler {
	return inHdr.hdr
}

// IsForCurrentShard returns true if this header is meant to be processed by the node from this shard
func (inHdr *SovereignInterceptedHeader) IsForCurrentShard() bool {
	return true
}

// Type returns the type of this intercepted data
func (inHdr *SovereignInterceptedHeader) Type() string {
	return "intercepted sovereign extended shard header"
}

// String returns the header's most important fields as string
func (inHdr *SovereignInterceptedHeader) String() string {
	return fmt.Sprintf("shardId=%d, shardEpoch=%d, round=%d, nonce=%d",
		inHdr.hdr.GetShardID(),
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
