package crypto_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	cryptoFactory "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/require"
)

func createMockArgsSignatureHolder() cryptoFactory.ArgsSignatureHolder {
	return cryptoFactory.ArgsSignatureHolder{
		PubKeys:              []string{"pubkey1"},
		PrivKeyBytes:         []byte("privKey"),
		MultiSignerContainer: &cryptoMocks.MultiSignerContainerMock{},
		KeyGenerator:         &cryptoMocks.KeyGenStub{},
	}
}

func TestNewSigner(t *testing.T) {
	t.Parallel()

	t.Run("nil multi signer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.MultiSignerContainer = nil

		signer, err := cryptoFactory.NewSignatureHolder(args)
		require.Nil(t, signer)
		require.Equal(t, cryptoFactory.ErrNilMultiSignerContainer, err)
	})

	t.Run("nil key generator", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.KeyGenerator = nil

		signer, err := cryptoFactory.NewSignatureHolder(args)
		require.Nil(t, signer)
		require.Equal(t, cryptoFactory.ErrNilKeyGenerator, err)
	})

	t.Run("nil private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PrivKeyBytes = nil

		signer, err := cryptoFactory.NewSignatureHolder(args)
		require.Nil(t, signer)
		require.Equal(t, cryptoFactory.ErrNoPrivateKeySet, err)
	})

	t.Run("no public keys", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{}

		signer, err := cryptoFactory.NewSignatureHolder(args)
		require.Nil(t, signer)
		require.Equal(t, cryptoFactory.ErrNoPublicKeySet, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		signer, err := cryptoFactory.NewSignatureHolder(args)
		require.Nil(t, err)
		require.False(t, check.IfNil(signer))
	})
}

func TestSignatureHolder_Create(t *testing.T) {
	t.Parallel()

	t.Run("empty pubkeys in list", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, err := cryptoFactory.NewSignatureHolder(args)
		require.Nil(t, err)
		require.NotNil(t, signer)

		pubKeys := []string{"pubKey1", ""}
		createdSigner, err := signer.Create(pubKeys)
		require.Nil(t, createdSigner)
		require.Equal(t, cryptoFactory.ErrEmptyPubKeyString, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, err := cryptoFactory.NewSignatureHolder(args)
		require.Nil(t, err)
		require.NotNil(t, signer)

		pubKeys := []string{"pubKey1", "pubKey2"}
		createdSigner, err := signer.Create(pubKeys)
		require.Nil(t, err)
		require.NotNil(t, createdSigner)
	})
}

func TestSignatureHolder_Reset(t *testing.T) {
	t.Parallel()

	t.Run("nil public keys", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, _ := cryptoFactory.NewSignatureHolder(args)
		err := signer.Reset(nil)
		require.Equal(t, cryptoFactory.ErrNilPublicKeys, err)
	})

	t.Run("empty pubkeys in list", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, _ := cryptoFactory.NewSignatureHolder(args)
		err := signer.Reset([]string{"pubKey1", ""})
		require.Equal(t, cryptoFactory.ErrEmptyPubKeyString, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, _ := cryptoFactory.NewSignatureHolder(args)
		err := signer.Reset([]string{"pubKey1", "pubKey2"})
		require.Nil(t, err)
	})
}

func TestSignatureHolder_CreateSignatureShare(t *testing.T) {
	t.Parallel()

	selfIndex := uint16(0)
	epoch := uint32(0)

	t.Run("nil message", func(t *testing.T) {
		t.Parallel()

		signer, _ := cryptoFactory.NewSignatureHolder(createMockArgsSignatureHolder())
		sigShare, err := signer.CreateSignatureShare(nil, selfIndex, epoch)
		require.Nil(t, sigShare)
		require.Equal(t, cryptoFactory.ErrNilMessage, err)
	})

	t.Run("create sig share failed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		expectedErr := errors.New("expected error")
		multiSigner := &cryptoMocks.MultiSignerStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSignatureHolder(args)
		sigShare, err := signer.CreateSignatureShare([]byte("msg1"), selfIndex, epoch)
		require.Nil(t, sigShare)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		expectedSigShare := []byte("sigShare")
		multiSigner := &cryptoMocks.MultiSignerStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return expectedSigShare, nil
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSignatureHolder(args)
		sigShare, err := signer.CreateSignatureShare([]byte("msg1"), selfIndex, epoch)
		require.Nil(t, err)
		require.Equal(t, expectedSigShare, sigShare)
	})
}

func TestSignatureHolder_VerifySignatureShare(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(1)
	epoch := uint32(0)
	msg := []byte("message")

	t.Run("invalid signature share", func(t *testing.T) {
		t.Parallel()

		signer, _ := cryptoFactory.NewSignatureHolder(createMockArgsSignatureHolder())
		err := signer.VerifySignatureShare(ownIndex, nil, msg, epoch)
		require.Equal(t, cryptoFactory.ErrInvalidSignature, err)
	})

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		signer, _ := cryptoFactory.NewSignatureHolder(createMockArgsSignatureHolder())
		err := signer.VerifySignatureShare(uint16(3), []byte("sigShare"), msg, epoch)
		require.Equal(t, cryptoFactory.ErrIndexOutOfBounds, err)
	})

	t.Run("signature share verification failed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2"}

		expectedErr := errors.New("expected error")
		multiSigner := &cryptoMocks.MultiSignerStub{
			VerifySignatureShareCalled: func(publicKey, message, sig []byte) error {
				return expectedErr
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), msg, epoch)
		require.Equal(t, expectedErr, err)
	})

	t.Run("failed to get current multi signer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2"}

		expectedErr := errors.New("expected error")
		args.MultiSignerContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return nil, expectedErr
			},
		}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), msg, epoch)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2"}

		multiSigner := &cryptoMocks.MultiSignerStub{
			VerifySignatureShareCalled: func(publicKey, message, sig []byte) error {
				return nil
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), msg, epoch)
		require.Nil(t, err)
	})
}

