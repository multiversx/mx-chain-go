package crypto_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	cryptoFactory "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsSigningHandler() cryptoFactory.ArgsSigningHandler {
	return cryptoFactory.ArgsSigningHandler{
		PubKeys:              []string{"pubkey1"},
		KeysHandler:          &testscommon.KeysHandlerStub{},
		MultiSignerContainer: &cryptoMocks.MultiSignerContainerMock{},
		KeyGenerator:         &cryptoMocks.KeyGenStub{},
		SingleSigner:         &cryptoMocks.SingleSignerStub{},
	}
}

func TestNewSigningHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil multi signer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.MultiSignerContainer = nil

		signer, err := cryptoFactory.NewSigningHandler(args)
		require.Nil(t, signer)
		require.Equal(t, cryptoFactory.ErrNilMultiSignerContainer, err)
	})
	t.Run("nil single signer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.SingleSigner = nil

		signer, err := cryptoFactory.NewSigningHandler(args)
		require.Nil(t, signer)
		require.Equal(t, cryptoFactory.ErrNilSingleSigner, err)
	})
	t.Run("nil key generator", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.KeyGenerator = nil

		signer, err := cryptoFactory.NewSigningHandler(args)
		require.Nil(t, signer)
		require.Equal(t, cryptoFactory.ErrNilKeyGenerator, err)
	})
	t.Run("nil keys handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.KeysHandler = nil

		signer, err := cryptoFactory.NewSigningHandler(args)
		require.Nil(t, signer)
		require.Equal(t, cryptoFactory.ErrNilKeysHandler, err)
	})
	t.Run("no public keys", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{}

		signer, err := cryptoFactory.NewSigningHandler(args)
		require.Nil(t, signer)
		require.Equal(t, cryptoFactory.ErrNoPublicKeySet, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		signer, err := cryptoFactory.NewSigningHandler(args)
		require.Nil(t, err)
		require.False(t, check.IfNil(signer))
	})
}

func TestSigningHandler_Create(t *testing.T) {
	t.Parallel()

	t.Run("empty pubkeys in list", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()

		signer, err := cryptoFactory.NewSigningHandler(args)
		require.Nil(t, err)
		require.NotNil(t, signer)

		pubKeys := []string{"pubKey1", ""}
		createdSigner, err := signer.Create(pubKeys)
		require.Nil(t, createdSigner)
		require.Equal(t, cryptoFactory.ErrEmptyPubKeyString, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()

		signer, err := cryptoFactory.NewSigningHandler(args)
		require.Nil(t, err)
		require.NotNil(t, signer)

		pubKeys := []string{"pubKey1", "pubKey2"}
		createdSigner, err := signer.Create(pubKeys)
		require.Nil(t, err)
		require.NotNil(t, createdSigner)
	})
}

func TestSigningHandler_Reset(t *testing.T) {
	t.Parallel()

	t.Run("nil public keys", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()

		signer, _ := cryptoFactory.NewSigningHandler(args)
		err := signer.Reset(nil)
		require.Equal(t, cryptoFactory.ErrNilPublicKeys, err)
	})
	t.Run("empty pubkeys in list", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()

		signer, _ := cryptoFactory.NewSigningHandler(args)
		err := signer.Reset([]string{"pubKey1", ""})
		require.Equal(t, cryptoFactory.ErrEmptyPubKeyString, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()

		signer, _ := cryptoFactory.NewSigningHandler(args)
		err := signer.Reset([]string{"pubKey1", "pubKey2"})
		require.Nil(t, err)
	})
}

