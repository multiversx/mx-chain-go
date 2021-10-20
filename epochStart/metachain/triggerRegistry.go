package metachain

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// LoadState loads into trigger the saved state
func (t *trigger) LoadState(key []byte) error {
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("getting start of epoch trigger state", "key", trigInternalKey)

	d, err := t.triggerStorage.Get(trigInternalKey)
	if err != nil {
		return err
	}

	state, err := t.UnmarshallTrigger(d)
	if err != nil {
		return err
	}

	t.mutTrigger.Lock()
	t.triggerStateKey = key
	t.currentRound = state.CurrentRound
	t.epochFinalityAttestingRound = state.EpochFinalityAttestingRound
	t.currEpochStartRound = state.CurrEpochStartRound
	t.prevEpochStartRound = state.PrevEpochStartRound
	t.epoch = state.Epoch
	t.epochStartMetaHash = state.EpochStartMetaHash
	t.epochStartMeta = state.EpochStartMeta
	t.mutTrigger.Unlock()

	return nil
}

func (t *trigger) UnmarshallTrigger(data []byte) (*block.MetaTriggerRegistry, error) {
	state := &block.MetaTriggerRegistry{
		EpochStartMeta: &block.MetaBlock{},
	}
	err := t.marshalizer.Unmarshal(state, data)
	if err == nil {
		return state, nil
	}

	// for backwards compatibility
	err = json.Unmarshal(data, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// saveState saves the trigger state. Needs to be called under mutex
func (t *trigger) saveState(key []byte) error {
	metaHeader, ok := t.epochStartMeta.(*block.MetaBlock)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}
	registry := &block.MetaTriggerRegistry{}
	registry.CurrentRound = t.currentRound
	registry.EpochFinalityAttestingRound = t.epochFinalityAttestingRound
	registry.CurrEpochStartRound = t.currEpochStartRound
	registry.PrevEpochStartRound = t.prevEpochStartRound
	registry.Epoch = t.epoch
	registry.EpochStartMetaHash = t.epochStartMetaHash
	registry.EpochStartMeta = metaHeader
	triggerData, err := t.marshalizer.Marshal(registry)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("saving start of epoch trigger state", "key", trigInternalKey)

	return t.triggerStorage.Put(trigInternalKey, triggerData)
}
