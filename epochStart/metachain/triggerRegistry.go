package metachain

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
)

// TriggerRegistry holds the data required to correctly initialize the trigger when booting from saved state
type TriggerRegistry struct {
	Epoch                       uint32
	CurrentRound                uint64
	EpochFinalityAttestingRound uint64
	CurrEpochStartRound         uint64
	PrevEpochStartRound         uint64
	EpochStartMetaHash          []byte
	EpochStartMeta              *block.MetaBlock
}

// LoadState loads into trigger the saved state
func (t *trigger) LoadState(key []byte) error {
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("getting start of epoch trigger state", "key", trigInternalKey)

	data, err := t.triggerStorage.Get(trigInternalKey)
	if err != nil {
		return err
	}

	state := &TriggerRegistry{}
	err = json.Unmarshal(data, state)
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

// saveState saves the trigger state. Needs to be called under mutex
func (t *trigger) saveState(key []byte) error {
	registry := &TriggerRegistry{}
	registry.CurrentRound = t.currentRound
	registry.EpochFinalityAttestingRound = t.epochFinalityAttestingRound
	registry.CurrEpochStartRound = t.currEpochStartRound
	registry.PrevEpochStartRound = t.prevEpochStartRound
	registry.Epoch = t.epoch
	registry.EpochStartMetaHash = t.epochStartMetaHash
	registry.EpochStartMeta = t.epochStartMeta
	data, err := json.Marshal(registry)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("saving start of epoch trigger state", "key", trigInternalKey)

	return t.triggerStorage.Put(trigInternalKey, data)
}