func TestSigningHandler_CreateSignatureShareForPublicKey(t *testing.T) {
	t.Parallel()

	selfIndex := uint16(0)
	epoch := uint32(0)
	pkBytes := []byte("public key bytes")

	t.Run("nil message", func(t *testing.T) {
		t.Parallel()

		signer, _ := cryptoFactory.NewSigningHandler(createMockArgsSigningHandler())
		sigShare, err := signer.CreateSignatureShareForPublicKey(nil, selfIndex, epoch, pkBytes)
		require.Nil(t, sigShare)
		require.Equal(t, cryptoFactory.ErrNilMessage, err)
	})
	t.Run("create sig share failed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()

		expectedErr := errors.New("expected error")
		multiSigner := &cryptoMocks.MultiSignerStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSigningHandler(args)
		sigShare, err := signer.CreateSignatureShareForPublicKey([]byte("msg1"), selfIndex, epoch, pkBytes)
		require.Nil(t, sigShare)
		require.Equal(t, expectedErr, err)
	})
	t.Run("failed to get current multi signer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()

		expectedErr := errors.New("expected error")
		args.MultiSignerContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return nil, expectedErr
			},
		}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		sigShare, err := signer.CreateSignatureShareForPublicKey([]byte("message"), uint16(0), epoch, pkBytes)
		require.Nil(t, sigShare)
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		getHandledPrivateKeyCalled := false

		expectedSigShare := []byte("sigShare")
		multiSigner := &cryptoMocks.MultiSignerStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return expectedSigShare, nil
			},
		}
		args.KeysHandler = &testscommon.KeysHandlerStub{
			GetHandledPrivateKeyCalled: func(providedPkBytes []byte) crypto.PrivateKey {
				assert.Equal(t, pkBytes, providedPkBytes)
				getHandledPrivateKeyCalled = true

				return &cryptoMocks.PrivateKeyStub{}
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSigningHandler(args)
		sigShare, err := signer.CreateSignatureShareForPublicKey([]byte("msg1"), selfIndex, epoch, pkBytes)
		require.Nil(t, err)
		require.Equal(t, expectedSigShare, sigShare)
		assert.True(t, getHandledPrivateKeyCalled)
	})
}

func TestSigningHandler_VerifySignatureShare(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(1)
	epoch := uint32(0)
	msg := []byte("message")

	t.Run("invalid signature share", func(t *testing.T) {
		t.Parallel()

		signer, _ := cryptoFactory.NewSigningHandler(createMockArgsSigningHandler())
		err := signer.VerifySignatureShare(ownIndex, nil, msg, epoch)
		require.Equal(t, cryptoFactory.ErrInvalidSignature, err)
	})
	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		signer, _ := cryptoFactory.NewSigningHandler(createMockArgsSigningHandler())
		err := signer.VerifySignatureShare(uint16(3), []byte("sigShare"), msg, epoch)
		require.Equal(t, cryptoFactory.ErrIndexOutOfBounds, err)
	})
	t.Run("signature share verification failed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2"}

		expectedErr := errors.New("expected error")
		multiSigner := &cryptoMocks.MultiSignerStub{
			VerifySignatureShareCalled: func(publicKey, message, sig []byte) error {
				return expectedErr
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSigningHandler(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), msg, epoch)
		require.Equal(t, expectedErr, err)
	})
	t.Run("failed to get current multi signer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2"}

		expectedErr := errors.New("expected error")
		args.MultiSignerContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return nil, expectedErr
			},
		}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), msg, epoch)
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2"}

		multiSigner := &cryptoMocks.MultiSignerStub{
			VerifySignatureShareCalled: func(publicKey, message, sig []byte) error {
				return nil
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSigningHandler(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), msg, epoch)
		require.Nil(t, err)
	})
}

func TestSigningHandler_StoreSignatureShare(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(2)

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		signer, err := cryptoFactory.NewSigningHandler(createMockArgsSigningHandler())
		require.Nil(t, err)

		err = signer.StoreSignatureShare(uint16(2), []byte("sigShare"))
		require.Equal(t, cryptoFactory.ErrIndexOutOfBounds, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		multiSigner := &cryptoMocks.MultiSignerStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return []byte("sigshare"), nil
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSigningHandler(args)

		sigShare := []byte("signature share")

		err := signer.StoreSignatureShare(ownIndex, sigShare)
		require.Nil(t, err)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, err)
		require.Equal(t, sigShare, sigShareRead)
	})
}

