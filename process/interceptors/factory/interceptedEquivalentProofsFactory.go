package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/sync"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ArgInterceptedEquivalentProofsFactory is the DTO used to create a new instance of interceptedEquivalentProofsFactory
type ArgInterceptedEquivalentProofsFactory struct {
	ArgInterceptedDataFactory
	ProofsPool dataRetriever.ProofsPool
}

type interceptedEquivalentProofsFactory struct {
	marshaller        marshal.Marshalizer
	shardCoordinator  sharding.Coordinator
	headerSigVerifier consensus.HeaderSigVerifier
	proofsPool        dataRetriever.ProofsPool
	hasher            hashing.Hasher
	proofSizeChecker  common.FieldsSizeChecker
	km                sync.KeyRWMutexHandler
}

// NewInterceptedEquivalentProofsFactory creates a new instance of interceptedEquivalentProofsFactory
func NewInterceptedEquivalentProofsFactory(args ArgInterceptedEquivalentProofsFactory) *interceptedEquivalentProofsFactory {
	return &interceptedEquivalentProofsFactory{
		marshaller:        args.CoreComponents.InternalMarshalizer(),
		shardCoordinator:  args.ShardCoordinator,
		headerSigVerifier: args.HeaderSigVerifier,
		proofsPool:        args.ProofsPool,
		hasher:            args.CoreComponents.Hasher(),
		proofSizeChecker:  args.CoreComponents.FieldsSizeChecker(),
		km:                sync.NewKeyRWMutex(),
	}
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (factory *interceptedEquivalentProofsFactory) Create(buff []byte, _ core.PeerID) (process.InterceptedData, error) {
	args := interceptedBlocks.ArgInterceptedEquivalentProof{
		DataBuff:          buff,
		Marshaller:        factory.marshaller,
		ShardCoordinator:  factory.shardCoordinator,
		HeaderSigVerifier: factory.headerSigVerifier,
		Proofs:            factory.proofsPool,
		Hasher:            factory.hasher,
		ProofSizeChecker:  factory.proofSizeChecker,
		KeyRWMutexHandler: factory.km,
	}
	return interceptedBlocks.NewInterceptedEquivalentProof(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (factory *interceptedEquivalentProofsFactory) IsInterfaceNil() bool {
	return factory == nil
}
