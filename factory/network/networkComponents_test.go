package network_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	errErd "github.com/multiversx/mx-chain-go/errors"
	networkComp "github.com/multiversx/mx-chain-go/factory/network"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewNetworkComponentsFactory_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetNetworkFactoryArgs()
	args.StatusHandler = nil
	ncf, err := networkComp.NewNetworkComponentsFactory(args)
	require.Nil(t, ncf)
	require.Equal(t, errErd.ErrNilStatusHandler, err)
}

func TestNewNetworkComponentsFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetNetworkFactoryArgs()
	args.Marshalizer = nil
	ncf, err := networkComp.NewNetworkComponentsFactory(args)
	require.Nil(t, ncf)
	require.True(t, errors.Is(err, errErd.ErrNilMarshalizer))
}

func TestNewNetworkComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetNetworkFactoryArgs()
	ncf, err := networkComp.NewNetworkComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, ncf)
}

func TestNetworkComponentsFactory_CreateShouldErrDueToBadConfig(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetNetworkFactoryArgs()
	args.MainConfig = config.Config{}
	args.P2pConfig = p2pConfig.P2PConfig{}

	ncf, _ := networkComp.NewNetworkComponentsFactory(args)

	nc, err := ncf.Create()
	require.Error(t, err)
	require.Nil(t, nc)
}

func TestNetworkComponentsFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetNetworkFactoryArgs()
	ncf, _ := networkComp.NewNetworkComponentsFactory(args)

	nc, err := ncf.Create()
	require.NoError(t, err)
	require.NotNil(t, nc)
}

// ------------ Test NetworkComponents --------------------
func TestNetworkComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetNetworkFactoryArgs()
	ncf, _ := networkComp.NewNetworkComponentsFactory(args)

	nc, err := ncf.Create()
	require.Nil(t, err)

	err = nc.Close()
	require.NoError(t, err)
}
