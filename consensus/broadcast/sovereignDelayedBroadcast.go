package broadcast

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/storage/cache"
)

type sovereignDelayedBroadcastData struct {
	*delayedBlockBroadcaster
}

func NewSovereignDelayedBlockBroadcaster(args *ArgsDelayedBlockBroadcaster) (*sovereignDelayedBroadcastData, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, spos.ErrNilShardCoordinator
	}
	if check.IfNil(args.InterceptorsContainer) {
		return nil, spos.ErrNilInterceptorsContainer
	}
	if check.IfNil(args.HeadersSubscriber) {
		return nil, spos.ErrNilHeadersSubscriber
	}
	if check.IfNil(args.AlarmScheduler) {
		return nil, spos.ErrNilAlarmScheduler
	}

	cacheHeaders, err := cache.NewLRUCache(sizeHeadersCache)
	if err != nil {
		return nil, err
	}

	dbb := &sovereignDelayedBroadcastData{
		&delayedBlockBroadcaster{
			alarm:                      args.AlarmScheduler,
			shardCoordinator:           args.ShardCoordinator,
			interceptorsContainer:      args.InterceptorsContainer,
			headersSubscriber:          args.HeadersSubscriber,
			valHeaderBroadcastData:     make([]*validatorHeaderBroadcastData, 0),
			valBroadcastData:           make([]*delayedBroadcastData, 0),
			delayedBroadcastData:       make([]*delayedBroadcastData, 0),
			maxDelayCacheSize:          args.LeaderCacheSize,
			maxValidatorDelayCacheSize: args.ValidatorCacheSize,
			mutDataForBroadcast:        sync.RWMutex{},
			cacheHeaders:               cacheHeaders,
			mutHeadersCache:            sync.RWMutex{},
		},
	}
	err = dbb.registerHeaderInterceptorCallback(dbb.interceptedHeader)
	if err != nil {
		return nil, err
	}

	err = dbb.registerMiniBlockInterceptorCallback(dbb.interceptedMiniBlockData)
	if err != nil {
		return nil, err
	}

	return dbb, nil
}

func (dbb *sovereignDelayedBroadcastData) registerHeaderInterceptorCallback(
	cb func(topic string, hash []byte, data interface{}),
) error {
	identifierShardHeader := factory.ShardBlocksTopic + dbb.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	interceptor, err := dbb.interceptorsContainer.Get(identifierShardHeader)
	if err != nil {
		return err
	}

	interceptor.RegisterHandler(cb)
	return nil
}

func (dbb *sovereignDelayedBroadcastData) registerMiniBlockInterceptorCallback(
	cb func(topic string, hash []byte, data interface{}),
) error {
	identifierMiniBlocks := factory.MiniBlocksTopic + dbb.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	interceptor, err := dbb.interceptorsContainer.Get(identifierMiniBlocks)
	if err != nil {
		return err
	}

	interceptor.RegisterHandler(cb)
	return nil
}

// SetValidatorData sets the data for consensus validator delayed broadcast
func (dbb *sovereignDelayedBroadcastData) SetValidatorData(broadcastData *delayedBroadcastData) error {
	return dbb.setValidatorData(broadcastData, dbb.extractMiniBlockHashesCrossFromMe)
}

func (dbb *sovereignDelayedBroadcastData) extractMiniBlockHashesCrossFromMe(header data.HeaderHandler) map[string]map[string]struct{} {
	mbHashesForShards := make(map[string]map[string]struct{})
	topic := factory.MiniBlocksTopic + dbb.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	mbs := dbb.extractMbsFromMeTo(header, core.SovereignChainShardId)
	if len(mbs) > 0 {
		mbHashesForShards[topic] = mbs
	}

	return mbHashesForShards
}
