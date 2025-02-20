package broadcast

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignDelayedBroadcastData struct {
	*delayedBlockBroadcaster
}

// NewSovereignDelayedBlockBroadcaster creates a sovereign delayed block broadcaster
func NewSovereignDelayedBlockBroadcaster(args *ArgsDelayedBlockBroadcaster) (*sovereignDelayedBroadcastData, error) {
	dbb, err := baseCreateDelayedBroadcaster(args)
	if err != nil {
		return nil, err
	}

	sdbb := &sovereignDelayedBroadcastData{
		dbb,
	}
	err = sdbb.registerHeaderInterceptorCallback(sdbb.interceptedHeader)
	if err != nil {
		return nil, err
	}

	err = sdbb.registerMiniBlockInterceptorCallback(sdbb.interceptedMiniBlockData)
	if err != nil {
		return nil, err
	}

	return sdbb, nil
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

func (dbb *sovereignDelayedBroadcastData) extractMiniBlockHashesCrossFromMe(_ data.HeaderHandler) map[string]map[string]struct{} {
	return make(map[string]map[string]struct{})
}

// IsInterfaceNil returns true if there is no value under the interface
func (dbb *sovereignDelayedBroadcastData) IsInterfaceNil() bool {
	return dbb == nil
}
