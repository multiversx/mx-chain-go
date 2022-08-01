package signing_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/consensus/signing"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/require"
)

func createMockArgsSignatureHolder() signing.ArgsSignatureHolder {
	return signing.ArgsSignatureHolder{
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

		signer, err := signing.NewSignatureHolder(args)
		require.Nil(t, signer)
		require.Equal(t, signing.ErrNilMultiSignerContainer, err)
	})

	t.Run("nil key generator", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.KeyGenerator = nil

		signer, err := signing.NewSignatureHolder(args)
		require.Nil(t, signer)
		require.Equal(t, signing.ErrNilKeyGenerator, err)
	})

	t.Run("nil private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PrivKeyBytes = nil

		signer, err := signing.NewSignatureHolder(args)
		require.Nil(t, signer)
		require.Equal(t, signing.ErrNoPrivateKeySet, err)
	})

	t.Run("no public keys", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{}

		signer, err := signing.NewSignatureHolder(args)
		require.Nil(t, signer)
		require.Equal(t, signing.ErrNoPublicKeySet, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		signer, err := signing.NewSignatureHolder(args)
		require.Nil(t, err)
		require.False(t, check.IfNil(signer))
	})
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("empty pubkeys in list", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, err := signing.NewSignatureHolder(args)
		require.Nil(t, err)
		require.NotNil(t, signer)

		pubKeys := []string{"pubKey1", ""}
		createdSigner, err := signer.Create(pubKeys, uint16(2))
		require.Nil(t, createdSigner)
		require.Equal(t, signing.ErrEmptyPubKeyString, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, err := signing.NewSignatureHolder(args)
		require.Nil(t, err)
		require.NotNil(t, signer)

		pubKeys := []string{"pubKey1", "pubKey2"}
		createdSigner, err := signer.Create(pubKeys, uint16(2))
		require.Nil(t, err)
		require.NotNil(t, createdSigner)
	})
}

func TestReset(t *testing.T) {
	t.Parallel()

	t.Run("nil public keys", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, _ := signing.NewSignatureHolder(args)
		err := signer.Reset(nil)
		require.Equal(t, signing.ErrNilPublicKeys, err)
	})

	t.Run("empty pubkeys in list", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, _ := signing.NewSignatureHolder(args)
		err := signer.Reset([]string{"pubKey1", ""})
		require.Equal(t, signing.ErrEmptyPubKeyString, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()

		signer, _ := signing.NewSignatureHolder(args)
		err := signer.Reset([]string{"pubKey1", "pubKey2"})
		require.Nil(t, err)
	})
}

func TestCreateSignatureShare(t *testing.T) {
	t.Parallel()

	selfIndex := uint16(0)
	epoch := uint32(0)

	t.Run("nil message", func(t *testing.T) {
		t.Parallel()

		signer, _ := signing.NewSignatureHolder(createMockArgsSignatureHolder())
		sigShare, err := signer.CreateSignatureShare(nil, selfIndex, epoch)
		require.Nil(t, sigShare)
		require.Equal(t, signing.ErrNilMessage, err)
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

		signer, _ := signing.NewSignatureHolder(args)
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

		signer, _ := signing.NewSignatureHolder(args)
		sigShare, err := signer.CreateSignatureShare([]byte("msg1"), selfIndex, epoch)
		require.Nil(t, err)
		require.Equal(t, expectedSigShare, sigShare)
	})
}

func TestVerifySignatureShare(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(1)
	epoch := uint32(0)
	msg := []byte("message")

	t.Run("invalid signature share", func(t *testing.T) {
		t.Parallel()

		signer, _ := signing.NewSignatureHolder(createMockArgsSignatureHolder())
		err := signer.VerifySignatureShare(ownIndex, nil, msg, epoch)
		require.Equal(t, signing.ErrInvalidSignature, err)
	})

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		signer, _ := signing.NewSignatureHolder(createMockArgsSignatureHolder())
		err := signer.VerifySignatureShare(uint16(3), []byte("sigShare"), msg, epoch)
		require.Equal(t, signing.ErrIndexOutOfBounds, err)
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

		signer, _ := signing.NewSignatureHolder(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), msg, epoch)
		require.Equal(t, expectedErr, err)
	})

	t.Run("signature share verification failed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2"}

		multiSigner := &cryptoMocks.MultiSignerStub{
			VerifySignatureShareCalled: func(publicKey, message, sig []byte) error {
				return nil
			},
		}
		args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigner)

		signer, _ := signing.NewSignatureHolder(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), msg, epoch)
		require.Nil(t, err)
	})
}

