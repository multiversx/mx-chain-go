package metachain

import (
	"encoding/json"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
)

type registryHandler interface {
	createRegistry(header data.HeaderHandler, t *trigger) (data.MetaTriggerRegistryHandler, error)
	UnmarshalTrigger(marshaller marshal.Marshalizer, data []byte) (data.MetaTriggerRegistryHandler, error)
}

type metaTriggerRegistryCreator struct {
}

type sovereignTriggerRegistryCreator struct {
}

func NewMetaTriggerHandler() *metaTriggerRegistryCreator {
	return &metaTriggerRegistryCreator{}
}

func NewSovereignTriggerHandler() *sovereignTriggerRegistryCreator {
	return &sovereignTriggerRegistryCreator{}
}

func (mt *metaTriggerRegistryCreator) createRegistry(header data.HeaderHandler, t *trigger) (data.MetaTriggerRegistryHandler, error) {
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

func (mt *metaTriggerRegistryCreator) UnmarshalTrigger(marshaller marshal.Marshalizer, data []byte) (data.MetaTriggerRegistryHandler, error) {
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

func (st *sovereignTriggerRegistryCreator) createRegistry(header data.HeaderHandler, t *trigger) (data.MetaTriggerRegistryHandler, error) {
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

func (st *sovereignTriggerRegistryCreator) UnmarshalTrigger(marshaller marshal.Marshalizer, data []byte) (data.MetaTriggerRegistryHandler, error) {
	state := &block.SovereignShardTriggerRegistry{
		SovereignChainHeader: &block.SovereignChainHeader{},
	}

	err := marshaller.Unmarshal(state, data)
	if err != nil {
		return nil, err
	}
	return state, nil
}

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
