package metachain

import (
	"encoding/json"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/epochStart"
)

type metaTriggerRegistryCreator struct {
}

// NewMetaTriggerRegistryCreator creates a new sovereign trigger registry creator/retriever
func NewMetaTriggerRegistryCreator() *metaTriggerRegistryCreator {
	return &metaTriggerRegistryCreator{}
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

// UnmarshalTrigger unmarshalls the meta trigger with provided marshaller if possible, otherwise with json, for backwards compatibility
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
