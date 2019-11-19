package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type interceptedMetaHeaderDataFactory struct {
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	singleSigVerifier crypto.SingleSigner
	multiSigVerifier  crypto.MultiSigVerifier
	nodesCoordinator  sharding.NodesCoordinator
	keyGen            crypto.KeyGenerator
}

// NewInterceptedMetaHeaderDataFactory creates an instance of interceptedMetaHeaderDataFactory
func NewInterceptedMetaHeaderDataFactory(argument *ArgInterceptedDataFactory) (*interceptedMetaHeaderDataFactory, error) {
	if argument == nil {
		return nil, process.ErrNilArguments
	}
	if check.IfNil(argument.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(argument.MultiSigVerifier) {
		return nil, process.ErrNilMultiSigVerifier
	}
	if check.IfNil(argument.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(argument.BlockKeyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(argument.BlockSigner) {
		return nil, process.ErrNilSingleSigner
	}

	return &interceptedMetaHeaderDataFactory{
		marshalizer:       argument.Marshalizer,
		hasher:            argument.Hasher,
		shardCoordinator:  argument.ShardCoordinator,
		multiSigVerifier:  argument.MultiSigVerifier,
		nodesCoordinator:  argument.NodesCoordinator,
		keyGen:            argument.BlockKeyGen,
		singleSigVerifier: argument.BlockSigner,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (imhdf *interceptedMetaHeaderDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := &interceptedBlocks.ArgInterceptedBlockHeader{
		HdrBuff:           buff,
		Marshalizer:       imhdf.marshalizer,
		Hasher:            imhdf.hasher,
		SingleSigVerifier: imhdf.singleSigVerifier,
		MultiSigVerifier:  imhdf.multiSigVerifier,
		NodesCoordinator:  imhdf.nodesCoordinator,
		ShardCoordinator:  imhdf.shardCoordinator,
		KeyGen:            imhdf.keyGen,
	}

	return interceptedBlocks.NewInterceptedMetaHeader(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (imhdf *interceptedMetaHeaderDataFactory) IsInterfaceNil() bool {
	if imhdf == nil {
		return true
	}
	return false
}
