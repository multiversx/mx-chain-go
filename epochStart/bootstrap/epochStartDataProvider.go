package bootstrap

import (
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("registration")
var _ process.Interceptor = (*simpleMetaBlockInterceptor)(nil)

const requestSuffix = "_REQUEST"
const delayBetweenRequests = 200 * time.Millisecond
const thresholdForConsideringMetaBlockCorrect = 0.4
const numRequestsToSendOnce = 4

// ComponentsNeededForBootstrap holds the components which need to be initialized from network
type ComponentsNeededForBootstrap struct {
	EpochStartMetaBlock *block.MetaBlock
	NodesConfig         *sharding.NodesSetup
}

// epochStartDataProvider will handle requesting the needed data to start when joining late the network
type epochStartDataProvider struct {
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	messenger              p2p.Messenger
	nodesConfigProvider    NodesConfigProviderHandler
	metaBlockInterceptor   MetaBlockInterceptorHandler
	shardHeaderInterceptor ShardHeaderInterceptorHandler
	metaBlockResolver      MetaBlockResolverHandler
}

// ArgsEpochStartDataProvider holds the arguments needed for creating an epoch start data provider component
type ArgsEpochStartDataProvider struct {
	Messenger              p2p.Messenger
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	NodesConfigProvider    NodesConfigProviderHandler
	MetaBlockInterceptor   MetaBlockInterceptorHandler
	ShardHeaderInterceptor ShardHeaderInterceptorHandler
	MetaBlockResolver      MetaBlockResolverHandler
}

// NewEpochStartDataProvider will return a new instance of epochStartDataProvider
func NewEpochStartDataProvider(args ArgsEpochStartDataProvider) (*epochStartDataProvider, error) {
	if check.IfNil(args.Messenger) {
		return nil, ErrNilMessenger
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, ErrNilNodesConfigProvider
	}
	if check.IfNil(args.MetaBlockInterceptor) {
		return nil, ErrNilMetaBlockInterceptor
	}
	if check.IfNil(args.ShardHeaderInterceptor) {
		return nil, ErrNilShardHeaderInterceptor
	}
	if check.IfNil(args.MetaBlockResolver) {
		return nil, ErrNilMetaBlockResolver
	}
	return &epochStartDataProvider{
		marshalizer:            args.Marshalizer,
		hasher:                 args.Hasher,
		messenger:              args.Messenger,
		nodesConfigProvider:    args.NodesConfigProvider,
		metaBlockInterceptor:   args.MetaBlockInterceptor,
		shardHeaderInterceptor: args.ShardHeaderInterceptor,
		metaBlockResolver:      args.MetaBlockResolver,
	}, nil
}

// Bootstrap will handle requesting and receiving the needed information the node will bootstrap from
func (esdp *epochStartDataProvider) Bootstrap() (*ComponentsNeededForBootstrap, error) {
	err := esdp.initTopicsAndInterceptors()
	if err != nil {
		return nil, err
	}
	defer func() {
		esdp.resetTopicsAndInterceptors()
	}()

	epochNumForRequestingTheLatestAvailable := uint32(math.MaxUint32)
	metaBlock, err := esdp.getEpochStartMetaBlock(epochNumForRequestingTheLatestAvailable)
	if err != nil {
		return nil, err
	}
	prevMetaBlock, err := esdp.getEpochStartMetaBlock(metaBlock.Epoch - 1)
	if err != nil {
		return nil, err
	}
	log.Info("previous meta block", "epoch", prevMetaBlock.Epoch)
	nodesConfig, err := esdp.nodesConfigProvider.GetNodesConfigForMetaBlock(metaBlock)
	if err != nil {
		return nil, err
	}

	return &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: metaBlock,
		NodesConfig:         nodesConfig,
	}, nil
}

func (esdp *epochStartDataProvider) initTopicsAndInterceptors() error {
	err := esdp.messenger.CreateTopic(factory.MetachainBlocksTopic, true)
	if err != nil {
		return err
	}

	err = esdp.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, esdp.metaBlockInterceptor)
	if err != nil {
		return err
	}

	return nil
}

func (esdp *epochStartDataProvider) resetTopicsAndInterceptors() {
	err := esdp.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
	}
	err = esdp.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic + requestSuffix)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
	}
}

func (esdp *epochStartDataProvider) getEpochStartMetaBlock(epoch uint32) (*block.MetaBlock, error) {
	err := esdp.requestMetaBlock(epoch)
	if err != nil {
		return nil, err
	}
	for {
		numConnectedPeers := len(esdp.messenger.Peers())
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
		mb, errConsensusNotReached := esdp.metaBlockInterceptor.GetMetaBlock(threshold, epoch)
		if errConsensusNotReached == nil {
			return mb, nil
		}
		log.Info("consensus not reached for epoch start meta block. re-requesting and trying again...")
		err = esdp.requestMetaBlock(epoch)
		if err != nil {
			return nil, err
		}
	}
}

func (esdp *epochStartDataProvider) requestMetaBlock(epoch uint32) error {
	// send more requests
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		log.Debug("sent request for epoch start metablock...")
		err := esdp.metaBlockResolver.RequestEpochStartMetaBlock(epoch)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (esdp *epochStartDataProvider) IsInterfaceNil() bool {
	return esdp == nil
}
