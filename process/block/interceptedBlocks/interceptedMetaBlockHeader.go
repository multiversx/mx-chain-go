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

// InterceptedMetaHeader represents the wrapper over the meta block header struct
type InterceptedMetaHeader struct {
	hdr                 *block.MetaBlock
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	multiSigVerifier    crypto.MultiSigVerifier
	chronologyValidator process.ChronologyValidator
	shardCoordinator    sharding.Coordinator
	hash                []byte
}

// NewInterceptedMetaHeader creates a new instance of InterceptedMetaHeader struct
func NewInterceptedMetaHeader(arg *ArgInterceptedBlockHeader) (*InterceptedMetaHeader, error) {
	err := checkArgument(arg)
	if err != nil {
		return nil, err
	}

	hdr := &block.MetaBlock{
		ShardInfo: make([]block.ShardData, 0),
		PeerInfo:  make([]block.PeerData, 0),
	}
	err = arg.Marshalizer.Unmarshal(hdr, arg.HdrBuff)
	if err != nil {
		return nil, err
	}

	inHdr := &InterceptedMetaHeader{
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

// CheckValidity checks if the received transaction is valid (not nil fields, valid sig and so on)
func (imh *InterceptedMetaHeader) CheckValidity() error {
	err := imh.integrity()
	if err != nil {
		return err
	}

	return imh.verifySig()
}

// integrity checks the integrity of the state block wrapper
func (imh *InterceptedMetaHeader) integrity() error {
	err := checkHeaderHandler(imh.HeaderHandler())
	if err != nil {
		return err
	}

	err = checkMetaShardInfo(imh.hdr.ShardInfo, imh.shardCoordinator)
	if err != nil {
		return err
	}

	return imh.chronologyValidator.ValidateReceivedBlock(
		sharding.MetachainShardId,
		imh.hdr.Epoch,
		imh.hdr.Nonce,
		imh.hdr.Round,
	)
}

// verifySig verifies a signature
func (imh *InterceptedMetaHeader) verifySig() error {
	// TODO: Check block signature after multisig will be implemented
	// TODO: the interceptors do not have access yet to consensus group selection to validate multisigs
	// TODO: verify that the block proposer is among the signers
	return nil
}

// IsForCurrentShard always returns true
func (imh *InterceptedMetaHeader) IsForCurrentShard() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (imh *InterceptedMetaHeader) IsInterfaceNil() bool {
	if imh == nil {
		return true
	}
	return false
}
