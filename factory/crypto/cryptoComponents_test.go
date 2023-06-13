package crypto_test

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-go/config"
	errErd "github.com/multiversx/mx-chain-go/errors"
	cryptoComp "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/factory/mock"
	integrationTestsMock "github.com/multiversx/mx-chain-go/integrationTests/mock"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCryptoComponentsFactory_NilCoreComponentsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCryptoArgs(nil)
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, errErd.ErrNilCoreComponents, err)
}

func TestNewCryptoComponentsFactory_NilValidatorPublicKeyConverterShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetCryptoArgs(nil)
	args.CoreComponentsHolder = &integrationTestsMock.CoreComponentsStub{
		ValidatorPubKeyConverterField: nil,
	}

	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, errErd.ErrNilPubKeyConverter, err)
}

func TestNewCryptoComponentsFactory_NilPemFileShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.ValidatorKeyPemFileName = ""
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, errErd.ErrNilPath, err)
}

func TestCryptoComponentsFactory_CreateCryptoParamsNilKeyLoaderShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.KeyLoader = nil
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)

	require.Nil(t, ccf)
	require.Equal(t, errErd.ErrNilKeyLoader, err)
}

func TestNewCryptoComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestNewCryptoComponentsFactory_DisabledSigShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.ImportModeNoSigCheck = true
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestNewCryptoComponentsFactory_CreateInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "invalid"
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.Equal(t, err, errErd.ErrInvalidConsensusConfig)
}

func TestCryptoComponentsFactory_CreateShouldErrDueToMissingConfig(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.Config = config.Config{
		ValidatorPubkeyConverter: config.PubkeyConfig{
			Length:          8,
			Type:            "hex",
			SignatureLength: 48,
		}}

	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.Nil(t, err)

	cc, err := ccf.Create()
	require.Error(t, err)
	require.Nil(t, cc)
}

func TestCryptoComponentsFactory_CreateInvalidMultiSigHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.Config.MultisigHasher.Type = "invalid"
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.Nil(t, err)

	cspf, err := ccf.Create()
	require.Nil(t, cspf)
	require.Equal(t, errErd.ErrMultiSigHasherMissmatch, err)
}

func TestCryptoComponentsFactory_CreateOK(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
	assert.Equal(t, 0, len(cc.GetManagedPeersHolder().GetManagedKeysByCurrentNode()))
	assert.Nil(t, cc.Close())
}

func TestCryptoComponentsFactory_CreateWithDisabledSig(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.IsInImportMode = true
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
	assert.Equal(t, 0, len(cc.GetManagedPeersHolder().GetManagedKeysByCurrentNode()))
}

func TestCryptoComponentsFactory_CreateWithAutoGenerateKey(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.NoKeyProvided = true
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
	assert.Equal(t, 0, len(cc.GetManagedPeersHolder().GetManagedKeysByCurrentNode()))
}

func TestCryptoComponentsFactory_CreateSingleSignerInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "invalid"
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	singleSigner, err := ccf.CreateSingleSigner(false)
	require.Nil(t, singleSigner)
	require.Equal(t, errErd.ErrInvalidConsensusConfig, err)
}

func TestCryptoComponentsFactory_CreateSingleSignerOK(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	singleSigner, err := ccf.CreateSingleSigner(false)
	require.Nil(t, err)
	require.NotNil(t, singleSigner)
}

func TestCryptoComponentsFactory_CreateMultiSignerInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "other"
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	multiSigner, err := ccf.CreateMultiSignerContainer(&mock.KeyGenMock{}, false)
	require.Nil(t, multiSigner)
	require.Equal(t, errErd.ErrInvalidConsensusConfig, err)
}

func TestCryptoComponentsFactory_CreateMultiSignerOK(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	suite, _ := ccf.GetSuite()
	blockSignKeyGen := signing.NewKeyGenerator(suite)

	multiSigner, err := ccf.CreateMultiSignerContainer(blockSignKeyGen, false)
	require.Nil(t, err)
	require.NotNil(t, multiSigner)
}

