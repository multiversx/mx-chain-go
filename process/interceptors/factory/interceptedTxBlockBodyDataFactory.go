package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type interceptedTxBlockBodyDataFactory struct {
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	shardCoordinator sharding.Coordinator
}

// NewInterceptedTxBlockBodyDataFactory creates an instance of interceptedTxBlockBodyDataFactory
func NewInterceptedTxBlockBodyDataFactory(argument *ArgInterceptedDataFactory) (*interceptedTxBlockBodyDataFactory, error) {
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

	return &interceptedTxBlockBodyDataFactory{
		marshalizer:      argument.Marshalizer,
		hasher:           argument.Hasher,
		shardCoordinator: argument.ShardCoordinator,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (itbbdf *interceptedTxBlockBodyDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := &interceptedBlocks.ArgInterceptedTxBlockBody{
		TxBlockBodyBuff:  buff,
		Marshalizer:      itbbdf.marshalizer,
		Hasher:           itbbdf.hasher,
		ShardCoordinator: itbbdf.shardCoordinator,
	}

	return interceptedBlocks.NewInterceptedTxBlockBody(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (itbbdf *interceptedTxBlockBodyDataFactory) IsInterfaceNil() bool {
	if itbbdf == nil {
		return true
	}
	return false
}
