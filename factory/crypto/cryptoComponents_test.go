package crypto_test

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go/config"
	errErd "github.com/ElrondNetwork/elrond-go/errors"
	cryptoComp "github.com/ElrondNetwork/elrond-go/factory/crypto"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewCryptoComponentsFactory_NiCoreComponentsHandlerShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetCryptoArgs(nil)
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, errErd.ErrNilCoreComponents, err)
}

func TestNewCryptoComponentsFactory_NilPemFileShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.ValidatorKeyPemFileName = ""
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, errErd.ErrNilPath, err)
}

func TestCryptoComponentsFactory_CreateCryptoParamsNilKeyLoaderShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.KeyLoader = nil
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)

	require.Nil(t, ccf)
	require.Equal(t, errErd.ErrNilKeyLoader, err)
}

func TestNewCryptoComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestNewCryptoComponentsFactory_DisabledSigShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.ImportModeNoSigCheck = true
	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestNewCryptoComponentsFactory_CreateInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestCryptoComponentsFactory_CreateWithDisabledSig(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.IsInImportMode = true
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestCryptoComponentsFactory_CreateWithAutoGenerateKey(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetCryptoArgs(coreComponents)
	args.NoKeyProvided = true
	ccf, _ := cryptoComp.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestCryptoComponentsFactory_CreateSingleSignerInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
