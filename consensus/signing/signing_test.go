package signing_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/signing"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/require"
)

func createArgsSignerMock() signing.ArgsSinger {
	return signing.ArgsSinger{
		PubKeys:      []string{"pubkey1"},
		OwnIndex:     uint16(0),
		PrivKey:      &cryptoMocks.PrivateKeyStub{},
		SingleSigner: &cryptoMocks.SingleSignerStub{},
		MultiSigner:  &cryptoMocks.MultiSignerNewStub{},
		KeyGenerator: &cryptoMocks.KeyGenStub{},
	}
}

func TestNewSigner(t *testing.T) {
	t.Parallel()

	t.Run("nil single signer", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		args.SingleSigner = nil

		signer, err := signing.NewSigner(args)
		require.Nil(t, signer)
		require.Equal(t, signing.ErrNilSingleSigner, err)
	})

	t.Run("nil multi signer", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		args.MultiSigner = nil

		signer, err := signing.NewSigner(args)
		require.Nil(t, signer)
		require.Equal(t, signing.ErrNilMultiSigner, err)
	})

	t.Run("nil key generator", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		args.KeyGenerator = nil

		signer, err := signing.NewSigner(args)
		require.Nil(t, signer)
		require.Equal(t, signing.ErrNilKeyGenerator, err)
	})

	t.Run("nil private key", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		args.PrivKey = nil

		signer, err := signing.NewSigner(args)
		require.Nil(t, signer)
		require.Equal(t, signing.ErrNilPrivateKey, err)
	})

	t.Run("no public keys", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		args.PubKeys = []string{}

		signer, err := signing.NewSigner(args)
		require.Nil(t, signer)
		require.Equal(t, signing.ErrNoPublicKeySet, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		signer, err := signing.NewSigner(args)
		require.Nil(t, err)
		require.NotNil(t, signer)
	})
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("empty pubkeys in list", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()

		signer, err := signing.NewSigner(args)
		require.Nil(t, err)
		require.NotNil(t, signer)

		pubKeys := []string{"pubKey1", ""}
		createdSigner, err := signer.Create(pubKeys, uint16(2))
		require.Nil(t, createdSigner)
		require.Equal(t, signing.ErrEmptyPubKeyString, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()

		signer, err := signing.NewSigner(args)
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

		args := createArgsSignerMock()

		signer, _ := signing.NewSigner(args)
		err := signer.Reset(nil, uint16(3))
		require.Equal(t, signing.ErrNilPublicKeys, err)
	})

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()

		signer, _ := signing.NewSigner(args)
		err := signer.Reset([]string{"pubKey1", "pubKey2"}, uint16(3))
		require.Equal(t, signing.ErrIndexOutOfBounds, err)
	})

	t.Run("empty pubkeys in list", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()

		signer, _ := signing.NewSigner(args)
		err := signer.Reset([]string{"pubKey1", ""}, uint16(1))
		require.Equal(t, signing.ErrEmptyPubKeyString, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()

		signer, _ := signing.NewSigner(args)
		err := signer.Reset([]string{"pubKey1", "pubKey2"}, uint16(1))
		require.Nil(t, err)
	})
}

func TestCreateSignatureShare(t *testing.T) {
	t.Parallel()

	t.Run("nil message", func(t *testing.T) {
		t.Parallel()

		signer, _ := signing.NewSigner(createArgsSignerMock())
		sigShare, err := signer.CreateSignatureShare(nil, []byte{})
		require.Nil(t, sigShare)
		require.Equal(t, signing.ErrNilMessage, err)
	})

	t.Run("create sig share failed", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()

		expectedErr := errors.New("expected error")
		args.MultiSigner = &cryptoMocks.MultiSignerNewStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}

		signer, _ := signing.NewSigner(args)
		sigShare, err := signer.CreateSignatureShare([]byte("msg1"), []byte{})
		require.Nil(t, sigShare)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()

		expectedSigShare := []byte("sigShare")
		args.MultiSigner = &cryptoMocks.MultiSignerNewStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return expectedSigShare, nil
			},
		}
		signer, _ := signing.NewSigner(args)
		sigShare, err := signer.CreateSignatureShare([]byte("msg1"), []byte{})
		require.Nil(t, err)
		require.Equal(t, expectedSigShare, sigShare)
	})
}

