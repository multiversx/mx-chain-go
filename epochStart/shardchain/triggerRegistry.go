package shardchain

import (
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
)

// LoadState loads into trigger the saved state
func (t *trigger) LoadState(key []byte) error {
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("getting start of epoch trigger state", "key", trigInternalKey)

	triggerData, err := t.triggerStorage.Get(trigInternalKey)
	if err != nil {
		return err
	}

	state, err := UnmarshalTrigger(t.marshaller, triggerData)
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
func UnmarshalTrigger(marshaller marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	trig, err := unmarshalTriggerV3(marshaller, data)
	if err == nil {
		return trig, nil
	}

	trig, err = unmarshalTriggerV2(marshaller, data)
	if err == nil {
		return trig, nil
	}

	trig, err = unmarshalTriggerV1(marshaller, data)
	if err == nil {
		return trig, nil
	}

	// for backwards compatibility
	return unmarshalTriggerJson(data)
}

// unmarshalTriggerJson unmarshalls the trigger with json, for backwards compatibility
func unmarshalTriggerJson(data []byte) (data.TriggerRegistryHandler, error) {
	trig := &block.ShardTriggerRegistry{EpochStartShardHeader: &block.Header{}}
	err := json.Unmarshal(data, trig)
	if err != nil {
		return nil, err
	}

	return trig, nil
}

// unmarshalTriggerV3 tries to unmarshal the data into a v3 trigger
func unmarshalTriggerV3(marshaller marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	triggerV3 := &block.ShardTriggerRegistryV3{
		EpochStartShardHeader: &block.HeaderV3{},
	}

	err := marshaller.Unmarshal(triggerV3, data)
	if err != nil {
		return nil, err
	}

	if check.IfNil(triggerV3.EpochStartShardHeader) {
		return nil, fmt.Errorf("%w while checking inner epoch start shard header", epochStart.ErrNilHeaderHandler)
	}
	return triggerV3, nil
}

// unmarshalTriggerV2 tries to unmarshal the data into a v2 trigger
func unmarshalTriggerV2(marshaller marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	triggerV2 := &block.ShardTriggerRegistryV2{
		EpochStartShardHeader: &block.HeaderV2{},
	}

	err := marshaller.Unmarshal(triggerV2, data)
	if err != nil {
		return nil, err
	}

	if check.IfNil(triggerV2.EpochStartShardHeader) || check.IfNil(triggerV2.EpochStartShardHeader.Header) {
		return nil, fmt.Errorf("%w while checking inner epoch start shard header", epochStart.ErrNilHeaderHandler)
	}
	return triggerV2, nil
}

// unmarshalTriggerV1 tries to unmarshal the data into a v1 trigger
func unmarshalTriggerV1(marshaller marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	triggerV1 := &block.ShardTriggerRegistry{
		EpochStartShardHeader: &block.Header{},
	}

	err := marshaller.Unmarshal(triggerV1, data)
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
	registry := t.createRegistryHandler()

	_ = registry.SetMetaEpoch(t.metaEpoch)
	_ = registry.SetEpoch(t.epoch)
	_ = registry.SetCurrentRoundIndex(t.currentRoundIndex)
	_ = registry.SetEpochStartRound(t.epochStartRound)
	_ = registry.SetEpochMetaBlockHash(t.epochMetaBlockHash)
	_ = registry.SetIsEpochStart(t.isEpochStart)
	_ = registry.SetNewEpochHeaderReceived(t.newEpochHdrReceived)
	_ = registry.SetEpochFinalityAttestingRound(t.epochFinalityAttestingRound)
	_ = registry.SetEpochStartHeaderHandler(t.epochStartShardHeader)

	marshalledRegistry, err := t.marshaller.Marshal(registry)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("saving start of epoch trigger state", "key", trigInternalKey)

	return t.triggerStorage.Put(trigInternalKey, marshalledRegistry)
}

func (t *trigger) createRegistryHandler() data.TriggerRegistryHandler {
	if t.epochStartShardHeader.IsHeaderV3() {
		return &block.ShardTriggerRegistryV3{
			EpochStartShardHeader: &block.HeaderV3{},
		}
	}

	_, ok := t.epochStartShardHeader.(*block.HeaderV2)
	if ok {
		return &block.ShardTriggerRegistryV2{
			EpochStartShardHeader: &block.HeaderV2{},
		}
	}

	return &block.ShardTriggerRegistry{
		EpochStartShardHeader: &block.Header{},
	}
}
