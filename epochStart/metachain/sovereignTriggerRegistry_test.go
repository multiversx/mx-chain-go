package metachain

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/stretchr/testify/require"
)

func createRegistryTrigger() *trigger {
	return &trigger{
		currentRound:                4,
		epochFinalityAttestingRound: 5,
		currEpochStartRound:         100,
		prevEpochStartRound:         99,
		epoch:                       6,
		epochStartMetaHash:          []byte("hash"),
	}
}

func requireTriggerUnmarshallSuccessfully(t *testing.T, registry data.MetaTriggerRegistryHandler, registryCreator registryHandler) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	registryData, err := marshaller.Marshal(registry)
	require.Nil(t, err)

	savedRegistry, err := registryCreator.UnmarshalTrigger(marshaller, registryData)
	require.Nil(t, err)
	require.Equal(t, registry, savedRegistry)
}

func TestSovereignTriggerRegistryCreator_UnmarshalTriggerAndCreateRegistry(t *testing.T) {
	t.Parallel()

	registryCreator := NewSovereignTriggerRegistryCreator()
	trig := createRegistryTrigger()

	registry, err := registryCreator.createRegistry(&block.HeaderV2{}, trig)
	require.Nil(t, registry)
	require.Equal(t, epochStart.ErrWrongTypeAssertion, err)

	sovHdr := &block.SovereignChainHeader{
		IsStartOfEpoch: true,
	}
	registry, err = registryCreator.createRegistry(sovHdr, trig)
	require.Nil(t, err)
	require.Equal(t, &block.SovereignShardTriggerRegistry{
		Epoch:                       trig.epoch,
		CurrentRound:                trig.currentRound,
		EpochFinalityAttestingRound: trig.epochFinalityAttestingRound,
		CurrEpochStartRound:         trig.currEpochStartRound,
		PrevEpochStartRound:         trig.prevEpochStartRound,
		EpochStartMetaHash:          trig.epochStartMetaHash,
		SovereignChainHeader:        sovHdr,
	}, registry)

	requireTriggerUnmarshallSuccessfully(t, registry, registryCreator)
}
