package metachain

import (
	"github.com/multiversx/mx-chain-go/common"
)

// LoadState loads into trigger the saved state
func (t *trigger) LoadState(key []byte) error {
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("getting start of epoch trigger state", "key", trigInternalKey)

	d, err := t.triggerStorage.Get(trigInternalKey)
	if err != nil {
		return err
	}

	state, err := t.registryHandler.UnmarshalTrigger(t.marshaller, d)
	if err != nil {
		return err
	}

	t.mutTrigger.Lock()
	t.triggerStateKey = key
	t.currentRound = state.GetCurrentRound()
	t.epochFinalityAttestingRound = state.GetEpochFinalityAttestingRound()
	t.currEpochStartRound = state.GetCurrEpochStartRound()
	t.prevEpochStartRound = state.GetPrevEpochStartRound()
	t.epoch = state.GetEpoch()
	t.epochStartMetaHash = state.GetEpochStartMetaHash()
	t.epochStartMeta = state.GetEpochStartHeaderHandler()
	t.mutTrigger.Unlock()

	return nil
}

// saveState saves the trigger state. Needs to be called under mutex
func (t *trigger) saveState(key []byte) error {
	registry, err := t.registryHandler.createRegistry(t.epochStartMeta, t)
	if err != nil {
		return err
	}

	triggerData, err := t.marshaller.Marshal(registry)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("saving start of epoch trigger state", "key", trigInternalKey)

	return t.triggerStorage.Put(trigInternalKey, triggerData)
}
