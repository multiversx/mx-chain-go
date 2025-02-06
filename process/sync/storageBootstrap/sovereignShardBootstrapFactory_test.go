package storageBootstrap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/sync"
)

func TestNewSovereignShardBootstrapFactory(t *testing.T) {
	t.Parallel()

	ssbf := NewSovereignShardBootstrapFactory()
	require.False(t, ssbf.IsInterfaceNil())
}

func TestSovereignShardBootstrapFactory_CreateShardBootstrapFactory(t *testing.T) {
	t.Parallel()

	ssbf := NewSovereignShardBootstrapFactory()
	_, err := ssbf.CreateBootstrapper(sync.ArgShardBootstrapper{})
	require.NotNil(t, err)

	bootStrapper, err := ssbf.CreateBootstrapper(getDefaultArgs())
	require.Nil(t, err)
	require.Equal(t, "*sync.SovereignChainShardBootstrap", fmt.Sprintf("%T", bootStrapper))
}
