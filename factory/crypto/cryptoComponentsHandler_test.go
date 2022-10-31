package crypto_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	cryptoComp "github.com/ElrondNetwork/elrond-go/factory/crypto"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedCryptoComponents --------------------
func TestManagedCryptoComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "invalid"
	cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
	managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	require.NoError(t, err)
	err = managedCryptoComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedCryptoComponents.BlockSignKeyGen())
}

func TestManagedCryptoComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
	managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedCryptoComponents.TxSingleSigner())
	require.Nil(t, managedCryptoComponents.BlockSigner())
	require.Nil(t, managedCryptoComponents.MultiSignerContainer())
	require.Nil(t, managedCryptoComponents.BlockSignKeyGen())
	require.Nil(t, managedCryptoComponents.TxSignKeyGen())
	require.Nil(t, managedCryptoComponents.MessageSignVerifier())

	err = managedCryptoComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedCryptoComponents.TxSingleSigner())
	require.NotNil(t, managedCryptoComponents.BlockSigner())
	require.NotNil(t, managedCryptoComponents.MultiSignerContainer())
	multiSigner, errGet := managedCryptoComponents.MultiSignerContainer().GetMultiSigner(0)
	require.NotNil(t, multiSigner)
	require.Nil(t, errGet)
	require.NotNil(t, managedCryptoComponents.BlockSignKeyGen())
	require.NotNil(t, managedCryptoComponents.TxSignKeyGen())
	require.NotNil(t, managedCryptoComponents.MessageSignVerifier())
}

func TestManagedCryptoComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	managedCryptoComponents := getManagedCryptoComponents(t)

	err := managedCryptoComponents.CheckSubcomponents()
	require.NoError(t, err)
}

func TestManagedCryptoComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	managedCryptoComponents := getManagedCryptoComponents(t)

	err := managedCryptoComponents.Close()
	require.NoError(t, err)
	multiSigner, errGet := managedCryptoComponents.GetMultiSigner(0)
	require.Nil(t, multiSigner)
	require.Equal(t, errors.ErrNilCryptoComponentsHolder, errGet)
}

func getManagedCryptoComponents(t *testing.T) factory.CryptoComponentsHandler {
	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
	require.NotNil(t, cryptoComponentsFactory)
	managedCryptoComponents, _ := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	require.NotNil(t, managedCryptoComponents)
	err := managedCryptoComponents.Create()
	require.NoError(t, err)

	return managedCryptoComponents
}

func TestManagedCryptoComponents_Clone(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
	managedCryptoComponents, _ := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
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
