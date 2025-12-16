package metachain

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
)

// LoadState loads into trigger the saved state
func (t *trigger) LoadState(key []byte) error {
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("getting start of epoch trigger state", "key", trigInternalKey)

	d, err := t.triggerStorage.Get(trigInternalKey)
	if err != nil {
		return err
	}

	state, err := epochStart.UnmarshalMetaTrigger(t.marshaller, d)
	if err != nil {
		return err
	}

	t.mutTrigger.Lock()
	t.triggerStateKey = key
	t.epochFinalityAttestingRound = state.GetEpochFinalityAttestingRound()
	t.currEpochStartRound = state.GetCurrEpochStartRound()
	t.prevEpochStartRound = state.GetPrevEpochStartRound()
	t.epoch = state.GetEpoch()
	t.epochStartMetaHash = state.GetEpochStartMetaHash()
	t.epochStartMeta = state.GetEpochStartMetaHeaderHandler()
	t.epochChangeProposed = state.GetEpochChangeProposed()
	t.mutTrigger.Unlock()

	return nil
}

// saveState saves the trigger state. Needs to be called under mutex
func (t *trigger) saveState(key []byte) error {
	metaHeader, ok := t.epochStartMeta.(data.MetaHeaderHandler)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	registry := epochStart.CreateMetaRegistryHandler(t.epochStartMeta)
	_ = registry.SetEpochFinalityAttestingRound(t.epochFinalityAttestingRound)
	_ = registry.SetCurrEpochStartRound(t.currEpochStartRound)
	_ = registry.SetPrevEpochStartRound(t.prevEpochStartRound)
	_ = registry.SetEpoch(t.epoch)
	_ = registry.SetEpochStartMetaHash(t.epochStartMetaHash)
	_ = registry.SetEpochStartMetaHeaderHandler(metaHeader)
	_ = registry.SetEpochChangeProposed(t.epochChangeProposed)

	triggerData, err := t.marshaller.Marshal(registry)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("saving start of epoch trigger state", "key", trigInternalKey)

	return t.triggerStorage.Put(trigInternalKey, triggerData)
}
