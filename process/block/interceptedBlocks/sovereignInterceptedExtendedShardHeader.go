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

// ArgsSovereignInterceptedHeader is a struct placeholder for args needed to create a new instance of a sovereign extended header interceptor
type ArgsSovereignInterceptedHeader struct {
	Marshaller  marshal.Marshalizer
	Hasher      hashing.Hasher
	HeaderBytes []byte
}

type sovereignInterceptedHeader struct {
	hdr  data.ShardHeaderExtendedHandler
	hash []byte
}

// NewSovereignInterceptedHeader creates a new instance of a sovereign extended header interceptor
func NewSovereignInterceptedHeader(args ArgsSovereignInterceptedHeader) (*sovereignInterceptedHeader, error) {
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

	return &sovereignInterceptedHeader{
		hdr:  hdr,
		hash: args.Hasher.Compute(string(args.HeaderBytes)),
	}, nil
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

// CheckValidity checks if the received header is valid (basic checks such as not nil fields)
func (inHdr *sovereignInterceptedHeader) CheckValidity() error {
	return nil
}

// Hash returns the hash of received extended header
func (inHdr *sovereignInterceptedHeader) Hash() []byte {
	return inHdr.hash
}

// GetExtendedHeader returns intercepted extended shard header
func (inHdr *sovereignInterceptedHeader) GetExtendedHeader() data.ShardHeaderExtendedHandler {
	return inHdr.hdr
}

// IsForCurrentShard returns true
func (inHdr *sovereignInterceptedHeader) IsForCurrentShard() bool {
	return true
}

// Type returns the type of this intercepted data
func (inHdr *sovereignInterceptedHeader) Type() string {
	return "intercepted sovereign extended shard header"
}

// String returns the header's most important fields as string
func (inHdr *sovereignInterceptedHeader) String() string {
	return fmt.Sprintf("%s, shardId=%d, shardEpoch=%d, round=%d, nonce=%d",
		inHdr.Type(),
		inHdr.hdr.GetShardID(),
		inHdr.hdr.GetEpoch(),
		inHdr.hdr.GetRound(),
		inHdr.hdr.GetNonce(),
	)
}

// Identifiers returns the identifiers used in requests
func (inHdr *sovereignInterceptedHeader) Identifiers() [][]byte {
	keyNonce := []byte(fmt.Sprintf("%d-%d", inHdr.hdr.GetShardID(), inHdr.hdr.GetNonce()))
	keyEpoch := []byte(core.EpochStartIdentifier(inHdr.hdr.GetEpoch()))
	keyType := []byte(inHdr.Type())

	return [][]byte{keyType, inHdr.hash, keyNonce, keyEpoch}
}

// IsInterfaceNil returns true if there is no value under the interface
func (inHdr *sovereignInterceptedHeader) IsInterfaceNil() bool {
	return inHdr == nil
}
