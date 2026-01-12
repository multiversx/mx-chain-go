package shardchain

import (
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

	state, err := epochStart.UnmarshalShardTrigger(t.marshaller, triggerData)
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

// saveState saves the trigger state. Needs to be called under mutex
func (t *trigger) saveState(key []byte) error {
	registry := epochStart.CreateShardRegistryHandler(t.epochStartShardHeader)
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
