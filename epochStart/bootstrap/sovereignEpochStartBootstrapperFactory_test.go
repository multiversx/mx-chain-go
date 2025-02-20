package bootstrap

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
)

func TestNewSovereignEpochStartBootstrapperFactory(t *testing.T) {
	t.Parallel()

	sebf := NewSovereignEpochStartBootstrapperFactory()
	require.False(t, sebf.IsInterfaceNil())
}

func TestSovereignEpochStartBootstrapperFactory_CreateEpochStartBootstrapper(t *testing.T) {
	t.Parallel()

	sebf := NewSovereignEpochStartBootstrapperFactory()
	seb, err := sebf.CreateEpochStartBootstrapper(getDefaultArgs())
	require.Nil(t, err)
	require.Equal(t, "*bootstrap.sovereignChainEpochStartBootstrap", fmt.Sprintf("%T", seb))
}

func TestSovereignEpochStartBootstrapperFactory_CreateStorageEpochStartBootstrapper(t *testing.T) {
	t.Parallel()

	sebf := NewSovereignEpochStartBootstrapperFactory()
	arg := ArgsStorageEpochStartBootstrap{
		ArgsEpochStartBootstrap:    getDefaultArgs(),
		ImportDbConfig:             config.ImportDbConfig{},
		ChanGracefullyClose:        make(chan endProcess.ArgEndProcess, 1),
		TimeToWaitForRequestedData: 1,
	}
	esb, err := sebf.CreateStorageEpochStartBootstrapper(arg)
	require.Nil(t, err)
	require.Equal(t, "*bootstrap.storageEpochStartBootstrap", fmt.Sprintf("%T", esb))

}
