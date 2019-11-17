package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedMetaHeader represents the wrapper over the meta block header struct
type InterceptedMetaHeader struct {
	hdr              *block.MetaBlock
	sigVerifier      *headerSigVerifier
	hasher           hashing.Hasher
	shardCoordinator sharding.Coordinator
	hash             []byte
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

	sigVerifier := &headerSigVerifier{
		marshalizer:       arg.Marshalizer,
		hasher:            arg.Hasher,
		nodesCoordinator:  arg.NodesCoordinator,
		singleSigVerifier: arg.SingleSigVerifier,
		multiSigVerifier:  arg.MultiSigVerifier,
		keyGen:            arg.KeyGen,
	}

	inHdr := &InterceptedMetaHeader{
		hdr:              hdr,
		hasher:           arg.Hasher,
		sigVerifier:      sigVerifier,
		shardCoordinator: arg.ShardCoordinator,
	}
	//wire-up the "virtual" function
	inHdr.sigVerifier.copyHeaderWithoutSig = inHdr.copyHeaderWithoutSig
	inHdr.processFields(arg.HdrBuff)

	return inHdr, nil
}

func createMetaHdr(marshalizer marshal.Marshalizer, hdrBuff []byte) (*block.MetaBlock, error) {
	hdr := &block.MetaBlock{
		ShardInfo: make([]block.ShardData, 0),
		PeerInfo:  make([]block.PeerData, 0),
	}
	err := marshalizer.Unmarshal(hdr, hdrBuff)
	if err != nil {
		return nil, err
	}

	return hdr, nil
}

func (imh *InterceptedMetaHeader) copyHeaderWithoutSig(header data.HeaderHandler) data.HeaderHandler {
	//it is virtually impossible here to have a wrong type assertion case
	hdr := header.(*block.MetaBlock)

	headerCopy := *hdr
	headerCopy.Signature = nil
	headerCopy.PubKeysBitmap = nil

	return &headerCopy
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

	err = imh.sigVerifier.verifyRandSeed(imh.hdr)
	if err != nil {
		return err
	}

	return imh.sigVerifier.verifySig(imh.hdr)
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

// IsInterfaceNil returns true if there is no value under the interface
func (imh *InterceptedMetaHeader) IsInterfaceNil() bool {
	if imh == nil {
		return true
	}
	return false
}
