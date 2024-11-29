package api

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	factoryDisabled "github.com/multiversx/mx-chain-go/factory/disabled"
	"github.com/stretchr/testify/require"
)

func TestSovereignApiProcessorCompsCreator_CreateAPIComps(t *testing.T) {
	factory := NewSovereignAPIProcessorCompsCreator()
	require.False(t, factory.IsInterfaceNil())

	disabledComps := &APIProcessComps{
		StakingDataProviderAPI: factoryDisabled.NewDisabledStakingDataProvider(),
		AuctionListSelector:    factoryDisabled.NewDisabledAuctionListSelector(),
	}

	args := createArgs(core.SovereignChainShardId)
	apiComps, err := factory.CreateAPIComps(args)
	require.Nil(t, err)
	require.NotEqual(t, apiComps, disabledComps)
}
