package storageBootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignShardBootstrapFactory(t *testing.T) {
	t.Parallel()

	ssbf, err := NewSovereignShardBootstrapFactory(nil)

	require.Nil(t, ssbf)
	require.Equal(t, errors.ErrNilShardBootstrapFactory, err)

	sbf, _ := NewShardBootstrapFactory()
	ssbf, err = NewSovereignShardBootstrapFactory(sbf)

	require.NotNil(t, ssbf)
	require.Nil(t, err)
}

func TestSovereignShardBootstrapFactory_CreateShardBootstrapFactory(t *testing.T) {
	t.Parallel()

	sbf, _ := NewShardBootstrapFactory()
	ssbf, _ := NewSovereignShardBootstrapFactory(sbf)

	_, err := ssbf.CreateShardBootstrapFactory(sync.ArgShardBootstrapper{})

	require.NotNil(t, err)

	bootStrapper, err := ssbf.CreateShardBootstrapFactory(getDefaultArgs())

	require.NotNil(t, bootStrapper)
	require.Nil(t, err)
}

func TestSovereignShardBootstrapFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sbf, _ := NewShardBootstrapFactory()
	ssbf, _ := NewSovereignShardBootstrapFactory(sbf)

	require.False(t, ssbf.IsInterfaceNil())

	ssbf = nil
	require.True(t, ssbf.IsInterfaceNil())
}
