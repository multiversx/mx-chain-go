package shardchain

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// LoadState loads into trigger the saved state
func (t *trigger) LoadState(key []byte) error {
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("getting start of epoch trigger state", "key", trigInternalKey)

	triggerData, err := t.triggerStorage.Get(trigInternalKey)
	if err != nil {
		return err
	}

	state, err := UnmarshalTrigger(t.marshalizer, triggerData)
	if err != nil {
		return err
	}

	t.mutTrigger.Lock()
	t.triggerStateKey = key
	t.epoch = state.GetEpoch()
	t.metaEpoch = state.GetMetaEpoch()
	t.currentRoundIndex = state.GetCurrentRoundIndex()
	t.epochStartRound = state.GetEpochStartRound()
	t.epochMetaBlockHash = state.GetEpochMetaBlockHash()
	t.isEpochStart = state.GetIsEpochStart()
	t.newEpochHdrReceived = state.GetNewEpochHeaderReceived()
	t.epochFinalityAttestingRound = state.GetEpochFinalityAttestingRound()
	t.epochStartShardHeader = state.GetEpochStartHeaderHandler()
	t.mutTrigger.Unlock()

	return nil
}

// UnmarshalTrigger unmarshalls the trigger
func UnmarshalTrigger(marshalizer marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	trig, err := UnmarshalTriggerV2(marshalizer, data)
	if err == nil {
		return trig, nil
	}

	return UnmarshalTriggerV1(marshalizer, data)
}

// UnmarshalTriggerV2 tries to unmarshal the data into a v2 trigger
func UnmarshalTriggerV2(marshalizer marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	triggerV2 := &block.ShardTriggerRegistryV2{
		EpochStartShardHeader: &block.HeaderV2{},
	}

	err := marshalizer.Unmarshal(triggerV2, data)
	if err != nil {
		return nil, err
	}

	if check.IfNil(triggerV2.EpochStartShardHeader) || check.IfNil(triggerV2.EpochStartShardHeader.Header) {
		return nil, fmt.Errorf("%w while checking inner epoch start shard header", epochStart.ErrNilHeaderHandler)
	}
	return triggerV2, nil
}

// UnmarshalTriggerV1 tries to unmarshal the data into a v1 trigger
func UnmarshalTriggerV1(marshalizer marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	triggerV1 := &block.ShardTriggerRegistry{
		EpochStartShardHeader: &block.Header{},
	}

	err := marshalizer.Unmarshal(triggerV1, data)
	if err != nil {
		return nil, err
	}

	if check.IfNil(triggerV1.EpochStartShardHeader) {
		return nil, fmt.Errorf("%w while checking inner epoch start shard header", epochStart.ErrNilHeaderHandler)
	}
	return triggerV1, nil
}

// saveState saves the trigger state. Needs to be called under mutex
func (t *trigger) saveState(key []byte) error {
	registryV2 := &block.ShardTriggerRegistryV2{
		EpochStartShardHeader: &block.HeaderV2{},
	}
	registryV1 := &block.ShardTriggerRegistry{
		EpochStartShardHeader: &block.Header{},
	}
	var registry data.TriggerRegistryHandler

	_, ok := t.epochStartShardHeader.(*block.HeaderV2)
	if ok {
		registry = registryV2
	} else {
		registry = registryV1
	}

	err := registry.SetMetaEpoch(t.metaEpoch)
	err = registry.SetEpoch(t.epoch)
	err = registry.SetCurrentRoundIndex(t.currentRoundIndex)
	err = registry.SetEpochStartRound(t.epochStartRound)
	err = registry.SetEpochMetaBlockHash(t.epochMetaBlockHash)
	err = registry.SetIsEpochStart(t.isEpochStart)
	err = registry.SetNewEpochHeaderReceived(t.newEpochHdrReceived)
	err = registry.SetEpochFinalityAttestingRound(t.epochFinalityAttestingRound)
	err = registry.SetEpochStartHeaderHandler(t.epochStartShardHeader)

	marshalledRegistry, err := t.marshalizer.Marshal(registry)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("saving start of epoch trigger state", "key", trigInternalKey)

	return t.triggerStorage.Put(trigInternalKey, marshalledRegistry)
}
