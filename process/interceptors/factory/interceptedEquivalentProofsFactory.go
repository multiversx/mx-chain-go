package factory

import (
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
	"github.com/multiversx/mx-chain-go/sharding"
)

type interceptedEquivalentProofsFactory struct {
	marshaller        marshal.Marshalizer
	shardCoordinator  sharding.Coordinator
	headerSigVerifier consensus.HeaderSigVerifier
}

// NewInterceptedEquivalentProofsFactory creates a new instance of interceptedEquivalentProofsFactory
func NewInterceptedEquivalentProofsFactory(args ArgInterceptedDataFactory) *interceptedEquivalentProofsFactory {
	return &interceptedEquivalentProofsFactory{
		marshaller:        args.CoreComponents.InternalMarshalizer(),
		shardCoordinator:  args.ShardCoordinator,
		headerSigVerifier: args.HeaderSigVerifier,
	}
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (factory *interceptedEquivalentProofsFactory) Create(buff []byte) (process.InterceptedData, error) {
	args := interceptedBlocks.ArgInterceptedEquivalentProof{
		DataBuff:          buff,
		Marshaller:        factory.marshaller,
		ShardCoordinator:  factory.shardCoordinator,
		HeaderSigVerifier: factory.headerSigVerifier,
	}
	return interceptedBlocks.NewInterceptedEquivalentProof(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (factory *interceptedEquivalentProofsFactory) IsInterfaceNil() bool {
	return factory == nil
}