func TestVerifySignatureShare(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(1)
	msg := []byte("message")

	t.Run("nil signature share", func(t *testing.T) {
		t.Parallel()

		signer, _ := signing.NewSigner(createArgsSignerMock())
		err := signer.VerifySignatureShare(ownIndex, nil, msg, []byte(""))
		require.Equal(t, signing.ErrNilSignature, err)
	})

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		signer, _ := signing.NewSigner(createArgsSignerMock())
		err := signer.VerifySignatureShare(uint16(3), []byte("sigShare"), msg, []byte(""))
		require.Equal(t, signing.ErrIndexOutOfBounds, err)
	})

	t.Run("signature share verification failed", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		args.PubKeys = []string{"pk1", "pk2"}

		expectedErr := errors.New("expected error")
		args.MultiSigner = &cryptoMocks.MultiSignerNewStub{
			VerifySignatureShareCalled: func(publicKey, message, sig []byte) error {
				return expectedErr
			},
		}
		signer, _ := signing.NewSigner(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), []byte{}, msg)
		require.Equal(t, expectedErr, err)
	})

	t.Run("signature share verification failed", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		args.PubKeys = []string{"pk1", "pk2"}

		args.MultiSigner = &cryptoMocks.MultiSignerNewStub{
			VerifySignatureShareCalled: func(publicKey, message, sig []byte) error {
				return nil
			},
		}
		signer, _ := signing.NewSigner(args)

		err := signer.VerifySignatureShare(uint16(1), []byte("sigShare"), []byte{}, msg)
		require.Nil(t, err)
	})
}

func TestStoreSignatureShare(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(2)
	msg := []byte("message")

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		signer, _ := signing.NewSigner(createArgsSignerMock())
		err := signer.StoreSignatureShare(uint16(2), []byte("sigShare"))
		require.Equal(t, signing.ErrIndexOutOfBounds, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}
		args.OwnIndex = ownIndex
		args.MultiSigner = &cryptoMocks.MultiSignerNewStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return []byte("sigshare"), nil
			},
		}

		signer, _ := signing.NewSigner(args)

		sigShare, err := signer.CreateSignatureShare(msg, []byte{})
		require.Nil(t, err)

		err = signer.StoreSignatureShare(ownIndex, sigShare)
		require.Nil(t, err)

		sigShareRead, err := signer.SignatureShare(ownIndex)
		require.Nil(t, err)
		require.Equal(t, sigShare, sigShareRead)
	})
}

func TestAggregateSigs(t *testing.T) {
	t.Parallel()

	bitmap := make([]byte, 1)
	bitmap[0] = 0x07

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgsSignerMock()
		args.PubKeys = []string{"pk1", "pk2", "pk3", "pk4"}

		expectedAggSig := []byte("agg sig")
		args.MultiSigner = &cryptoMocks.MultiSignerNewStub{
			CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
				return []byte("sigshare"), nil
			},
			AggregateSigsCalled: func(pubKeysSigners, signatures [][]byte) ([]byte, error) {
				require.Equal(t, len(args.PubKeys)-1, len(pubKeysSigners))
				require.Equal(t, len(args.PubKeys)-1, len(signatures))
				return expectedAggSig, nil
			},
		}

		signer, _ := signing.NewSigner(args)

		for i := 0; i < len(args.PubKeys); i++ {
			_ = signer.StoreSignatureShare(uint16(i), []byte("sigShare"))
		}

		aggSig, err := signer.AggregateSigs(bitmap)
		require.Nil(t, err)
		require.Equal(t, expectedAggSig, aggSig)
	})
}
