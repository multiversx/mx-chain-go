package metachain

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/stretchr/testify/require"
)

func TestMetaTriggerRegistryCreator_UnmarshalTriggerAndCreateRegistry(t *testing.T) {
	t.Parallel()

	registryCreator := NewMetaTriggerRegistryCreator()
	trig := createRegistryTrigger()

	registry, err := registryCreator.createRegistry(&block.HeaderV2{}, trig)
	require.Nil(t, registry)
	require.Equal(t, epochStart.ErrWrongTypeAssertion, err)

	metaHdr := &block.MetaBlock{
		Epoch: 4,
	}
	registry, err = registryCreator.createRegistry(metaHdr, trig)
	require.Nil(t, err)
	require.Equal(t, &block.MetaTriggerRegistry{
		Epoch:                       trig.epoch,
		CurrentRound:                trig.currentRound,
		EpochFinalityAttestingRound: trig.epochFinalityAttestingRound,
		CurrEpochStartRound:         trig.currEpochStartRound,
		PrevEpochStartRound:         trig.prevEpochStartRound,
		EpochStartMetaHash:          trig.epochStartMetaHash,
		EpochStartMeta:              metaHdr,
	}, registry)

	requireTriggerUnmarshallSuccessfully(t, registry, registryCreator)
}
