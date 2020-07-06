package factory_test

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

const dummyPk = "629e1245577afb7717ccb46b6ff3649bdd6a1311514ad4a7695da13f801cc277ee24e730a7fa8aa6c612159b4328db17" +
	"35692d0bded3a2264ba621d6bda47a981d60e17dd306d608e0875a0ba19639fb0844661f519472a175ca9ed2f33fbe16"
const dummySk = "cea01c0bf060187d90394802ff223078e47527dc8aa33a922744fb1d06029c4b"

type LoadKeysFunc func(string, int) ([]byte, string, error)

func TestNewCryptoComponentsFactory_NiCoreComponentsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs(nil)
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilCoreComponents, err)
}

func TestNewCryptoComponentsFactory_NilPemFileShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.ValidatorKeyPemFileName = ""
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilPath, err)
}

func TestNewCryptoComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestNewCryptoComponentsFactory_CreateInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "invalid"
	ccf, _ := factory.NewCryptoComponentsFactory(args)
	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.Equal(t, err, factory.ErrInvalidConsensusConfig)
}

func TestCryptoComponentsFactory_CreateShouldErrDueToMissingConfig(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config = config.Config{
		ValidatorPubkeyConverter: config.PubkeyConfig{
			Length:          8,
			Type:            "hex",
			SignatureLength: 48,
		}}

	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, err)
	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	cc, err := ccf.Create()
	require.Error(t, err)
	require.Nil(t, cc)
}

func TestCryptoComponentsFactory_CreateInvalidMultiSigHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.MultisigHasher.Type = "invalid"
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, err)

	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)
	cspf, err := ccf.Create()
	require.Nil(t, cspf)
	require.Equal(t, factory.ErrMultiSigHasherMissmatch, err)
}

func TestCryptoComponentsFactory_CreateOK(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, _ := factory.NewCryptoComponentsFactory(args)
	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestCryptoComponentsFactory_CreateSingleSignerInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "invalid"
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)
	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	singleSigner, err := ccf.CreateSingleSigner()
	require.Nil(t, singleSigner)
	require.Equal(t, factory.ErrInvalidConsensusConfig, err)
}

func TestCryptoComponentsFactory_CreateSingleSignerOK(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)
	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	singleSigner, err := ccf.CreateSingleSigner()
	require.Nil(t, err)
	require.NotNil(t, singleSigner)
}

func TestCryptoComponentsFactory_GetMultiSigHasherFromConfigInvalidHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.Consensus.Type = ""
	args.Config.MultisigHasher.Type = ""
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	multiSigHasher, err := ccf.GetMultiSigHasherFromConfig()
	require.Nil(t, multiSigHasher)
	require.Equal(t, factory.ErrMissingMultiHasherConfig, err)
}

func TestCryptoComponentsFactory_GetMultiSigHasherFromConfigMissmatchConsensusTypeMultiSigHasher(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.MultisigHasher.Type = "sha256"
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	multiSigHasher, err := ccf.GetMultiSigHasherFromConfig()
	require.Nil(t, multiSigHasher)
	require.Equal(t, factory.ErrMultiSigHasherMissmatch, err)
}

func TestCryptoComponentsFactory_GetMultiSigHasherFromConfigOK(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "bls"
	args.Config.MultisigHasher.Type = "blake2b"
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	multiSigHasher, err := ccf.GetMultiSigHasherFromConfig()
	require.Nil(t, err)
	require.NotNil(t, multiSigHasher)
}

func TestCryptoComponentsFactory_CreateMultiSignerInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "other"
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	multiSigner, err := ccf.CreateMultiSigner(mock.HasherMock{}, &factory.CryptoParams{}, &mock.KeyGenMock{})
	require.Nil(t, multiSigner)
	require.Equal(t, factory.ErrInvalidConsensusConfig, err)
}

func TestCryptoComponentsFactory_CreateMultiSignerOK(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	suite, _ := ccf.GetSuite()
	blockSignKeyGen := signing.NewKeyGenerator(suite)
	cp, _ := ccf.CreateCryptoParams(blockSignKeyGen)
	multisigHasher, _ := ccf.GetMultiSigHasherFromConfig()

	multiSigner, err := ccf.CreateMultiSigner(multisigHasher, cp, blockSignKeyGen)
	require.Nil(t, err)
	require.NotNil(t, multiSigner)
}

func TestCryptoComponentsFactory_GetSuiteInvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.Consensus.Type = ""
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	suite, err := ccf.GetSuite()
	require.Nil(t, suite)
	require.Equal(t, factory.ErrInvalidConsensusConfig, err)
}

