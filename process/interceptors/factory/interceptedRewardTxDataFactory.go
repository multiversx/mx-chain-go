package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/rewardTransaction"
	"github.com/multiversx/mx-chain-go/sharding"
)

var _ process.InterceptedDataFactory = (*interceptedRewardTxDataFactory)(nil)

type interceptedRewardTxDataFactory struct {
	protoMarshalizer marshal.Marshalizer
	hasher           hashing.Hasher
	pubkeyConverter  core.PubkeyConverter
	shardCoordinator sharding.Coordinator
}

// NewInterceptedRewardTxDataFactory creates an instance of interceptedRewardTxDataFactory
func NewInterceptedRewardTxDataFactory(argument *ArgInterceptedDataFactory) (*interceptedRewardTxDataFactory, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.CoreComponents) {
		return nil, process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(argument.CoreComponents.InternalMarshalizer()) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.CoreComponents.TxMarshalizer()) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.CoreComponents.Hasher()) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.CoreComponents.AddressPubKeyConverter()) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &interceptedRewardTxDataFactory{
		protoMarshalizer: argument.CoreComponents.InternalMarshalizer(),
		hasher:           argument.CoreComponents.Hasher(),
		pubkeyConverter:  argument.CoreComponents.AddressPubKeyConverter(),
		shardCoordinator: argument.ShardCoordinator,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (irtdf *interceptedRewardTxDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	return rewardTransaction.NewInterceptedRewardTransaction(
		buff,
		irtdf.protoMarshalizer,
		irtdf.hasher,
		irtdf.pubkeyConverter,
		irtdf.shardCoordinator,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (irtdf *interceptedRewardTxDataFactory) IsInterfaceNil() bool {
	return irtdf == nil
}
