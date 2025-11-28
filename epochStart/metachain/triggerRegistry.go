package metachain

import (
	"encoding/json"

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

	d, err := t.triggerStorage.Get(trigInternalKey)
	if err != nil {
		return err
	}

	state, err := UnmarshalTrigger(t.marshaller, d)
	if err != nil {
		return err
	}

	t.mutTrigger.Lock()
	t.triggerStateKey = key
	t.epochFinalityAttestingRound = state.EpochFinalityAttestingRound
	t.currEpochStartRound = state.CurrEpochStartRound
	t.prevEpochStartRound = state.PrevEpochStartRound
	t.epoch = state.Epoch
	t.epochStartMetaHash = state.EpochStartMetaHash
	t.epochStartMeta = state.EpochStartMeta
	t.mutTrigger.Unlock()

	return nil
}

// UnmarshalTrigger unmarshalls the trigger with json, for backwards compatibility
func UnmarshalTrigger(marshaller marshal.Marshalizer, data []byte) (*block.MetaTriggerRegistry, error) {
	state := &block.MetaTriggerRegistry{
		EpochStartMeta: &block.MetaBlock{},
	}

	err := marshaller.Unmarshal(state, data)
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
	// TODO: add new structures for the new headers and update this component
	metaHeader, ok := t.epochStartMeta.(data.MetaHeaderHandler)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	registry := t.createRegistryHandler()
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

func (t *trigger) createRegistryHandler() data.MetaTriggerRegistryHandler {
	if t.epochStartMeta.IsHeaderV3() {
		return &block.MetaTriggerRegistryV3{}
	}

	return &block.MetaTriggerRegistry{}
}
