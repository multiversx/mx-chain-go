package storageBootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/sync"
)

func TestNewSovereignShardBootstrapFactory(t *testing.T) {
	t.Parallel()

	ssbf, err := NewSovereignShardBootstrapFactory(nil)

	require.Nil(t, ssbf)
	require.Equal(t, errors.ErrNilShardBootstrapFactory, err)

	sbf := NewShardBootstrapFactory()
	ssbf, err = NewSovereignShardBootstrapFactory(sbf)

	require.NotNil(t, ssbf)
	require.Nil(t, err)
}

func TestSovereignShardBootstrapFactory_CreateShardBootstrapFactory(t *testing.T) {
	t.Parallel()

	sbf := NewShardBootstrapFactory()
	ssbf, _ := NewSovereignShardBootstrapFactory(sbf)

	_, err := ssbf.CreateBootstrapper(sync.ArgShardBootstrapper{})

	require.NotNil(t, err)

	bootStrapper, err := ssbf.CreateBootstrapper(getDefaultArgs())

	require.NotNil(t, bootStrapper)
	require.Nil(t, err)
}

func TestSovereignShardBootstrapFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sbf := NewShardBootstrapFactory()
	ssbf, _ := NewSovereignShardBootstrapFactory(sbf)

	require.False(t, ssbf.IsInterfaceNil())

	ssbf = nil
	require.True(t, ssbf.IsInterfaceNil())
}