func TestCryptoComponentsFactory_GetSuiteInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.Config.Consensus.Type = ""
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	suite, err := ccf.GetSuite()
	require.Nil(t, suite)
	require.Equal(t, errErd.ErrInvalidConsensusConfig, err)
}

func TestCryptoComponentsFactory_GetSuiteOK(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "bls"
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	suite, err := ccf.GetSuite()
	require.Nil(t, err)
	require.NotNil(t, suite)
}

func TestCryptoComponentsFactory_CreateCryptoParamsInvalidPrivateKeyByteArrayShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.KeyLoader = &mock.KeyLoaderStub{LoadKeyCalled: componentsMock.DummyLoadSkPkFromPemFile([]byte{}, componentsMock.DummyPk, nil)}
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	suite, _ := ccf.GetSuite()
	blockSignKeyGen := signing.NewKeyGenerator(suite)

	cryptoParams, err := ccf.CreateCryptoParams(blockSignKeyGen)
	require.Nil(t, cryptoParams)
	require.Equal(t, crypto.ErrInvalidParam, err)
}

func TestCryptoComponentsFactory_CreateCryptoParamsLoadKeysFailShouldErr(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.KeyLoader = &mock.KeyLoaderStub{LoadKeyCalled: componentsMock.DummyLoadSkPkFromPemFile([]byte{}, "", expectedError)}
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	suite, _ := ccf.GetSuite()
	blockSignKeyGen := signing.NewKeyGenerator(suite)

	cryptoParams, err := ccf.CreateCryptoParams(blockSignKeyGen)
	require.Nil(t, cryptoParams)
	require.Equal(t, expectedError, err)
}

func TestCryptoComponentsFactory_CreateCryptoParamsOK(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	suite, _ := ccf.GetSuite()
	blockSignKeyGen := signing.NewKeyGenerator(suite)

	cryptoParams, err := ccf.CreateCryptoParams(blockSignKeyGen)
	require.Nil(t, err)
	require.NotNil(t, cryptoParams)
}

func TestCryptoComponentsFactory_GetSkPkInvalidSkBytesShouldErr(t *testing.T) {
	t.Parallel()

	setSk := []byte("zxwY")
	setPk := []byte(componentsMock.DummyPk)
	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.KeyLoader = &mock.KeyLoaderStub{LoadKeyCalled: componentsMock.DummyLoadSkPkFromPemFile(setSk, string(setPk), nil)}
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	sk, pk, err := ccf.GetSkPk()
	require.NotNil(t, err)
	require.Nil(t, sk)
	require.Nil(t, pk)
}

func TestCryptoComponentsFactory_GetSkPkInvalidPkBytesShouldErr(t *testing.T) {
	t.Parallel()

	setSk := []byte(componentsMock.DummySk)
	setPk := "0"

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.KeyLoader = &mock.KeyLoaderStub{LoadKeyCalled: componentsMock.DummyLoadSkPkFromPemFile(setSk, setPk, nil)}
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	sk, pk, err := ccf.GetSkPk()
	require.NotNil(t, err)
	require.Nil(t, sk)
	require.Nil(t, pk)
}

func TestCryptoComponentsFactory_GetSkPkOK(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.Nil(t, err)

	expectedSk, _ := hex.DecodeString(componentsMock.DummySk)
	expectedPk, _ := hex.DecodeString(componentsMock.DummyPk)

	sk, pk, err := ccf.GetSkPk()
	require.Nil(t, err)
	require.Equal(t, expectedSk, sk)
	require.Equal(t, expectedPk, pk)
}

