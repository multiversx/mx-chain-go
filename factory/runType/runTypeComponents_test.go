package runType_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"

	"github.com/stretchr/testify/require"
)

func TestNewRunTypeComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil core components should error", func(t *testing.T) {
		args := createArgsRunTypeComponents()
		args.CoreComponents = nil
		rcf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, rcf)
		require.Equal(t, errors.ErrNilCoreComponents, err)
	})
	t.Run("nil crypto components should error", func(t *testing.T) {
		args := createArgsRunTypeComponents()
		args.CryptoComponents = nil
		rcf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, rcf)
		require.Equal(t, errors.ErrNilCryptoComponents, err)
	})
	t.Run("nil initial accounts should error", func(t *testing.T) {
		args := createArgsRunTypeComponents()
		args.InitialAccounts = nil
		rcf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, rcf)
		require.Equal(t, errors.ErrNilInitialAccounts, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createArgsRunTypeComponents()
		rcf, err := runType.NewRunTypeComponentsFactory(args)
		require.NotNil(t, rcf)
		require.NoError(t, err)
	})
}

func TestRunTypeComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	rcf, err := runType.NewRunTypeComponentsFactory(createArgsRunTypeComponents())
	require.NoError(t, err)
	require.NotNil(t, rcf)

	rc, err := rcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)
}

func TestRunTypeComponentsFactory_Close(t *testing.T) {
	t.Parallel()

	rcf, err := runType.NewRunTypeComponentsFactory(createArgsRunTypeComponents())
	require.NoError(t, err)
	require.NotNil(t, rcf)

	rc, err := rcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)

	require.NoError(t, rc.Close())
}
