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

	sovereignDelegatedListHandlerFactory := trieIteratorsFactory.NewSovereignDelegatedListProcessorFactory()
	require.False(t, sovereignDelegatedListHandlerFactory.IsInterfaceNil())
}

func TestSovereignTotalStakedValueProcessorFactory_CreateSovereignTotalStakedValueProcessorHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgs(core.SovereignChainShardId)

	sovereignDelegatedListHandler, err := trieIteratorsFactory.NewSovereignTotalStakedValueProcessorFactory().CreateTotalStakedValueProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*trieIterators.stakedValuesProcessor", fmt.Sprintf("%T", sovereignDelegatedListHandler))
}
