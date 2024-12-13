package factory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignShardRequestersContainerCreator_CreateShardRequestersContainerFactory(t *testing.T) {
	t.Parallel()

	creator := NewSovereignShardRequestersContainerCreator()
	require.False(t, creator.IsInterfaceNil())

	args := createFactoryArgs()
	container, err := creator.CreateShardRequestersContainerFactory(args)
	require.Nil(t, err)
	require.Equal(t, "*storagerequesterscontainer.sovereignShardRequestersContainerFactory", fmt.Sprintf("%T", container))
}
