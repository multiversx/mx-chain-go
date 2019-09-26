package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
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
	hdr                 *block.Header
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	multiSigVerifier    crypto.MultiSigVerifier
	chronologyValidator process.ChronologyValidator
	shardCoordinator    sharding.Coordinator
	hash                []byte
	isForCurrentShard   bool
}

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedHeader(arg *ArgInterceptedBlockHeader) (*InterceptedHeader, error) {
	err := checkArgument(arg)
	if err != nil {
		return nil, err
	}

	hdr := &block.Header{
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
		MetaBlockHashes:  make([][]byte, 0),
	}
	err = arg.Marshalizer.Unmarshal(hdr, arg.HdrBuff)
	if err != nil {
		return nil, err
	}

	inHdr := &InterceptedHeader{
		hdr:                 hdr,
		marshalizer:         arg.Marshalizer,
		hasher:              arg.Hasher,
		multiSigVerifier:    arg.MultiSigVerifier,
		chronologyValidator: arg.ChronologyValidator,
		shardCoordinator:    arg.ShardCoordinator,
	}
	inHdr.processFields(arg.HdrBuff)

	return inHdr, nil
}

func (inHdr *InterceptedHeader) processFields(txBuff []byte) {
	inHdr.hash = inHdr.hasher.Compute(string(txBuff))

	isHeaderForCurrentShard := inHdr.shardCoordinator.SelfId() == inHdr.HeaderHandler().GetShardID()
	isMetachainShardCoordinator := inHdr.shardCoordinator.SelfId() == sharding.MetachainShardId
	inHdr.isForCurrentShard = isHeaderForCurrentShard || isMetachainShardCoordinator
}

// CheckValidity checks if the received transaction is valid (not nil fields, valid sig and so on)
func (inHdr *InterceptedHeader) CheckValidity() error {
	err := inHdr.integrity()
	if err != nil {
		return err
	}

	return inHdr.verifySig()
}

// integrity checks the integrity of the state block wrapper
func (inHdr *InterceptedHeader) integrity() error {
	err := checkHeaderHandler(inHdr.HeaderHandler())
	if err != nil {
		return err
	}

	err = checkMiniblocks(inHdr.hdr.MiniBlockHeaders, inHdr.shardCoordinator)
	if err != nil {
		return err
	}

	return inHdr.chronologyValidator.ValidateReceivedBlock(
		inHdr.hdr.ShardId,
		inHdr.hdr.Epoch,
		inHdr.hdr.Nonce,
		inHdr.hdr.Round,
	)
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

// verifySig verifies a signature
func (inHdr *InterceptedHeader) verifySig() error {
	// TODO: Check block signature after multisig will be implemented
	// TODO: the interceptors do not have access yet to consensus group selection to validate multisigs
	// TODO: verify that the block proposer is among the signers and in the bitmap
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (inHdr *InterceptedHeader) IsInterfaceNil() bool {
	if inHdr == nil {
		return true
	}
	return false
}
