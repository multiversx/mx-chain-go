package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/sync"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/p2p"

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
	ProofsPool  dataRetriever.ProofsPool
	HeadersPool dataRetriever.HeadersPool
}

type interceptedEquivalentProofsFactory struct {
	marshaller        marshal.Marshalizer
	shardCoordinator  sharding.Coordinator
	headerSigVerifier consensus.HeaderSigVerifier
	proofsPool        dataRetriever.ProofsPool
	headersPool       dataRetriever.HeadersPool
	hasher            hashing.Hasher
	proofSizeChecker  common.FieldsSizeChecker
	km                sync.KeyRWMutexHandler
	validityAttester  process.ValidityAttester
}

// NewInterceptedEquivalentProofsFactory creates a new instance of interceptedEquivalentProofsFactory
func NewInterceptedEquivalentProofsFactory(args ArgInterceptedEquivalentProofsFactory) *interceptedEquivalentProofsFactory {
	return &interceptedEquivalentProofsFactory{
		marshaller:        args.CoreComponents.InternalMarshalizer(),
		shardCoordinator:  args.ShardCoordinator,
		headerSigVerifier: args.HeaderSigVerifier,
		proofsPool:        args.ProofsPool,
		headersPool:       args.HeadersPool,
		hasher:            args.CoreComponents.Hasher(),
		proofSizeChecker:  args.CoreComponents.FieldsSizeChecker(),
		km:                sync.NewKeyRWMutex(),
		validityAttester:  args.ValidityAttester,
	}
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (factory *interceptedEquivalentProofsFactory) Create(buff []byte, _ core.PeerID, broadcastMethod p2p.BroadcastMethod) (process.InterceptedData, error) {
	args := interceptedBlocks.ArgInterceptedEquivalentProof{
		DataBuff:          buff,
		Marshaller:        factory.marshaller,
		ShardCoordinator:  factory.shardCoordinator,
		HeaderSigVerifier: factory.headerSigVerifier,
		Proofs:            factory.proofsPool,
		HeadersPool:       factory.headersPool,
		Hasher:            factory.hasher,
		ProofSizeChecker:  factory.proofSizeChecker,
		KeyRWMutexHandler: factory.km,
		ValidityAttester:  factory.validityAttester,
	}
	return interceptedBlocks.NewInterceptedEquivalentProof(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (factory *interceptedEquivalentProofsFactory) IsInterfaceNil() bool {
	return factory == nil
}
