package factory_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignDirectStakedListProcessorFactory(t *testing.T) {
	t.Parallel()

	directStakedListHandlerFactory := trieIteratorsFactory.NewSovereignDirectStakedListProcessorFactory()
	require.False(t, directStakedListHandlerFactory.IsInterfaceNil())
}

func TestSovereignDirectStakedListHandlerFactory_CreateDirectStakedListProcessorHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgs(core.SovereignChainShardId)

	sovereignDirectStakedListHandler, err := trieIteratorsFactory.NewSovereignDirectStakedListProcessorFactory().CreateDirectStakedListProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*trieIterators.directStakedListProcessor", fmt.Sprintf("%T", sovereignDirectStakedListHandler))
}
