package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/unsigned"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type interceptedUnsignedTxDataFactory struct {
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	addrConverter    state.AddressConverter
	shardCoordinator sharding.Coordinator
}

// NewInterceptedUnsignedTxDataFactory creates an instance of interceptedUnsignedTxDataFactory
func NewInterceptedUnsignedTxDataFactory(argument *ArgInterceptedDataFactory) (*interceptedUnsignedTxDataFactory, error) {
	if argument == nil {
		return nil, process.ErrNilArguments
	}
	if check.IfNil(argument.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.AddrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &interceptedUnsignedTxDataFactory{
		marshalizer:      argument.Marshalizer,
		hasher:           argument.Hasher,
		addrConverter:    argument.AddrConv,
		shardCoordinator: argument.ShardCoordinator,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (iutdf *interceptedUnsignedTxDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	return unsigned.NewInterceptedUnsignedTransaction(
		buff,
		iutdf.marshalizer,
		iutdf.hasher,
		iutdf.addrConverter,
		iutdf.shardCoordinator,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (iutdf *interceptedUnsignedTxDataFactory) IsInterfaceNil() bool {
	if iutdf == nil {
		return true
	}
	return false
}