func TestSigningHandler_SignatureShare(t *testing.T) {
	t.Parallel()

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		index := uint16(1)
		sigShare := []byte("sig share")

		args := createMockArgsSigningHandler()

		signer, _ := cryptoFactory.NewSigningHandler(args)

		_ = signer.StoreSignatureShare(index, sigShare)

		sigShareRead, err := signer.SignatureShare(index)
		require.Nil(t, sigShareRead)
		require.Equal(t, cryptoFactory.ErrIndexOutOfBounds, err)
	})
	t.Run("nil element at index", func(t *testing.T) {
		t.Parallel()

		ownIndex := uint16(1)

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		_ = signer.StoreSignatureShare(ownIndex, nil)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, sigShareRead)
		require.Equal(t, cryptoFactory.ErrNilElement, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ownIndex := uint16(1)
		sigShare := []byte("sig share")

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		_ = signer.StoreSignatureShare(ownIndex, sigShare)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, err)
		require.Equal(t, sigShare, sigShareRead)
	})
}

func TestSigningHandler_AggregateSigs(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)

	t.Run("nil bitmap", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		aggSig, err := signer.AggregateSigs(nil, epoch)
		require.Nil(t, aggSig)
		require.Equal(t, cryptoFactory.ErrNilBitmap, err)
	})
	t.Run("bitmap mismatch", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4",
			"pk5", "pk6", "pk7", "pk8", "pk9"}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		aggSig, err := signer.AggregateSigs(bitmap, epoch)
		require.Nil(t, aggSig)
		require.Equal(t, cryptoFactory.ErrBitmapMismatch, err)
	})
	t.Run("failed to get aggregated sig", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		expectedErr := errors.New("expected error")
		multiSigner := &cryptoMocks.MultiSignerStub{
			AggregateSigsCalled: func(pubKeysSigners, signatures [][]byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSigningHandler(args)

		for i := 0; i < len(args.PubKeys); i++ {
			_ = signer.StoreSignatureShare(uint16(i), []byte("sigShare"))
		}

		aggSig, err := signer.AggregateSigs(bitmap, epoch)
		require.Nil(t, aggSig)
		require.Equal(t, expectedErr, err)
	})
	t.Run("failed to get current multi signer", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSigningHandler()

		expectedErr := errors.New("expected error")
		args.MultiSignerContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return nil, expectedErr
			},
		}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		aggSig, err := signer.AggregateSigs(bitmap, epoch)
		require.Nil(t, aggSig)
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		expectedAggSig := []byte("agg sig")
		multiSigner := &cryptoMocks.MultiSignerStub{
			AggregateSigsCalled: func(pubKeysSigners, signatures [][]byte) ([]byte, error) {
				require.Equal(t, len(args.PubKeys)-1, len(pubKeysSigners))
				require.Equal(t, len(args.PubKeys)-1, len(signatures))
				return expectedAggSig, nil
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSigningHandler(args)

		for i := 0; i < len(args.PubKeys); i++ {
			_ = signer.StoreSignatureShare(uint16(i), []byte("sigShare"))
		}

		aggSig, err := signer.AggregateSigs(bitmap, epoch)
		require.Nil(t, err)
		require.Equal(t, expectedAggSig, aggSig)
	})
}

func TestSigningHandler_Verify(t *testing.T) {
	t.Parallel()

	message := []byte("message")
	epoch := uint32(0)

	t.Run("verify agg sig should fail", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		err := signer.Verify(message, nil, epoch)
		require.Equal(t, cryptoFactory.ErrNilBitmap, err)
	})
	t.Run("bitmap mismatch", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4",
			"pk5", "pk6", "pk7", "pk8", "pk9"}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		err := signer.Verify(message, bitmap, epoch)
		require.Equal(t, cryptoFactory.ErrBitmapMismatch, err)
	})
	t.Run("verify agg sig should fail", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		expectedErr := errors.New("expected error")
		multiSigner := &cryptoMocks.MultiSignerStub{
			VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message, aggSig []byte) error {
				return expectedErr
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSigningHandler(args)

		err := signer.Verify(message, bitmap, epoch)
		require.Equal(t, expectedErr, err)
	})
	t.Run("failed to get current multi signer", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSigningHandler()

		expectedErr := errors.New("expected error")
		args.MultiSignerContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return nil, expectedErr
			},
		}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		err := signer.Verify(message, bitmap, epoch)
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSigningHandler()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		expAggSig := []byte("aggSig")

		multiSigner := &cryptoMocks.MultiSignerStub{
			VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message, aggSig []byte) error {
				require.Equal(t, len(args.PubKeys)-1, len(pubKeysSigners))
				require.Equal(t, expAggSig, aggSig)
				return nil
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSigningHandler(args)

		_ = signer.SetAggregatedSig(expAggSig)

		err := signer.Verify(message, bitmap, epoch)
		require.Nil(t, err)
	})
}