func TestCryptoComponentsFactory_MultiKey(t *testing.T) {
	t.Parallel()

	t.Run("internal error, LoadAllKeys returns different lengths for private and public keys", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetCryptoArgs(coreComponents)

		privateKeys, publicKeys := createBLSPrivatePublicKeys()

		args.KeyLoader = &mock.KeyLoaderStub{
			LoadKeyCalled: func(relativePath string, skIndex int) ([]byte, string, error) {
				return privateKeys[0], publicKeys[0], nil
			},
			LoadAllKeysCalled: func(path string) ([][]byte, []string, error) {
				return privateKeys[2:], publicKeys[1:], nil
			},
		}

		ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
		require.Nil(t, err)

		cc, err := ccf.Create()
		assert.Nil(t, cc)
		assert.Contains(t, err.Error(), "key loading error for the allValidatorsKeys file: mismatch number of private and public keys")
	})
	t.Run("encoded private key can not be hex decoded", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetCryptoArgs(coreComponents)
		privateKeys, publicKeys := createBLSPrivatePublicKeys()

		args.KeyLoader = &mock.KeyLoaderStub{
			LoadKeyCalled: func(relativePath string, skIndex int) ([]byte, string, error) {
				return privateKeys[0], publicKeys[0], nil
			},
			LoadAllKeysCalled: func(path string) ([][]byte, []string, error) {
				return [][]byte{[]byte("not a hex")}, []string{"a"}, nil
			},
		}

		ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
		require.Nil(t, err)

		cc, err := ccf.Create()
		assert.Nil(t, cc)
		assert.Contains(t, err.Error(), "for encoded secret key, key index 0")
	})
	t.Run("encoded public key can not be hex decoded", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetCryptoArgs(coreComponents)
		privateKeys, publicKeys := createBLSPrivatePublicKeys()

		args.KeyLoader = &mock.KeyLoaderStub{
			LoadKeyCalled: func(relativePath string, skIndex int) ([]byte, string, error) {
				return privateKeys[0], publicKeys[0], nil
			},
			LoadAllKeysCalled: func(path string) ([][]byte, []string, error) {
				return [][]byte{[]byte("aa")}, []string{"not hex"}, nil
			},
		}

		ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
		require.Nil(t, err)

		cc, err := ccf.Create()
		assert.Nil(t, cc)
		assert.Contains(t, err.Error(), "for encoded public key not hex, key index 0")
	})
	t.Run("not a valid private key", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetCryptoArgs(coreComponents)
		privateKeys, publicKeys := createBLSPrivatePublicKeys()

		args.KeyLoader = &mock.KeyLoaderStub{
			LoadKeyCalled: func(relativePath string, skIndex int) ([]byte, string, error) {
				return privateKeys[0], publicKeys[0], nil
			},
			LoadAllKeysCalled: func(path string) ([][]byte, []string, error) {
				return [][]byte{[]byte("aa")}, []string{publicKeys[0]}, nil
			},
		}

		ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
		require.Nil(t, err)

		cc, err := ccf.Create()
		assert.Nil(t, cc)
		assert.Contains(t, err.Error(), "secret key, key index 0")
	})
	t.Run("wrong public string read from file", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetCryptoArgs(coreComponents)
		privateKeys, publicKeys := createBLSPrivatePublicKeys()

		args.KeyLoader = &mock.KeyLoaderStub{
			LoadKeyCalled: func(relativePath string, skIndex int) ([]byte, string, error) {
				return privateKeys[0], publicKeys[0], nil
			},
			LoadAllKeysCalled: func(path string) ([][]byte, []string, error) {
				return [][]byte{privateKeys[0]}, []string{publicKeys[1]}, nil
			},
		}

		ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
		require.Nil(t, err)

		cc, err := ccf.Create()
		assert.Nil(t, cc)
		assert.Contains(t, err.Error(), "public keys mismatch, read "+
			"ae33fdd47ca6eed4b4e33a87beb580a20e908898a88c1d91a8b376cc35e8240e5083696ba6a1eeaa4cf50431980c38086dd7acc535c7571fb952d5c025d27c422fca999eaeaa13451946504d2b0a0c5b08958da236a4877b08abbd8059218f05"+
			", generated ae33fdd47ca6eed4b4e33a87beb580a20e908898a88c1d91a8b376cc35e8240e5083696ba6a1eeaa4cf50431980c38086dd7acc535c7571fb952d5c025d27c422fca999eaeaa13451946504d2b0a0c5b08958da236a4877b08abbd8059218f05"+
			", key index 0")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetCryptoArgs(coreComponents)

		privateKeys, publicKeys := createBLSPrivatePublicKeys()

		args.KeyLoader = &mock.KeyLoaderStub{
			LoadKeyCalled: func(relativePath string, skIndex int) ([]byte, string, error) {
				return privateKeys[0], publicKeys[0], nil
			},
			LoadAllKeysCalled: func(path string) ([][]byte, []string, error) {
				return privateKeys[1:], publicKeys[1:], nil
			},
		}

		ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
		require.Nil(t, err)

		cc, err := ccf.Create()
		require.Nil(t, err)
		// should hold num(publicKeys) - 1 as the first was used as the main key of the node
		assert.Equal(t, len(publicKeys)-1, len(cc.GetManagedPeersHolder().GetManagedKeysByCurrentNode()))
		skBytes, _, _ := ccf.GetSkPk()
		assert.NotEqual(t, skBytes, privateKeys[0]) // should generate another private key and not use the one loaded with LoadKey call
	})
}

