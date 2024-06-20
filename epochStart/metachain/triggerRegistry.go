package metachain

import (
	"encoding/json"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
)

type MetaTriggerRegistryHandler interface {
	GetEpoch() uint32
	GetCurrentRound() uint64
	GetEpochFinalityAttestingRound() uint64
	GetCurrEpochStartRound() uint64
	GetPrevEpochStartRound() uint64
	GetEpochStartMetaHash() []byte
}

type triggerRegistryCreator interface {
	createRegistry(header data.HeaderHandler, t *trigger) (MetaTriggerRegistryHandler, error)
}

type metaTriggerRegistryCreator struct {
}

type sovereignTriggerRegistryCreator struct {
}

func (mt *metaTriggerRegistryCreator) createRegistry(header data.HeaderHandler, t *trigger) (MetaTriggerRegistryHandler, error) {
	metaHeader, ok := header.(*block.MetaBlock)
	if !ok {
		return nil, epochStart.ErrWrongTypeAssertion
	}

	registry := &block.MetaTriggerRegistry{}
	registry.CurrentRound = t.currentRound
	registry.EpochFinalityAttestingRound = t.epochFinalityAttestingRound
	registry.CurrEpochStartRound = t.currEpochStartRound
	registry.PrevEpochStartRound = t.prevEpochStartRound
	registry.Epoch = t.epoch
	registry.EpochStartMetaHash = t.epochStartMetaHash
	registry.EpochStartMeta = metaHeader

	return registry, nil
}

func (st *sovereignTriggerRegistryCreator) createRegistry(header data.HeaderHandler, t *trigger) (MetaTriggerRegistryHandler, error) {
	sovChainHdr, ok := header.(*block.SovereignChainHeader)
	if !ok {
		return nil, epochStart.ErrWrongTypeAssertion
	}

	registry := &block.SovereignShardTriggerRegistry{}
	registry.CurrentRound = t.currentRound
	registry.EpochFinalityAttestingRound = t.epochFinalityAttestingRound
	registry.CurrEpochStartRound = t.currEpochStartRound
	registry.PrevEpochStartRound = t.prevEpochStartRound
	registry.Epoch = t.epoch
	registry.EpochStartMetaHash = t.epochStartMetaHash
	registry.SovereignChainHeader = sovChainHdr

	return registry, nil
}

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
	registry, err := t.registryCreator.createRegistry(t.epochStartMeta, t)
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
