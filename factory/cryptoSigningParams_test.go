package factory

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCryptoSigningParamsFactory_NilPubKeyConverterShoulldErr(t *testing.T) {
	t.Parallel()

	cspf, err := NewCryptoSigningParamsFactory(nil, 0, "name", &mock.SuiteStub{})
	require.Nil(t, cspf)
	require.Equal(t, ErrNilPubKeyConverter, err)
}

func TestNewCryptoSigningParamsFactory_NilSuiteShouldErr(t *testing.T) {
	t.Parallel()

	cspf, err := NewCryptoSigningParamsFactory(&mock.PubkeyConverterStub{}, 0, "name", nil)
	require.Nil(t, cspf)
	require.Equal(t, ErrNilSuite, err)
}

func TestNewCryptoSigningParamsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cspf, err := NewCryptoSigningParamsFactory(&mock.PubkeyConverterStub{}, 0, "name", &mock.SuiteStub{})
	require.NoError(t, err)
	require.NotNil(t, cspf)
}

func TestCryptoSigningParamsFactory_Create_GetSkPkErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error while getting the sk and pk")
	cspf, _ := NewCryptoSigningParamsFactory(&mock.PubkeyConverterStub{}, 0, "name", &mock.SuiteStub{})

	cspf.SetSkPkProviderHandler(func() ([]byte, []byte, error) {
		return nil, nil, expectedErr
	})
	cp, err := cspf.Create()
	require.Equal(t, expectedErr, err)
	require.Nil(t, cp)
}

func TestCryptoSigningParamsFactory_Create_PubKeyMissmatchShouldErr(t *testing.T) {
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
	cspf, _ := NewCryptoSigningParamsFactory(&mock.PubkeyConverterStub{}, 0, "name", suite)

	cspf.SetSkPkProviderHandler(func() ([]byte, []byte, error) {
		return []byte("sk"), diffPubkey2, nil
	})
	cp, err := cspf.Create()
	require.Equal(t, ErrPublicKeyMismatch, err)
	require.Nil(t, cp)
}

func TestCryptoSigningParamsFactory_CreateShouldWork(t *testing.T) {
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
	cspf, _ := NewCryptoSigningParamsFactory(&mock.PubkeyConverterStub{}, 0, "name", suite)

	cspf.SetSkPkProviderHandler(func() ([]byte, []byte, error) {
		return []byte("sk"), pubKey, nil
	})
	cp, err := cspf.Create()
	require.NoError(t, err)
	require.NotNil(t, cp)
}

func TestCryptoSigningParamsFactory_GetSkPk_PathNotFound(t *testing.T) {
	t.Parallel()

	cspf, _ := NewCryptoSigningParamsFactory(&mock.PubkeyConverterStub{}, 0, "name", &mock.SuiteStub{})
	sk, pk, err := cspf.GetSkPk()
	require.Error(t, err)
	require.Nil(t, sk)
	require.Nil(t, pk)
}
