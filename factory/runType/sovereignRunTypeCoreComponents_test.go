package runType_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"
)

func TestNewSovereignRunTypeCoreComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil runType components factory", func(t *testing.T) {
		srccf, err := runType.NewSovereignRunTypeCoreComponentsFactory(nil)
		require.Nil(t, srccf)
		require.ErrorIs(t, errors.ErrNilRunTypeCoreComponentsFactory, err)
	})
	t.Run("should work", func(t *testing.T) {
		rccf := runType.NewRunTypeCoreComponentsFactory()
		srccf, err := runType.NewSovereignRunTypeCoreComponentsFactory(rccf)
		require.NotNil(t, srccf)
		require.NoError(t, err)
	})
}

func TestSovereignRunTypeCoreComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	rccf := runType.NewRunTypeCoreComponentsFactory()
	srccf, err := runType.NewSovereignRunTypeCoreComponentsFactory(rccf)
	require.NotNil(t, srccf)
	require.NoError(t, err)

	rcc := srccf.Create()
	require.NotNil(t, rcc)
}

func TestSovereignRunTypeCoreComponentsFactory_Close(t *testing.T) {
	t.Parallel()

	rccf := runType.NewRunTypeCoreComponentsFactory()
	srccf, err := runType.NewSovereignRunTypeCoreComponentsFactory(rccf)
	require.NotNil(t, srccf)
	require.NoError(t, err)

	rcc := srccf.Create()
	require.NotNil(t, rcc)

	require.NoError(t, rcc.Close())
}
