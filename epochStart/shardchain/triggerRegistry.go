package shardchain

import "encoding/json"

const keyPrefix = "epochStartTrigger_"

// TriggerRegistry holds the data required to correctly initialize the trigger when booting from saved state
type TriggerRegistry struct {
	Epoch                       uint32
	CurrentRoundIndex           int64
	EpochStartRound             uint64
	EpochMetaBlockHash          []byte
	IsEpochStart                bool
	EpochFinalityAttestingRound uint64
}

// LoadState loads into trigger the saved state
func (t *trigger) LoadState(key []byte) error {
	trigInternalKey := append([]byte(keyPrefix), key...)

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
	t.epoch = state.Epoch
	t.currentRoundIndex = state.CurrentRoundIndex
	t.epochStartRound = state.EpochStartRound
	t.epochMetaBlockHash = state.EpochMetaBlockHash
	t.isEpochStart = state.IsEpochStart
	t.epochFinalityAttestingRound = state.EpochFinalityAttestingRound
	t.mutTrigger.Unlock()

	return nil
}

// saveState saves the trigger state
func (t *trigger) saveState(key []byte) error {
	registry := &TriggerRegistry{}

	t.mutTrigger.RLock()
	registry.Epoch = t.epoch
	registry.CurrentRoundIndex = t.currentRoundIndex
	registry.EpochStartRound = t.epochStartRound
	registry.EpochMetaBlockHash = t.epochMetaBlockHash
	registry.IsEpochStart = t.isEpochStart
	registry.EpochFinalityAttestingRound = t.epochFinalityAttestingRound
	t.mutTrigger.RUnlock()

	data, err := json.Marshal(registry)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(keyPrefix), key...)
	log.Debug("saving start of epoch trigger state")

	return t.triggerStorage.Put(trigInternalKey, data)
}