func TestCryptoComponentsFactory_GetSuiteOK(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	args.Config.Consensus.Type = "bls"
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NotNil(t, ccf)
	require.Nil(t, err)

	suite, err := ccf.GetSuite()
	require.Nil(t, err)
	require.NotNil(t, suite)
}

func TestCryptoComponentsFactory_CreateCryptoParamsInvalidPrivateKeyByteArrayShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, _ := factory.NewCryptoComponentsFactory(args)
	keyLoader := dummyLoadSkPkFromPemFile([]byte{}, dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	suite, _ := ccf.GetSuite()
	blockSignKeyGen := signing.NewKeyGenerator(suite)

	cryptoParams, err := ccf.CreateCryptoParams(blockSignKeyGen)
	require.Nil(t, cryptoParams)
	require.Equal(t, crypto.ErrInvalidParam, err)
}

func TestCryptoComponentsFactory_CreateCryptoParamsLoadKeysFailShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, _ := factory.NewCryptoComponentsFactory(args)

	expectedError := errors.New("expected error")
	keyLoader := dummyLoadSkPkFromPemFile([]byte{}, "", expectedError)
	ccf.SetKeyLoader(keyLoader)

	suite, _ := ccf.GetSuite()
	blockSignKeyGen := signing.NewKeyGenerator(suite)

	cryptoParams, err := ccf.CreateCryptoParams(blockSignKeyGen)
	require.Nil(t, cryptoParams)
	require.Equal(t, expectedError, err)
}

func TestCryptoComponentsFactory_CreateCryptoParamsOK(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, _ := factory.NewCryptoComponentsFactory(args)

	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	suite, _ := ccf.GetSuite()
	blockSignKeyGen := signing.NewKeyGenerator(suite)

	cryptoParams, err := ccf.CreateCryptoParams(blockSignKeyGen)
	require.Nil(t, err)
	require.NotNil(t, cryptoParams)
}

func TestCryptoComponentsFactory_GetSkPkInvalidSkBytesShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, _ := factory.NewCryptoComponentsFactory(args)
	setSk := []byte("zxwY")
	setPk := []byte(dummyPk)

	keyLoader := dummyLoadSkPkFromPemFile(setSk, string(setPk), nil)
	ccf.SetKeyLoader(keyLoader)

	sk, pk, err := ccf.GetSkPk()
	require.NotNil(t, err)
	require.Nil(t, sk)
	require.Nil(t, pk)
}

func TestCryptoComponentsFactory_GetSkPkInvalidPkBytesShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, _ := factory.NewCryptoComponentsFactory(args)
	setSk := []byte(dummySk)
	setPk := "0"
	keyLoader := dummyLoadSkPkFromPemFile(setSk, setPk, nil)
	ccf.SetKeyLoader(keyLoader)

	sk, pk, err := ccf.GetSkPk()
	require.NotNil(t, err)
	require.Nil(t, sk)
	require.Nil(t, pk)
}

func TestCryptoComponentsFactory_GetSkPkOK(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getCryptoArgs(coreComponents)
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, err)

	expectedSk, _ := hex.DecodeString(dummySk)
	expectedPk, _ := hex.DecodeString(dummyPk)
	keyLoader := dummyLoadSkPkFromPemFile([]byte(dummySk), dummyPk, nil)
	ccf.SetKeyLoader(keyLoader)

	sk, pk, err := ccf.GetSkPk()
	require.Nil(t, err)
	require.Equal(t, expectedSk, sk)
	require.Equal(t, expectedPk, pk)
}

func getCryptoArgs(coreComponents factory.CoreComponentsHolder) factory.CryptoComponentsFactoryArgs {
	args := factory.CryptoComponentsFactoryArgs{
		Config: config.Config{
			GeneralSettings: config.GeneralSettingsConfig{ChainID: "undefined"},
			Consensus:       config.ConsensusConfig{Type: "bls"},
			MultisigHasher:  config.TypeConfig{Type: "blake2b"},
			Hasher:          config.TypeConfig{Type: "blake2b"},
		},
		SkIndex:                              0,
		ValidatorKeyPemFileName:              "validatorKey.pem",
		CoreComponentsHolder:                 coreComponents,
		ActivateBLSPubKeyMessageVerification: false,
	}

	return args
}

func dummyLoadSkPkFromPemFile(sk []byte, pk string, err error) LoadKeysFunc {
	return func(_ string, _ int) ([]byte, string, error) {
		return sk, pk, err
	}
}
