package shard

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPreProcessorContainerFactoryCreator_CreatePreProcessorContainerFactory(t *testing.T) {
	t.Parallel()

	f := NewPreProcessorContainerFactoryCreator()
	require.False(t, f.IsInterfaceNil())

	args := createMockPreProcessorsContainerFactoryArguments()
	containerFactory, err := f.CreatePreProcessorContainerFactory(args)
	require.Nil(t, err)
	require.Equal(t, "*shard.preProcessorsContainerFactory", fmt.Sprintf("%T", containerFactory))
}
