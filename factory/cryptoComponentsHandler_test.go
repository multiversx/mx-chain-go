package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedCryptoComponents --------------------
func TestManagedCryptoComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "invalid"
	cryptoComponentsFactory, _ := factory.NewCryptoComponentsFactory(args)
	managedCryptoComponents, err := factory.NewManagedCryptoComponents(cryptoComponentsFactory)
	require.NoError(t, err)
	err = managedCryptoComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedCryptoComponents.BlockSignKeyGen())
}

func TestManagedCryptoComponents_Create_ShouldWork(t *testing.T) {
	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := factory.NewCryptoComponentsFactory(args)
	managedCryptoComponents, err := factory.NewManagedCryptoComponents(cryptoComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedCryptoComponents.TxSingleSigner())
	require.Nil(t, managedCryptoComponents.BlockSigner())
	require.Nil(t, managedCryptoComponents.MultiSigner())
	require.Nil(t, managedCryptoComponents.BlockSignKeyGen())
	require.Nil(t, managedCryptoComponents.TxSignKeyGen())
	require.Nil(t, managedCryptoComponents.MessageSignVerifier())

	err = managedCryptoComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedCryptoComponents.TxSingleSigner())
	require.NotNil(t, managedCryptoComponents.BlockSigner())
	require.NotNil(t, managedCryptoComponents.MultiSigner())
	require.NotNil(t, managedCryptoComponents.BlockSignKeyGen())
	require.NotNil(t, managedCryptoComponents.TxSignKeyGen())
	require.NotNil(t, managedCryptoComponents.MessageSignVerifier())
}

func TestManagedCryptoComponents_Close(t *testing.T) {
	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := factory.NewCryptoComponentsFactory(args)
	managedCryptoComponents, _ := factory.NewManagedCryptoComponents(cryptoComponentsFactory)
	err := managedCryptoComponents.Create()
	require.NoError(t, err)

	err = managedCryptoComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedCryptoComponents.MultiSigner())
}

func TestManagedCryptoComponents_Clone(t *testing.T) {
	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := factory.NewCryptoComponentsFactory(args)
	managedCryptoComponents, _ := factory.NewManagedCryptoComponents(cryptoComponentsFactory)
	err := managedCryptoComponents.Create()
	require.NoError(t, err)

	clonedBeforeCreate := managedCryptoComponents.Clone()
	require.Equal(t, managedCryptoComponents, clonedBeforeCreate)

	_ = managedCryptoComponents.Create()
	clonedAfterCreate := managedCryptoComponents.Clone()
	require.Equal(t, managedCryptoComponents, clonedAfterCreate)

	_ = managedCryptoComponents.Close()
	clonedAfterClose := managedCryptoComponents.Clone()
	require.Equal(t, managedCryptoComponents, clonedAfterClose)
}
