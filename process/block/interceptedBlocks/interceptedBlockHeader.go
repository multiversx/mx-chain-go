package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedHeader represents the wrapper over HeaderWrapper struct.
// It implements Newer and Hashed interfaces
type InterceptedHeader struct {
	hdr               *block.Header
	sigVerifier       process.InterceptedHeaderSigVerifier
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	hash              []byte
	isForCurrentShard bool
	chainID           []byte
}

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedHeader(arg *ArgInterceptedBlockHeader) (*InterceptedHeader, error) {
	err := checkBlockHeaderArgument(arg)
	if err != nil {
		return nil, err
	}

	hdr, err := createShardHdr(arg.Marshalizer, arg.HdrBuff)
	if err != nil {
		return nil, err
	}

	inHdr := &InterceptedHeader{
		hdr:              hdr,
		hasher:           arg.Hasher,
		sigVerifier:      arg.HeaderSigVerifier,
		shardCoordinator: arg.ShardCoordinator,
		chainID:          arg.ChainID,
	}
	inHdr.processFields(arg.HdrBuff)

	return inHdr, nil
}

func createShardHdr(marshalizer marshal.Marshalizer, hdrBuff []byte) (*block.Header, error) {
	hdr := &block.Header{
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
		MetaBlockHashes:  make([][]byte, 0),
	}
	err := marshalizer.Unmarshal(hdr, hdrBuff)
	if err != nil {
		return nil, err
	}

	return hdr, nil
}

func (inHdr *InterceptedHeader) processFields(txBuff []byte) {
	inHdr.hash = inHdr.hasher.Compute(string(txBuff))

	isHeaderForCurrentShard := inHdr.shardCoordinator.SelfId() == inHdr.HeaderHandler().GetShardID()
	isMetachainShardCoordinator := inHdr.shardCoordinator.SelfId() == sharding.MetachainShardId
	inHdr.isForCurrentShard = isHeaderForCurrentShard || isMetachainShardCoordinator
}

// CheckValidity checks if the received header is valid (not nil fields, valid sig and so on)
func (inHdr *InterceptedHeader) CheckValidity() error {
	err := inHdr.integrity()
	if err != nil {
		return err
	}

	err = inHdr.sigVerifier.VerifyRandSeedAndLeaderSignature(inHdr.hdr)
	if err != nil {
		return err
	}

	err = inHdr.sigVerifier.VerifySignature(inHdr.hdr)
	if err != nil {
		return err
	}

	return inHdr.hdr.CheckChainID(inHdr.chainID)
}

// integrity checks the integrity of the header block wrapper
func (inHdr *InterceptedHeader) integrity() error {
	err := checkHeaderHandler(inHdr.HeaderHandler())
	if err != nil {
		return err
	}

	err = checkMiniblocks(inHdr.hdr.MiniBlockHeaders, inHdr.shardCoordinator)
	if err != nil {
		return err
	}

	return nil
}

// Hash gets the hash of this header
func (inHdr *InterceptedHeader) Hash() []byte {
	return inHdr.hash
}

// HeaderHandler returns the HeaderHandler pointer that holds the data
func (inHdr *InterceptedHeader) HeaderHandler() data.HeaderHandler {
	return inHdr.hdr
}

// IsForCurrentShard returns true if this header is meant to be processed by the node from this shard
func (inHdr *InterceptedHeader) IsForCurrentShard() bool {
	return inHdr.isForCurrentShard
}

// IsInterfaceNil returns true if there is no value under the interface
func (inHdr *InterceptedHeader) IsInterfaceNil() bool {
	if inHdr == nil {
		return true
	}
	return false
}
