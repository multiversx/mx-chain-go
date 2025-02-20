package factory_test

import (
	"fmt"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTotalStakedValueProcessorFactory(t *testing.T) {
	t.Parallel()

	totalStakedValueHandlerFactory := trieIteratorsFactory.NewTotalStakedListProcessorFactory()
	require.False(t, totalStakedValueHandlerFactory.IsInterfaceNil())
}

func TestTotalStakedValueProcessorFactory_CreateTotalStakedValueProcessorHandlerDisabledProcessor(t *testing.T) {
	t.Parallel()

	args := createMockArgs(0)

	totalStakedValueHandlerFactory, err := trieIteratorsFactory.NewTotalStakedListProcessorFactory().CreateTotalStakedValueProcessorHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*disabled.stakeValuesProcessor", fmt.Sprintf("%T", totalStakedValueHandlerFactory))
}

func TestTotalStakedValueProcessorFactory_CreateTotalStakedValueProcessorHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgs(core.MetachainShardId)

	totalStakedValueHandlerFactory, err := trieIteratorsFactory.NewTotalStakedListProcessorFactory().CreateTotalStakedValueProcessorHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*trieIterators.stakedValuesProcessor", fmt.Sprintf("%T", totalStakedValueHandlerFactory))
}
