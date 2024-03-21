package runType_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/stretchr/testify/require"
)

func TestNewRunTypeComponentsFactory(t *testing.T) {
	t.Parallel()

	rcf, err := runType.NewRunTypeComponentsFactory(nil)
	require.Nil(t, rcf)
	require.Equal(t, errors.ErrNilCoreComponents, err)

	rcf, err = runType.NewRunTypeComponentsFactory(createCoreComponents())
	require.NotNil(t, rcf)
	require.NoError(t, err)
}

func TestRunTypeComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	rcf, err := runType.NewRunTypeComponentsFactory(createCoreComponents())
	require.NoError(t, err)
	require.NotNil(t, rcf)

	rc, err := rcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)
}

func TestRunTypeComponentsFactory_Close(t *testing.T) {
	t.Parallel()

	rcf, err := runType.NewRunTypeComponentsFactory(createCoreComponents())
	require.NoError(t, err)
	require.NotNil(t, rcf)

	rc, err := rcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)

	require.NoError(t, rc.Close())
}
