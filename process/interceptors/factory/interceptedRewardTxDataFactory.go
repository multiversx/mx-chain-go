package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type interceptedRewardTxDataFactory struct {
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	addrConverter    state.AddressConverter
	shardCoordinator sharding.Coordinator
}

// NewInterceptedRewardTxDataFactory creates an instance of interceptedRewardTxDataFactory
func NewInterceptedRewardTxDataFactory(argument *ArgInterceptedDataFactory) (*interceptedRewardTxDataFactory, error) {
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

	return &interceptedRewardTxDataFactory{
		marshalizer:      argument.Marshalizer,
		hasher:           argument.Hasher,
		addrConverter:    argument.AddrConv,
		shardCoordinator: argument.ShardCoordinator,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (irtdf *interceptedRewardTxDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	return rewardTransaction.NewInterceptedRewardTransaction(
		buff,
		irtdf.marshalizer,
		irtdf.hasher,
		irtdf.addrConverter,
		irtdf.shardCoordinator,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (irtdf *interceptedRewardTxDataFactory) IsInterfaceNil() bool {
	if irtdf == nil {
		return true
	}
	return false
}
