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
	ProofsPool      dataRetriever.ProofsPool
	HeadersPool     dataRetriever.HeadersPool
	Storage         dataRetriever.StorageService
	PeerShardMapper process.PeerShardMapper
}

type interceptedEquivalentProofsFactory struct {
	marshaller         marshal.Marshalizer
	shardCoordinator   sharding.Coordinator
	headerSigVerifier  consensus.HeaderSigVerifier
	proofsPool         dataRetriever.ProofsPool
	headersPool        dataRetriever.HeadersPool
	storage            dataRetriever.StorageService
	hasher             hashing.Hasher
	proofSizeChecker   common.FieldsSizeChecker
	km                 sync.KeyRWMutexHandler
	eligibleNodesCache process.EligibleNodesCache
}

// NewInterceptedEquivalentProofsFactory creates a new instance of interceptedEquivalentProofsFactory
func NewInterceptedEquivalentProofsFactory(args ArgInterceptedEquivalentProofsFactory) (*interceptedEquivalentProofsFactory, error) {
	enc, err := newEligibleNodesCache(args.PeerShardMapper, args.NodesCoordinator)
	if err != nil {
		return nil, err
	}

	return &interceptedEquivalentProofsFactory{
		marshaller:         args.CoreComponents.InternalMarshalizer(),
		shardCoordinator:   args.ShardCoordinator,
		headerSigVerifier:  args.HeaderSigVerifier,
		proofsPool:         args.ProofsPool,
		headersPool:        args.HeadersPool,
		storage:            args.Storage,
		hasher:             args.CoreComponents.Hasher(),
		proofSizeChecker:   args.CoreComponents.FieldsSizeChecker(),
		km:                 sync.NewKeyRWMutex(),
		eligibleNodesCache: enc,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (factory *interceptedEquivalentProofsFactory) Create(buff []byte, messageOriginator core.PeerID) (process.InterceptedData, error) {
	args := interceptedBlocks.ArgInterceptedEquivalentProof{
		DataBuff:           buff,
		Marshaller:         factory.marshaller,
		ShardCoordinator:   factory.shardCoordinator,
		HeaderSigVerifier:  factory.headerSigVerifier,
		Proofs:             factory.proofsPool,
		Headers:            factory.headersPool,
		Hasher:             factory.hasher,
		ProofSizeChecker:   factory.proofSizeChecker,
		KeyRWMutexHandler:  factory.km,
		EligibleNodesCache: factory.eligibleNodesCache,
		MessageOriginator:  messageOriginator,
	}
	return interceptedBlocks.NewInterceptedEquivalentProof(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (factory *interceptedEquivalentProofsFactory) IsInterfaceNil() bool {
	return factory == nil
}
