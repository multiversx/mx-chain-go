package runType_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignRunTypeComponentsFactory(t *testing.T) {
	t.Parallel()

	srcf, err := runType.NewSovereignRunTypeComponentsFactory(nil)
	require.Nil(t, srcf)
	require.ErrorIs(t, errors.ErrNilRunTypeComponentsFactory, err)

	rcf, _ := runType.NewRunTypeComponentsFactory()
	srcf, err = runType.NewSovereignRunTypeComponentsFactory(rcf)
	require.NotNil(t, srcf)
	require.NoError(t, err)
}

func TestSovereignRunTypeComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	rcf, _ := runType.NewRunTypeComponentsFactory()
	srcf, _ := runType.NewSovereignRunTypeComponentsFactory(rcf)

	rc, err := srcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)
}

func TestSovereignRunTypeComponentsFactory_Close(t *testing.T) {
	t.Parallel()

	rcf, _ := runType.NewRunTypeComponentsFactory()
	srcf, _ := runType.NewSovereignRunTypeComponentsFactory(rcf)

	rc, err := srcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)

	require.NoError(t, rc.Close())
}
