package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedHeader represents the wrapper over HeaderWrapper struct.
// It implements Newer and Hashed interfaces
type InterceptedHeader struct {
	hdr               *block.Header
	sigVerifier       *headerSigVerifier
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	hash              []byte
	isForCurrentShard bool
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

	sigVerifier := &headerSigVerifier{
		marshalizer:       arg.Marshalizer,
		hasher:            arg.Hasher,
		nodesCoordinator:  arg.NodesCoordinator,
		multiSigVerifier:  arg.MultiSigVerifier,
		singleSigVerifier: arg.SingleSigVerifier,
		keyGen:            arg.KeyGen,
	}

	inHdr := &InterceptedHeader{
		hdr:              hdr,
		hasher:           arg.Hasher,
		sigVerifier:      sigVerifier,
		shardCoordinator: arg.ShardCoordinator,
	}
	//wire-up the "virtual" function
	inHdr.sigVerifier.copyHeaderWithoutSig = inHdr.copyHeaderWithoutSig
	inHdr.sigVerifier.copyHeaderWithoutLeaderSig = inHdr.copyHeaderWithoutLeaderSig
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

func (inHdr *InterceptedHeader) copyHeaderWithoutSig(header data.HeaderHandler) data.HeaderHandler {
	//it is virtually impossible here to have a wrong type assertion case
	hdr := header.(*block.Header)

	headerCopy := *hdr
	headerCopy.Signature = nil
	headerCopy.PubKeysBitmap = nil
	headerCopy.LeaderSignature = nil

	return &headerCopy
}

func (inHdr *InterceptedHeader) copyHeaderWithoutLeaderSig(header data.HeaderHandler) data.HeaderHandler {
	//it is virtually impossible here to have a wrong type assertion case
	hdr := header.(*block.Header)

	headerCopy := *hdr
	headerCopy.LeaderSignature = nil

	return &headerCopy
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

	err = inHdr.sigVerifier.verifyRandSeedAndLeaderSignature(inHdr.hdr)
	if err != nil {
		return err
	}

	return inHdr.sigVerifier.verifySig(inHdr.hdr)
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
