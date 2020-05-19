package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.InterceptedDataFactory = (*interceptedMiniblockDataFactory)(nil)

type interceptedMiniblockDataFactory struct {
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	shardCoordinator sharding.Coordinator
}

// NewInterceptedMiniblockDataFactory creates an instance of interceptedMiniblockDataFactory
func NewInterceptedMiniblockDataFactory(argument *ArgInterceptedDataFactory) (*interceptedMiniblockDataFactory, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.CoreComponents) {
		return nil, process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(argument.CoreComponents.InternalMarshalizer()) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.CoreComponents.Hasher()) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &interceptedMiniblockDataFactory{
		marshalizer:      argument.CoreComponents.InternalMarshalizer(),
		hasher:           argument.CoreComponents.Hasher(),
		shardCoordinator: argument.ShardCoordinator,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (imfd *interceptedMiniblockDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := &interceptedBlocks.ArgInterceptedMiniblock{
		MiniblockBuff:    buff,
		Marshalizer:      imfd.marshalizer,
		Hasher:           imfd.hasher,
		ShardCoordinator: imfd.shardCoordinator,
	}

	return interceptedBlocks.NewInterceptedMiniblock(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (imfd *interceptedMiniblockDataFactory) IsInterfaceNil() bool {
	return imfd == nil
}
