package components

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
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
			ProcessConfigsHandlerCalled: func() common.ProcessConfigsHandler {
				return &testscommon.ProcessConfigsHandlerStub{}
			},
		},
		AllValidatorKeysPemFileName: "allValidatorKeys.pem",
		BypassTxSignatureCheck:      true,
		BypassBlockSignatureCheck:   false,
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
	t.Run("should work with bypass blocks sig check", func(t *testing.T) {
		t.Parallel()

		args := createArgsCryptoComponentsHolder()
		args.BypassBlockSignatureCheck = true
		comp, err := CreateCryptoComponents(args)
		require.NoError(t, err)
		require.NotNil(t, comp)
		require.Equal(t, "*singlesig.DisabledSingleSig", fmt.Sprintf("%T", comp.blockSigner))

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	t.Run("should install deterministic fast consensus crypto", func(t *testing.T) {
		t.Parallel()

		args := createArgsCryptoComponentsHolder()
		args.EnableFastConsensusCrypto = true
		comp, err := CreateCryptoComponents(args)
		require.NoError(t, err)
		require.IsType(t, &fastConsensusSigner{}, comp.BlockSigner())
		require.IsType(t, &fastConsensusMultiSignerContainer{}, comp.MultiSignerContainer())

		message := []byte("consensus message")
		signature, err := comp.ConsensusSigningHandler().CreateSignatureForPublicKey(
			message,
			comp.PublicKeyBytes(),
		)
		require.NoError(t, err)
		require.NoError(t, comp.ConsensusSigningHandler().VerifySingleSignature(
			comp.PublicKeyBytes(),
			message,
			signature,
		))
		require.Error(t, comp.ConsensusSigningHandler().VerifySingleSignature(
			comp.PublicKeyBytes(),
			[]byte("different message"),
			signature,
		))

		peerID := core.PeerID("simulator peer")
		peerSignature, err := comp.PeerSignatureHandler().GetPeerSignature(
			comp.PrivateKey(),
			peerID.Bytes(),
		)
		require.NoError(t, err)
		require.NoError(t, comp.PeerSignatureHandler().VerifyPeerSignature(
			comp.PublicKeyBytes(),
			peerID,
			peerSignature,
		))

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
	require.Nil(t, comp.Close())
}

func TestFastConsensusCrypto_SigningHandlerBindsAggregateToBitmap(t *testing.T) {
	args := createArgsCryptoComponentsHolder()
	args.EnableFastConsensusCrypto = true

	first, err := CreateCryptoComponents(args)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, first.Close())
	}()
	second, err := CreateCryptoComponents(args)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, second.Close())
	}()

	message := []byte("proof payload")
	firstShare, err := first.ConsensusSigningHandler().CreateSignatureForPublicKey(
		message,
		first.PublicKeyBytes(),
	)
	require.NoError(t, err)
	secondShare, err := second.ConsensusSigningHandler().CreateSignatureForPublicKey(
		message,
		second.PublicKeyBytes(),
	)
	require.NoError(t, err)

	publicKeys := []string{
		string(first.PublicKeyBytes()),
		string(second.PublicKeyBytes()),
	}
	bitmap := []byte{0b00000011}
	aggregatedSignature, err := first.ConsensusSigningHandler().AggregateSigsWithKeys(
		publicKeys,
		bitmap,
		[][]byte{firstShare, secondShare},
		0,
	)
	require.NoError(t, err)
	require.NoError(t, first.ConsensusSigningHandler().VerifyAggregatedSigWithKeys(
		publicKeys,
		bitmap,
		message,
		aggregatedSignature,
		0,
	))

	require.Error(t, first.ConsensusSigningHandler().VerifyAggregatedSigWithKeys(
		publicKeys,
		[]byte{0b00000001},
		message,
		aggregatedSignature,
		0,
	))
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
	require.Nil(t, comp.Close())
}

func TestCryptoComponentsHolder_Clone(t *testing.T) {
	t.Parallel()

	comp, err := CreateCryptoComponents(createArgsCryptoComponentsHolder())
	require.NoError(t, err)

	compClone := comp.Clone()
	require.Equal(t, comp, compClone)
	require.False(t, comp == compClone) // pointer testing
	require.Nil(t, comp.Close())
}