func TestSignatureHolder_StoreSignatureShare(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(2)
	epoch := uint32(0)
	msg := []byte("message")

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		signer, _ := cryptoFactory.NewSignatureHolder(createMockArgsSignatureHolder())
		err := signer.StoreSignatureShare(uint16(2), []byte("sigShare"))
		require.Equal(t, cryptoFactory.ErrIndexOutOfBounds, err)
	})

	t.Run("failed to get current multi signer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		expectedErr := errors.New("expected error")
		args.MultiSignerContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return nil, expectedErr
			},
		}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		sigShare, err := signer.CreateSignatureShare(msg, uint16(0), epoch)
		require.Nil(t, sigShare)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		multiSigner := &cryptoMocks.MultiSignerStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return []byte("sigshare"), nil
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		sigShare, err := signer.CreateSignatureShare(msg, uint16(0), epoch)
		require.Nil(t, err)

		err = signer.StoreSignatureShare(ownIndex, sigShare)
		require.Nil(t, err)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, err)
		require.Equal(t, sigShare, sigShareRead)
	})
}

func TestSignatureHolder_SignatureShare(t *testing.T) {
	t.Parallel()

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		index := uint16(1)
		sigShare := []byte("sig share")

		args := createMockArgsSignatureHolder()

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		_ = signer.StoreSignatureShare(index, sigShare)

		sigShareRead, err := signer.SignatureShare(index)
		require.Nil(t, sigShareRead)
		require.Equal(t, cryptoFactory.ErrIndexOutOfBounds, err)
	})

	t.Run("nil element at index", func(t *testing.T) {
		t.Parallel()

		ownIndex := uint16(1)

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		_ = signer.StoreSignatureShare(ownIndex, nil)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, sigShareRead)
		require.Equal(t, cryptoFactory.ErrNilElement, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ownIndex := uint16(1)
		sigShare := []byte("sig share")

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		_ = signer.StoreSignatureShare(ownIndex, sigShare)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, err)
		require.Equal(t, sigShare, sigShareRead)
	})
}

func TestSignatureHolder_AggregateSigs(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)

	t.Run("nil bitmap", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		aggSig, err := signer.AggregateSigs(nil, epoch)
		require.Nil(t, aggSig)
		require.Equal(t, cryptoFactory.ErrNilBitmap, err)
	})

	t.Run("bitmap mismatch", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4",
			"pk5", "pk6", "pk7", "pk8", "pk9"}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		aggSig, err := signer.AggregateSigs(bitmap, epoch)
		require.Nil(t, aggSig)
		require.Equal(t, cryptoFactory.ErrBitmapMismatch, err)
	})

	t.Run("failed to get aggregated sig", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		expectedErr := errors.New("expected error")
		multiSigner := &cryptoMocks.MultiSignerStub{
			AggregateSigsCalled: func(pubKeysSigners, signatures [][]byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSignatureHolder(args)

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

		args := createMockArgsSignatureHolder()

		expectedErr := errors.New("expected error")
		args.MultiSignerContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return nil, expectedErr
			},
		}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		aggSig, err := signer.AggregateSigs(bitmap, epoch)
		require.Nil(t, aggSig)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSignatureHolder()
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

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		for i := 0; i < len(args.PubKeys); i++ {
			_ = signer.StoreSignatureShare(uint16(i), []byte("sigShare"))
		}

		aggSig, err := signer.AggregateSigs(bitmap, epoch)
		require.Nil(t, err)
		require.Equal(t, expectedAggSig, aggSig)
	})
}

func TestSignatureHolder_Verify(t *testing.T) {
	t.Parallel()

	message := []byte("message")
	epoch := uint32(0)

	t.Run("verify agg sig should fail", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		err := signer.Verify(message, nil, epoch)
		require.Equal(t, cryptoFactory.ErrNilBitmap, err)
	})

	t.Run("bitmap mismatch", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4",
			"pk5", "pk6", "pk7", "pk8", "pk9"}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		err := signer.Verify(message, bitmap, epoch)
		require.Equal(t, cryptoFactory.ErrBitmapMismatch, err)
	})

	t.Run("verify agg sig should fail", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		expectedErr := errors.New("expected error")
		multiSigner := &cryptoMocks.MultiSignerStub{
			VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message, aggSig []byte) error {
				return expectedErr
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		err := signer.Verify(message, bitmap, epoch)
		require.Equal(t, expectedErr, err)
	})

	t.Run("failed to get current multi signer", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSignatureHolder()

		expectedErr := errors.New("expected error")
		args.MultiSignerContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return nil, expectedErr
			},
		}

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		err := signer.Verify(message, bitmap, epoch)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSignatureHolder()
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

		signer, _ := cryptoFactory.NewSignatureHolder(args)

		_ = signer.SetAggregatedSig(expAggSig)

		err := signer.Verify(message, bitmap, epoch)
		require.Nil(t, err)
	})
}
