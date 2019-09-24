package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
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
	isForMyShard        bool
}

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedHeader(arg *ArgInterceptedBlockHeader) (*InterceptedHeader, error) {
	if arg == nil {
		return nil, process.ErrNilArguments
	}
	if arg.HdrBuff == nil {
		return nil, process.ErrNilBuffer
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(arg.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(arg.MultiSigVerifier) {
		return nil, process.ErrNilMultiSigVerifier
	}
	if check.IfNil(arg.ChronologyValidator) {
		return nil, process.ErrNilChronologyValidator
	}
	if check.IfNil(arg.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	hdr := &block.Header{
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
		MetaBlockHashes:  make([][]byte, 0),
	}
	err := arg.Marshalizer.Unmarshal(hdr, arg.HdrBuff)
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
	inHdr.isForMyShard = isHeaderForCurrentShard || isMetachainShardCoordinator
}

// CheckValidity checks if the received transaction is valid (not nil fields, valid sig and so on)
func (inHdr *InterceptedHeader) CheckValidity() error {
	err := inHdr.integrity()
	if err != nil {
		return err
	}

	err = inHdr.verifySig()
	if err != nil {
		return err
	}

	return nil
}

// integrity checks the integrity of the state block wrapper
func (inHdr *InterceptedHeader) integrity() error {
	if inHdr.HeaderHandler().GetPubKeysBitmap() == nil {
		return process.ErrNilPubKeysBitmap
	}
	if inHdr.HeaderHandler().GetPrevHash() == nil {
		return process.ErrNilPreviousBlockHash
	}
	if inHdr.HeaderHandler().GetSignature() == nil {
		return process.ErrNilSignature
	}
	if inHdr.HeaderHandler().GetRootHash() == nil {
		return process.ErrNilRootHash
	}
	if inHdr.HeaderHandler().GetRandSeed() == nil {
		return process.ErrNilRandSeed
	}
	if inHdr.HeaderHandler().GetPrevRandSeed() == nil {
		return process.ErrNilPrevRandSeed
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

// Shard returns the shard ID for which this header is addressed
func (inHdr *InterceptedHeader) Shard() uint32 {
	return inHdr.hdr.ShardId
}

// HeaderHandler returns the HeaderHandler pointer that holds the data
func (inHdr *InterceptedHeader) HeaderHandler() data.HeaderHandler {
	return inHdr.hdr
}

// IsForMyShard returns true if this header is meant to be processed by the node from this shard
func (inHdr *InterceptedHeader) IsForMyShard() bool {
	return inHdr.isForMyShard
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
