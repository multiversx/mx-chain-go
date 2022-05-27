package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type interceptedDirectConnectionInfoFactory struct {
	marshaller       marshal.Marshalizer
	shardCoordinator sharding.Coordinator
}

// NewInterceptedDirectConnectionInfoFactory creates an instance of interceptedDirectConnectionInfoFactory
func NewInterceptedDirectConnectionInfoFactory(args ArgInterceptedDataFactory) (*interceptedDirectConnectionInfoFactory, error) {
	err := checkInterceptedDirectConnectionInfoFactoryArgs(args)
	if err != nil {
		return nil, err
	}

	return &interceptedDirectConnectionInfoFactory{
		marshaller:       args.CoreComponents.InternalMarshalizer(),
		shardCoordinator: args.ShardCoordinator,
	}, nil
}

func checkInterceptedDirectConnectionInfoFactoryArgs(args ArgInterceptedDataFactory) error {
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
func (idcif *interceptedDirectConnectionInfoFactory) Create(buff []byte) (process.InterceptedData, error) {
	args := p2p.ArgInterceptedDirectConnectionInfo{
		Marshaller:  idcif.marshaller,
		DataBuff:    buff,
		NumOfShards: idcif.shardCoordinator.NumberOfShards(),
	}

	return p2p.NewInterceptedDirectConnectionInfo(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (idcif *interceptedDirectConnectionInfoFactory) IsInterfaceNil() bool {
	return idcif == nil
}
