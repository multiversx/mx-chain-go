package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type interceptedValidatorInfoFactory struct {
	marshaller       marshal.Marshalizer
	shardCoordinator sharding.Coordinator
}

// NewInterceptedValidatorInfoFactory creates an instance of interceptedValidatorInfoFactory
func NewInterceptedValidatorInfoFactory(args ArgInterceptedDataFactory) (*interceptedValidatorInfoFactory, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &interceptedValidatorInfoFactory{
		marshaller:       args.CoreComponents.InternalMarshalizer(),
		shardCoordinator: args.ShardCoordinator,
	}, nil
}

func checkArgs(args ArgInterceptedDataFactory) error {
	if check.IfNil(args.CoreComponents) {
		return process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.CoreComponents.InternalMarshalizer()) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}

	return nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (isvif *interceptedValidatorInfoFactory) Create(buff []byte) (process.InterceptedData, error) {
	args := p2p.ArgInterceptedShardValidatorInfo{
		Marshaller:  isvif.marshaller,
		DataBuff:    buff,
		NumOfShards: isvif.shardCoordinator.NumberOfShards(),
	}

	return p2p.NewInterceptedShardValidatorInfo(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (isvif *interceptedValidatorInfoFactory) IsInterfaceNil() bool {
	return isvif == nil
}
