package bootstrap

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
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
}

type metaBlockInterceptorHandler interface {
	process.Interceptor
	GetMetaBlock(target int) (*block.MetaBlock, error)
}

type shardHeaderInterceptorHandler interface {
	process.Interceptor
	GetAllReceivedShardHeaders() []block.ShardData
}

type metaBlockResolverHandler interface {
	RequestEpochStartMetaBlock() error
}

// epochStartDataProvider will handle requesting the needed data to start when joining late the network
type epochStartDataProvider struct {
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	messenger              p2p.Messenger
	metaBlockInterceptor   metaBlockInterceptorHandler
	shardHeaderInterceptor shardHeaderInterceptorHandler
	metaBlockResolver      metaBlockResolverHandler
}

// NewEpochStartDataProvider will return a new instance of epochStartDataProvider
func NewEpochStartDataProvider(
	messenger p2p.Messenger,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (*epochStartDataProvider, error) {
	metaBlockInterceptor := NewSimpleMetaBlockInterceptor(marshalizer, hasher)
	shardHdrInterceptor := NewSimpleShardHeaderInterceptor(marshalizer)
	metaBlockResolver, err := NewSimpleMetaBlocksResolver(messenger, marshalizer)
	if err != nil {
		return nil, err
	}

	return &epochStartDataProvider{
		marshalizer:            marshalizer,
		hasher:                 hasher,
		messenger:              messenger,
		metaBlockInterceptor:   metaBlockInterceptor,
		shardHeaderInterceptor: shardHdrInterceptor,
		metaBlockResolver:      metaBlockResolver,
	}, nil
}

// Bootstrap will handle requesting and receiving the needed information the node will bootstrap from
func (esdp *epochStartDataProvider) Bootstrap() (*ComponentsNeededForBootstrap, error) {
	metaBlock, err := esdp.getEpochStartMetaBlock()
	if err != nil {
		return nil, err
	}

	return &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: metaBlock,
	}, nil
}

func (esdp *epochStartDataProvider) getEpochStartMetaBlock() (*block.MetaBlock, error) {
	err := esdp.messenger.CreateTopic(factory.MetachainBlocksTopic, true)
	if err != nil {
		return nil, err
	}

	err = esdp.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, esdp.metaBlockInterceptor)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = esdp.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
		if err != nil {
			log.Info("error unregistering message processor", "error", err)
		}
		err = esdp.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic + requestSuffix)
		if err != nil {
			log.Info("error unregistering message processor", "error", err)
		}
	}()

	err = esdp.requestMetaBlock()
	if err != nil {
		return nil, err
	}

	for {
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(len(esdp.messenger.Peers())))
		mb, errConsensusNotReached := esdp.metaBlockInterceptor.GetMetaBlock(threshold)
		if errConsensusNotReached == nil {
			return mb, nil
		}
		log.Info("consensus not reached for epoch start meta block. re-requesting and trying again...")
		err = esdp.requestMetaBlock()
		if err != nil {
			return nil, err
		}
	}
}

func (esdp *epochStartDataProvider) requestMetaBlock() error {
	// send more requests
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		log.Debug("sent request for epoch start metablock...")
		err := esdp.metaBlockResolver.RequestEpochStartMetaBlock()
		if err != nil {
			return err
		}
	}

	return nil
}
