package crypto_test

import (
	"strings"
	"testing"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	cryptoComp "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestManagedCryptoComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(nil)
		require.Equal(t, errorsMx.ErrNilCryptoComponentsFactory, err)
		require.Nil(t, managedCryptoComponents)
	})
	t.Run("invalid args should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetCryptoArgs(coreComponents)
		args.Config.Consensus.Type = "invalid"
		cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
		managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
		require.NoError(t, err)
		err = managedCryptoComponents.Create()
		require.Error(t, err)
		require.Nil(t, managedCryptoComponents.BlockSignKeyGen())
	})
	t.Run("pub key mismatch", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetCryptoArgs(coreComponents)
		args.Config.Consensus.Type = "disabled"
		cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
		managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
		require.NoError(t, err)
		err = managedCryptoComponents.Create()
		require.True(t, strings.Contains(err.Error(), errorsMx.ErrPublicKeyMismatch.Error()))
	})
	t.Run("should work with activateBLSPubKeyMessageVerification", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetCryptoArgs(coreComponents)
		args.ActivateBLSPubKeyMessageVerification = true
		cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
		managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
		require.NoError(t, err)
		err = managedCryptoComponents.Create()
		require.NoError(t, err)
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

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
		require.Nil(t, managedCryptoComponents.PublicKey())
		require.Nil(t, managedCryptoComponents.PrivateKey())
		require.Nil(t, managedCryptoComponents.P2pPrivateKey())
		require.Nil(t, managedCryptoComponents.P2pPublicKey())
		require.Empty(t, managedCryptoComponents.PublicKeyString())
		require.Nil(t, managedCryptoComponents.PublicKeyBytes())
		require.Nil(t, managedCryptoComponents.P2pPrivateKey())
		require.Nil(t, managedCryptoComponents.P2pSingleSigner())
		require.Nil(t, managedCryptoComponents.PeerSignatureHandler())
		require.Nil(t, managedCryptoComponents.P2pKeyGen())
		require.Nil(t, managedCryptoComponents.ManagedPeersHolder())
		multiSigner, errGet := managedCryptoComponents.GetMultiSigner(0)
		require.Nil(t, multiSigner)
		require.Equal(t, errorsMx.ErrNilCryptoComponentsHolder, errGet)

		err = managedCryptoComponents.Create()
		require.NoError(t, err)
		require.NotNil(t, managedCryptoComponents.TxSingleSigner())
		require.NotNil(t, managedCryptoComponents.BlockSigner())
		require.NotNil(t, managedCryptoComponents.MultiSignerContainer())
		multiSigner, errGet = managedCryptoComponents.GetMultiSigner(0)
		require.NotNil(t, multiSigner)
		require.Nil(t, errGet)
		require.NotNil(t, managedCryptoComponents.BlockSignKeyGen())
		require.NotNil(t, managedCryptoComponents.TxSignKeyGen())
		require.NotNil(t, managedCryptoComponents.MessageSignVerifier())
		require.NotNil(t, managedCryptoComponents.PublicKey())
		require.NotNil(t, managedCryptoComponents.PrivateKey())
		require.NotNil(t, managedCryptoComponents.P2pPrivateKey())
		require.NotNil(t, managedCryptoComponents.P2pPublicKey())
		require.NotEmpty(t, managedCryptoComponents.PublicKeyString())
		require.NotNil(t, managedCryptoComponents.PublicKeyBytes())
		require.NotNil(t, managedCryptoComponents.P2pSingleSigner())
		require.NotNil(t, managedCryptoComponents.PeerSignatureHandler())
		require.NotNil(t, managedCryptoComponents.P2pKeyGen())
		require.NotNil(t, managedCryptoComponents.ManagedPeersHolder())

		require.Equal(t, factory.CryptoComponentsName, managedCryptoComponents.String())

		err = managedCryptoComponents.Close()
		require.NoError(t, err)

		err = managedCryptoComponents.Close()
		require.NoError(t, err)
	})
}

func TestNewManagedCryptoComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
	managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	require.NoError(t, err)
	require.Equal(t, errorsMx.ErrNilCryptoComponents, managedCryptoComponents.CheckSubcomponents())

	err = managedCryptoComponents.Create()
	require.NoError(t, err)
	require.Nil(t, managedCryptoComponents.CheckSubcomponents())
}

func TestNewManagedCryptoComponents_SetMultiSignerContainer(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
	managedCryptoComponents, _ := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	_ = managedCryptoComponents.Create()

	require.Equal(t, errorsMx.ErrNilMultiSignerContainer, managedCryptoComponents.SetMultiSignerContainer(nil))
	require.Nil(t, managedCryptoComponents.SetMultiSignerContainer(&mock.CryptoComponentsStub{}))
}

func TestManagedCryptoComponents_Clone(t *testing.T) {
	t.Parallel()

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

func TestNewManagedCryptoComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(nil)
	require.Equal(t, errorsMx.ErrNilCryptoComponentsFactory, err)
	require.True(t, managedCryptoComponents.IsInterfaceNil())

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(args)
	managedCryptoComponents, err = cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	require.NoError(t, err)
	require.False(t, managedCryptoComponents.IsInterfaceNil())
}
