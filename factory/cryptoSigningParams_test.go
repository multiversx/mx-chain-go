package factory

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCryptoSigningParamsLoader_NilPubKeyConverterShoulldErr(t *testing.T) {
	t.Parallel()

	cspf, err := NewCryptoSigningParamsLoader(nil, 0, "name", &mock.SuiteStub{}, false)
	require.Nil(t, cspf)
	require.Equal(t, ErrNilPubKeyConverter, err)
}

func TestNewCryptoSigningParamsLoader_NilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	cspf, err := NewCryptoSigningParamsLoader(&mock.PubkeyConverterStub{}, 0, "name", nil, false)
	require.Nil(t, cspf)
	require.Equal(t, ErrNilSuite, err)
}

func TestNewCryptoSigningParamsLoader_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cspf, err := NewCryptoSigningParamsLoader(&mock.PubkeyConverterStub{}, 0, "name", &mock.SuiteStub{}, false)
	require.NoError(t, err)
	require.NotNil(t, cspf)
}

func TestCryptoSigningParamsLoader_Create_GetSkPkErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error while getting the sk and pk")
	cspf, _ := NewCryptoSigningParamsLoader(&mock.PubkeyConverterStub{}, 0, "name", &mock.SuiteStub{}, false)

	cspf.SetSkPkProviderHandler(func() ([]byte, []byte, error) {
		return nil, nil, expectedErr
	})
	cp, err := cspf.Get()
	require.Equal(t, expectedErr, err)
	require.Nil(t, cp)
}

func TestCryptoSigningParamsLoader_Create_PubKeyMissmatchShouldErr(t *testing.T) {
	t.Parallel()

	diffPubKey1, diffPubkey2 := []byte("public key1"), []byte("public key2")
	suite := &mock.SuiteStub{
		CreatePointStub: func() crypto.Point {
			return nil
		},
		CreatePointForScalarStub: func(_ crypto.Scalar) (crypto.Point, error) {
			return &mock.PointMock{
				MarshalBinaryStub: func(_, _ int) ([]byte, error) {
					return diffPubKey1, nil
				},
			}, nil
		},
		CreateScalarStub: func() crypto.Scalar {
			return &mock.ScalarMock{
				UnmarshalBinaryStub: func(bytes []byte) (int, error) {
					return 2, nil
				},
			}
		},
	}
	cspf, _ := NewCryptoSigningParamsLoader(&mock.PubkeyConverterStub{}, 0, "name", suite, false)

	cspf.SetSkPkProviderHandler(func() ([]byte, []byte, error) {
		return []byte("sk"), diffPubkey2, nil
	})
	cp, err := cspf.Get()
	require.Equal(t, ErrPublicKeyMismatch, err)
	require.Nil(t, cp)
}

func TestCryptoSigningParamsLoader_CreateShouldWork(t *testing.T) {
	t.Parallel()

	pubKey := []byte("public key")
	suite := &mock.SuiteStub{
		CreatePointStub: func() crypto.Point {
			return nil
		},
		CreatePointForScalarStub: func(_ crypto.Scalar) (crypto.Point, error) {
			return &mock.PointMock{
				MarshalBinaryStub: func(_, _ int) ([]byte, error) {
					return pubKey, nil
				},
			}, nil
		},
		CreateScalarStub: func() crypto.Scalar {
			return &mock.ScalarMock{
				UnmarshalBinaryStub: func(bytes []byte) (int, error) {
					return 2, nil
				},
			}
		},
	}
	cspf, _ := NewCryptoSigningParamsLoader(&mock.PubkeyConverterStub{}, 0, "name", suite, false)

	cspf.SetSkPkProviderHandler(func() ([]byte, []byte, error) {
		return []byte("sk"), pubKey, nil
	})
	cp, err := cspf.Get()
	require.NoError(t, err)
	require.NotNil(t, cp)
}

func TestCryptoSigningParamsLoader_GetSkPk_PathNotFound(t *testing.T) {
	t.Parallel()

	cspf, _ := NewCryptoSigningParamsLoader(&mock.PubkeyConverterStub{}, 0, "name", &mock.SuiteStub{}, false)
	sk, pk, err := cspf.GetSkPk()
	require.Error(t, err)
	require.Nil(t, sk)
	require.Nil(t, pk)
}