func TestStoreSignatureShare(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(2)
	epoch := uint32(0)
	msg := []byte("message")

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		signer, _ := signing.NewSignatureHolder(createMockArgsSignatureHolder())
		err := signer.StoreSignatureShare(uint16(2), []byte("sigShare"))
		require.Equal(t, signing.ErrIndexOutOfBounds, err)
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

		signer, _ := signing.NewSignatureHolder(args)

		sigShare, err := signer.CreateSignatureShare(msg, uint16(0), epoch)
		require.Nil(t, err)

		err = signer.StoreSignatureShare(ownIndex, sigShare)
		require.Nil(t, err)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, err)
		require.Equal(t, sigShare, sigShareRead)
	})
}

func TestSignatureShare(t *testing.T) {
	t.Parallel()

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		index := uint16(1)
		sigShare := []byte("sig share")

		args := createMockArgsSignatureHolder()

		signer, _ := signing.NewSignatureHolder(args)

		_ = signer.StoreSignatureShare(index, sigShare)

		sigShareRead, err := signer.SignatureShare(index)
		require.Nil(t, sigShareRead)
		require.Equal(t, signing.ErrIndexOutOfBounds, err)
	})

	t.Run("nil element at index", func(t *testing.T) {
		t.Parallel()

		ownIndex := uint16(1)

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := signing.NewSignatureHolder(args)

		_ = signer.StoreSignatureShare(ownIndex, nil)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, sigShareRead)
		require.Equal(t, signing.ErrNilElement, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ownIndex := uint16(1)
		sigShare := []byte("sig share")

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := signing.NewSignatureHolder(args)

		_ = signer.StoreSignatureShare(ownIndex, sigShare)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, err)
		require.Equal(t, sigShare, sigShareRead)
	})
}

func TestAggregateSigs(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)

	t.Run("nil bitmap", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := signing.NewSignatureHolder(args)

		aggSig, err := signer.AggregateSigs(nil, epoch)
		require.Nil(t, aggSig)
		require.Equal(t, signing.ErrNilBitmap, err)
	})

	t.Run("bitmap mismatch", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4",
			"pk5", "pk6", "pk7", "pk8", "pk9"}

		signer, _ := signing.NewSignatureHolder(args)

		aggSig, err := signer.AggregateSigs(bitmap, epoch)
		require.Nil(t, aggSig)
		require.Equal(t, signing.ErrBitmapMismatch, err)
	})

	t.Run("should work", func(t *testing.T) {
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

		signer, _ := signing.NewSignatureHolder(args)

		for i := 0; i < len(args.PubKeys); i++ {
			_ = signer.StoreSignatureShare(uint16(i), []byte("sigShare"))
		}

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

		signer, _ := signing.NewSignatureHolder(args)

		for i := 0; i < len(args.PubKeys); i++ {
			_ = signer.StoreSignatureShare(uint16(i), []byte("sigShare"))
		}

		aggSig, err := signer.AggregateSigs(bitmap, epoch)
		require.Nil(t, err)
		require.Equal(t, expectedAggSig, aggSig)
	})
}

func TestVerify(t *testing.T) {
	t.Parallel()

	message := []byte("message")
	epoch := uint32(0)

	t.Run("verify agg sig should fail", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		signer, _ := signing.NewSignatureHolder(args)

		err := signer.Verify(message, nil, epoch)
		require.Equal(t, signing.ErrNilBitmap, err)
	})

	t.Run("bitmap mismatch", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, 1)
		bitmap[0] = 0x07

		args := createMockArgsSignatureHolder()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4",
			"pk5", "pk6", "pk7", "pk8", "pk9"}

		signer, _ := signing.NewSignatureHolder(args)

		err := signer.Verify(message, bitmap, epoch)
		require.Equal(t, signing.ErrBitmapMismatch, err)
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

		signer, _ := signing.NewSignatureHolder(args)

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

		signer, _ := signing.NewSignatureHolder(args)

		_ = signer.SetAggregatedSig(expAggSig)

		err := signer.Verify(message, bitmap, epoch)
		require.Nil(t, err)
	})
}
