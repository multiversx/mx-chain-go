package shardchain

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// TriggerRegistry holds the data required to correctly initialize the trigger when booting from saved state
type TriggerRegistry struct {
	IsEpochStart                bool
	NewEpochHeaderReceived      bool
	Epoch                       uint32
	MetaEpoch                   uint32
	CurrentRoundIndex           int64
	EpochStartRound             uint64
	EpochFinalityAttestingRound uint64
	EpochMetaBlockHash          []byte
	EpochStartShardHeader       data.HeaderHandler
}

// LoadState loads into trigger the saved state
func (t *trigger) LoadState(key []byte) error {
	trigInternalKey := append([]byte(core.TriggerRegistryKeyPrefix), key...)
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
	t.epoch = state.Epoch
	t.metaEpoch = state.MetaEpoch
	t.currentRoundIndex = state.CurrentRoundIndex
	t.epochStartRound = state.EpochStartRound
	t.epochMetaBlockHash = state.EpochMetaBlockHash
	t.isEpochStart = state.IsEpochStart
	t.newEpochHdrReceived = state.NewEpochHeaderReceived
	t.epochFinalityAttestingRound = state.EpochFinalityAttestingRound
	t.epochStartShardHeader = state.EpochStartShardHeader
	t.mutTrigger.Unlock()

	return nil
}

// saveState saves the trigger state. Needs to be called under mutex
func (t *trigger) saveState(key []byte) error {
	registry := &TriggerRegistry{}

	registry.MetaEpoch = t.metaEpoch
	registry.Epoch = t.epoch
	registry.CurrentRoundIndex = t.currentRoundIndex
	registry.EpochStartRound = t.epochStartRound
	registry.EpochMetaBlockHash = t.epochMetaBlockHash
	registry.IsEpochStart = t.isEpochStart
	registry.NewEpochHeaderReceived = t.newEpochHdrReceived
	registry.EpochFinalityAttestingRound = t.epochFinalityAttestingRound
	registry.EpochStartShardHeader = t.epochStartShardHeader

	//TODO: change to protoMarshalizer
	data, err := json.Marshal(registry)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(core.TriggerRegistryKeyPrefix), key...)
	log.Debug("saving start of epoch trigger state", "key", trigInternalKey)

	return t.triggerStorage.Put(trigInternalKey, data)
}