func TestSigningHandler_CreateSignatureForPublicKey(t *testing.T) {
	t.Parallel()

	args := createMockArgsSigningHandler()
	getHandledPrivateKeyCalled := false
	pkBytes := []byte("public key bytes")

	expectedSigShare := []byte("sigShare")
	args.KeysHandler = &testscommon.KeysHandlerStub{
		GetHandledPrivateKeyCalled: func(providedPkBytes []byte) crypto.PrivateKey {
			assert.Equal(t, pkBytes, providedPkBytes)
			getHandledPrivateKeyCalled = true

			return &cryptoMocks.PrivateKeyStub{}
		},
	}
	args.SingleSigner = &cryptoMocks.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return expectedSigShare, nil
		},
	}

	signer, _ := cryptoFactory.NewSigningHandler(args)
	sigShare, err := signer.CreateSignatureForPublicKey([]byte("msg1"), pkBytes)
	require.Nil(t, err)
	require.Equal(t, expectedSigShare, sigShare)
	assert.True(t, getHandledPrivateKeyCalled)
}

func TestSigningHandler_VerifySingleSignature(t *testing.T) {
	t.Parallel()

	t.Run("not a valid public key should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsSigningHandler()
		args.KeyGenerator = &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				return nil, expectedErr
			},
		}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		err := signer.VerifySingleSignature([]byte("pk"), []byte("msg"), []byte("sig"))
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedPkBytes := []byte("pk")
		providedMsg := []byte("msg")
		providedSig := []byte("sig")
		pk := &cryptoMocks.PublicKeyStub{}

		verifyCalled := false
		args := createMockArgsSigningHandler()
		args.KeyGenerator = &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				assert.Equal(t, providedPkBytes, b)
				return pk, nil
			},
		}
		args.SingleSigner = &cryptoMocks.SingleSignerStub{
			VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				assert.Equal(t, pk, public)
				assert.Equal(t, providedMsg, msg)
				assert.Equal(t, providedSig, sig)
				verifyCalled = true

				return nil
			},
		}

		signer, _ := cryptoFactory.NewSigningHandler(args)

		err := signer.VerifySingleSignature(providedPkBytes, providedMsg, providedSig)
		assert.Nil(t, err)
		assert.True(t, verifyCalled)
	})
}

func TestSigningHandler_ShallowClone(t *testing.T) {
	t.Parallel()

	args := createMockArgsSigningHandler()
	args.PubKeys = []string{"pk1"}

	verifySigCalled := false
	expectedSigShare := []byte("sigShare1")
	expectedMsg := []byte("msg")

	args.MultiSignerContainer = &cryptoMocks.MultiSignerContainerMock{
		MultiSigner: &cryptoMocks.MultisignerMock{
			VerifySignatureShareCalled: func(publicKey []byte, message []byte, sig []byte) error {
				require.Equal(t, []byte("pk1"), publicKey)
				require.Equal(t, expectedSigShare, sig)
				require.Equal(t, expectedMsg, message)

				verifySigCalled = true
				return nil
			},
		},
	}

	signer, _ := cryptoFactory.NewSigningHandler(args)
	clone := signer.ShallowClone()
	require.False(t, clone.IsInterfaceNil())

	err := clone.StoreSignatureShare(0, expectedSigShare)
	require.Nil(t, err)

	sigShare, err := clone.SignatureShare(0)
	require.Nil(t, err)
	require.Equal(t, expectedSigShare, sigShare)

	err = clone.VerifySignatureShare(0, expectedSigShare, expectedMsg, 0)
	require.Nil(t, err)
	require.True(t, verifySigCalled)
}
