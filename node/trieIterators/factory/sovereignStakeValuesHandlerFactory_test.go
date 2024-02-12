package factory_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignTotalStakedValueProcessorFactory(t *testing.T) {
	t.Parallel()

	sovereignTotalStakedValueHandlerFactory := trieIteratorsFactory.NewSovereignTotalStakedValueProcessorFactory()
	require.False(t, sovereignTotalStakedValueHandlerFactory.IsInterfaceNil())
}

func TestSovereignTotalStakedValueProcessorFactory_CreateSovereignTotalStakedValueProcessorHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgs(core.SovereignChainShardId)

	sovereignTotalStakedValueHandlerFactory, err := trieIteratorsFactory.NewSovereignTotalStakedValueProcessorFactory().CreateTotalStakedValueProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*trieIterators.stakedValuesProcessor", fmt.Sprintf("%T", sovereignTotalStakedValueHandlerFactory))
}
