package factory_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/stretchr/testify/require"
)

func TestNewDirectStakedListProcessorFactory(t *testing.T) {
	t.Parallel()

	directStakedListHandlerFactory := trieIteratorsFactory.NewDirectStakedListProcessorFactory()
	require.False(t, directStakedListHandlerFactory.IsInterfaceNil())
}

func TestDirectStakedListProcessorFactory_CreateDirectStakedListProcessorHandlerDisabledProcessor(t *testing.T) {
	t.Parallel()

	args := createMockArgs(0)

	directStakedListHandler, err := trieIteratorsFactory.NewDirectStakedListProcessorFactory().CreateDirectStakedListProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*disabled.directStakedListProcessor", fmt.Sprintf("%T", directStakedListHandler))
}

func TestDirectStakedListProcessorFactory_CreateDirectStakedListProcessorHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgs(core.MetachainShardId)

	directStakedListHandler, err := trieIteratorsFactory.NewDirectStakedListProcessorFactory().CreateDirectStakedListProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*trieIterators.directStakedListProcessor", fmt.Sprintf("%T", directStakedListHandler))
}
