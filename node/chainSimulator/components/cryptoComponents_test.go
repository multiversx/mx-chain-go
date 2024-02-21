package components

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/stretchr/testify/require"
)

func createArgsCryptoComponentsHolder() ArgsCryptoComponentsHolder {
	return ArgsCryptoComponentsHolder{
		Config: config.Config{
			Consensus: config.ConsensusConfig{
				Type: "bls",
			},
			MultisigHasher: config.TypeConfig{
				Type: "blake2b",
			},
			PublicKeyPIDSignature: config.CacheConfig{
				Capacity: 1000,
				Type:     "LRU",
			},
		},
		EnableEpochsConfig: config.EnableEpochs{
			BLSMultiSignerEnableEpoch: []config.MultiSignerConfig{
				{
					EnableEpoch: 0,
					Type:        "no-KOSK",
				},
				{
					EnableEpoch: 10,
					Type:        "KOSK",
				},
			},
		},
		Preferences: config.Preferences{},
		CoreComponentsHolder: &factory.CoreComponentsHolderStub{
			ValidatorPubKeyConverterCalled: func() core.PubkeyConverter {
				return &testscommon.PubkeyConverterStub{
					EncodeCalled: func(pkBytes []byte) (string, error) {
						return "public key", nil
					},
				}
			},
		},
		AllValidatorKeysPemFileName: "allValidatorKeys.pem",
		BypassTxSignatureCheck:      false,
	}
}

func TestCreateCryptoComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateCryptoComponents(createArgsCryptoComponentsHolder())
		require.NoError(t, err)
		require.NotNil(t, comp)

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	t.Run("should work with bypass tx sig check", func(t *testing.T) {
		t.Parallel()

		args := createArgsCryptoComponentsHolder()
		args.BypassTxSignatureCheck = true
		comp, err := CreateCryptoComponents(args)
		require.NoError(t, err)
		require.NotNil(t, comp)

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	t.Run("NewCryptoComponentsFactory failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCryptoComponentsHolder()
		args.CoreComponentsHolder = &factory.CoreComponentsHolderStub{
			ValidatorPubKeyConverterCalled: func() core.PubkeyConverter {
				return nil
			},
		}
		comp, err := CreateCryptoComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("managedCryptoComponents.Create failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCryptoComponentsHolder()
		args.CoreComponentsHolder = &factory.CoreComponentsHolderStub{
			ValidatorPubKeyConverterCalled: func() core.PubkeyConverter {
				return &testscommon.PubkeyConverterStub{
					EncodeCalled: func(pkBytes []byte) (string, error) {
						return "", expectedErr
					},
				}
			},
		}
		comp, err := CreateCryptoComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
}

func TestCryptoComponentsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var comp *cryptoComponentsHolder
	require.True(t, comp.IsInterfaceNil())

	comp, _ = CreateCryptoComponents(createArgsCryptoComponentsHolder())
	require.False(t, comp.IsInterfaceNil())
}

func TestCryptoComponentsHolder_GettersSetters(t *testing.T) {
	t.Parallel()

	comp, err := CreateCryptoComponents(createArgsCryptoComponentsHolder())
	require.NoError(t, err)

	require.NotNil(t, comp.PublicKey())
	require.NotNil(t, comp.PrivateKey())
	require.NotEmpty(t, comp.PublicKeyString())
	require.NotEmpty(t, comp.PublicKeyBytes())
	require.NotNil(t, comp.P2pPublicKey())
	require.NotNil(t, comp.P2pPrivateKey())
	require.NotNil(t, comp.P2pSingleSigner())
	require.NotNil(t, comp.TxSingleSigner())
	require.NotNil(t, comp.BlockSigner())
	container := comp.MultiSignerContainer()
	require.NotNil(t, container)
	require.Nil(t, comp.SetMultiSignerContainer(nil))
	require.Nil(t, comp.MultiSignerContainer())
	require.Nil(t, comp.SetMultiSignerContainer(container))
	signer, err := comp.GetMultiSigner(0)
	require.NoError(t, err)
	require.NotNil(t, signer)
	require.NotNil(t, comp.PeerSignatureHandler())
	require.NotNil(t, comp.BlockSignKeyGen())
	require.NotNil(t, comp.TxSignKeyGen())
	require.NotNil(t, comp.P2pKeyGen())
	require.NotNil(t, comp.MessageSignVerifier())
	require.NotNil(t, comp.ConsensusSigningHandler())
	require.NotNil(t, comp.ManagedPeersHolder())
	require.NotNil(t, comp.KeysHandler())
	require.Nil(t, comp.CheckSubcomponents())
	require.Empty(t, comp.String())
}

func TestCryptoComponentsHolder_Clone(t *testing.T) {
	t.Parallel()

	comp, err := CreateCryptoComponents(createArgsCryptoComponentsHolder())
	require.NoError(t, err)

	compClone := comp.Clone()
	require.Equal(t, comp, compClone)
	require.False(t, comp == compClone) // pointer testing
}
