package runType_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/testscommon/headerSigVerifier"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"

	"github.com/stretchr/testify/require"
)

func createSovRunTypeArgs() runType.ArgsSovereignRunTypeComponents {
	rcf, _ := runType.NewRunTypeComponentsFactory(createArgsRunTypeComponents())

	return runType.ArgsSovereignRunTypeComponents{
		RunTypeComponentsFactory: rcf,
		Config: config.SovereignConfig{
			GenesisConfig: config.GenesisConfig{
				NativeESDT: "WEGLD",
			},
		},
		DataCodec:     &sovereign.DataCodecMock{},
		TopicsChecker: &sovereign.TopicsCheckerMock{},
		ExtraVerifier: &headerSigVerifier.ExtraHeaderSigVerifierHandlerMock{},
	}
}

func TestNewSovereignRunTypeComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil runType components factory", func(t *testing.T) {
		sovArgs := createSovRunTypeArgs()
		sovArgs.RunTypeComponentsFactory = nil
		srcf, err := runType.NewSovereignRunTypeComponentsFactory(sovArgs)
		require.Nil(t, srcf)
		require.ErrorIs(t, errors.ErrNilRunTypeComponentsFactory, err)
	})
	t.Run("nil data codec", func(t *testing.T) {
		sovArgs := createSovRunTypeArgs()
		sovArgs.DataCodec = nil
		srcf, err := runType.NewSovereignRunTypeComponentsFactory(sovArgs)
		require.Nil(t, srcf)
		require.ErrorIs(t, errors.ErrNilDataCodec, err)
	})
	t.Run("nil topics checker", func(t *testing.T) {
		sovArgs := createSovRunTypeArgs()
		sovArgs.TopicsChecker = nil
		srcf, err := runType.NewSovereignRunTypeComponentsFactory(sovArgs)
		require.Nil(t, srcf)
		require.ErrorIs(t, errors.ErrNilTopicsChecker, err)
	})
	t.Run("should work", func(t *testing.T) {
		srcf, err := runType.NewSovereignRunTypeComponentsFactory(createSovRunTypeArgs())
		require.NotNil(t, srcf)
		require.NoError(t, err)
	})
}

func TestSovereignRunTypeComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	srcf, _ := runType.NewSovereignRunTypeComponentsFactory(createSovRunTypeArgs())

	rc, err := srcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)
}

func TestSovereignRunTypeComponentsFactory_Close(t *testing.T) {
	t.Parallel()

	srcf, _ := runType.NewSovereignRunTypeComponentsFactory(createSovRunTypeArgs())

	rc, err := srcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)

	require.NoError(t, rc.Close())
}
