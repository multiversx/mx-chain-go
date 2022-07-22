package crypto_test

import (
	"testing"
)

func Test_NewMultiSignerContainer(t *testing.T) {

}

func TestContainer_GetMultiSigner(t *testing.T) {

}

func TestContainer_IsInterfaceNil(t *testing.T) {

}

func TestContainer_createMultiSigner(t *testing.T) {

}

func TestContainer_createLowLevelSigner(t *testing.T) {

}

func TestContainer_getMultiSigHasherFromConfig(t *testing.T) {

}

func TestContainer_sortMultiSignerConfig(t *testing.T) {

}

//
//func TestCryptoComponentsFactory_GetMultiSigHasherFromConfigInvalidHasherShouldErr(t *testing.T) {
//	t.Parallel()
//	if testing.Short() {
//		t.Skip("this is not a short test")
//	}
//
//	coreComponents := componentsMock.GetCoreComponents()
//	args := componentsMock.GetCryptoArgs(coreComponents)
//	args.Config.Consensus.Type = ""
//	args.Config.MultisigHasher.Type = ""
//	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
//	require.NotNil(t, ccf)
//	require.Nil(t, err)
//
//	multiSigHasher, err := cryptoComp.GetMultiSigHasherFromConfig()
//	require.Nil(t, multiSigHasher)
//	require.Equal(t, errErd.ErrMissingMultiHasherConfig, err)
//}
//
//func TestCryptoComponentsFactory_GetMultiSigHasherFromConfigMismatchConsensusTypeMultiSigHasher(t *testing.T) {
//	t.Parallel()
//	if testing.Short() {
//		t.Skip("this is not a short test")
//	}
//
//	coreComponents := componentsMock.GetCoreComponents()
//	args := componentsMock.GetCryptoArgs(coreComponents)
//	args.Config.MultisigHasher.Type = "sha256"
//	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
//	require.NotNil(t, ccf)
//	require.Nil(t, err)
//
//	multiSigHasher, err := cryptoComp.GetMultiSigHasherFromConfig()
//	require.Nil(t, multiSigHasher)
//	require.Equal(t, errErd.ErrMultiSigHasherMissmatch, err)
//}
//
//func TestCryptoComponentsFactory_GetMultiSigHasherFromConfigOK(t *testing.T) {
//	t.Parallel()
//	if testing.Short() {
//		t.Skip("this is not a short test")
//	}
//
//	coreComponents := componentsMock.GetCoreComponents()
//	args := componentsMock.GetCryptoArgs(coreComponents)
//	args.Config.Consensus.Type = "bls"
//	args.Config.MultisigHasher.Type = "blake2b"
//	ccf, err := cryptoComp.NewCryptoComponentsFactory(args)
//	require.NotNil(t, ccf)
//	require.Nil(t, err)
//
//	multiSigHasher, err := cryptoComp.GetMultiSigHasherFromConfig()
//	require.Nil(t, err)
//	require.NotNil(t, multiSigHasher)
//}
