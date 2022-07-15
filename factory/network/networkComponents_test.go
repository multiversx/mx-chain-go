package network_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	errErd "github.com/ElrondNetwork/elrond-go/errors"
	componentsMock "github.com/ElrondNetwork/elrond-go/factory/mock/components"
	networkComp "github.com/ElrondNetwork/elrond-go/factory/network"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/stretchr/testify/require"
)

func TestNewNetworkComponentsFactory_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetNetworkFactoryArgs()
	args.StatusHandler = nil
	ncf, err := networkComp.NewNetworkComponentsFactory(args)
	require.Nil(t, ncf)
	require.Equal(t, errErd.ErrNilStatusHandler, err)
}

func TestNewNetworkComponentsFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetNetworkFactoryArgs()
	args.Marshalizer = nil
	ncf, err := networkComp.NewNetworkComponentsFactory(args)
	require.Nil(t, ncf)
	require.True(t, errors.Is(err, errErd.ErrNilMarshalizer))
}

func TestNewNetworkComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetNetworkFactoryArgs()
	ncf, err := networkComp.NewNetworkComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, ncf)
}

func TestNetworkComponentsFactory_CreateShouldErrDueToBadConfig(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetNetworkFactoryArgs()
	args.MainConfig = config.Config{}
	args.P2pConfig = config.P2PConfig{}

	ncf, _ := networkComp.NewNetworkComponentsFactory(args)

	nc, err := ncf.Create()
	require.Error(t, err)
	require.Nil(t, nc)
}

func TestNetworkComponentsFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetNetworkFactoryArgs()
	ncf, _ := networkComp.NewNetworkComponentsFactory(args)
	ncf.SetListenAddress(libp2p.ListenLocalhostAddrWithIp4AndTcp)

	nc, err := ncf.Create()
	require.NoError(t, err)
	require.NotNil(t, nc)
}

// ------------ Test NetworkComponents --------------------
func TestNetworkComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetNetworkFactoryArgs()
	ncf, _ := networkComp.NewNetworkComponentsFactory(args)

	nc, _ := ncf.Create()

	err := nc.Close()
	require.NoError(t, err)
}