func createBLSPrivatePublicKeys() ([][]byte, []string) {
	privateKeys := [][]byte{
		[]byte("13508f73f4bac43014ca5cdf16903bed4dcfd60f74123346f933e1cd0042ca52"),
		[]byte("4195470da0224c832b8cb3227fdfa2431fac50efce332dadc2e970c0977f6d3b"),
		[]byte("a65c1bbef47b833c2c2262cc341ad34745a10034b68e117e60c3391ed3503b49"),
		[]byte("1bccb646905256d20da48c875ddfa7db92a182bd0b7738560d7df6ead909892e"),
		[]byte("f9c165744018c13a098b7e4915d24d86379f8a06fab28bd7970092ab5a19fd41"),
	}

	publicKeys := []string{
		"c1135f9b0fcae055218cc7916f626f3da33e2ccc0252fd8036be35d4e2d93b9b54a6c355b3e6520b49d32ca005a757156e2a8dc1b14e5c7773a294f6ea1faecae6739e5b3d832eab7f36ff9a6c200ca471a948dcf7671291347b79c3f1b63e93",
		"ae33fdd47ca6eed4b4e33a87beb580a20e908898a88c1d91a8b376cc35e8240e5083696ba6a1eeaa4cf50431980c38086dd7acc535c7571fb952d5c025d27c422fca999eaeaa13451946504d2b0a0c5b08958da236a4877b08abbd8059218f05",
		"85c447d1a50ac4dd6c8b38dd39ade1126caf16654d59ea8ef7dbec658b36ccd5cbcd946f9c4acccacdde564739f224002ea0a2ce083abd60cafbcae9d817f674966e49c3b3322a2028c64fa74b01610e25e3f9ceb7c2b2077eaed83ca6e08090",
		"c550ac126ce520d7a6bbd7d5f375273df8f1a8c6d74f44d3e3b71872fc65e436ef52696674883ff27de8f674bc4c7713c4b7bdabe2292c069e81e5c8e131ea8a90215ee4038a882687e532a8c27dea94c5c2ca9f7f6072163bb0b6151c93b00a",
		"1ecad7660e5a77a09661207c7a22f4a84f9be98ac520d4cc875d0caea2fc98f0ab67c2b4966a3ba1cefaa6013517b30f8c43327b46896111886fe2ba10e66d2589aea52e9bf3d72f630c051733fddba412e7c7768b80c8fb7b7e104156db0a0a",
	}

	return privateKeys, publicKeys
}
