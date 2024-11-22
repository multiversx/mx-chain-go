package runType_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/factory/runType"
)

func TestNewRunTypeCoreComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		rccf := runType.NewRunTypeCoreComponentsFactory()
		require.NotNil(t, rccf)
	})
}

func TestRunTypeCoreComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	rccf := runType.NewRunTypeCoreComponentsFactory()
	require.NotNil(t, rccf)

	rcc := rccf.Create()
	require.NotNil(t, rcc)
}

func TestRunTypeCoreComponentsFactory_Close(t *testing.T) {
	t.Parallel()

	rccf := runType.NewRunTypeCoreComponentsFactory()
	require.NotNil(t, rccf)

	rcc := rccf.Create()
	require.NotNil(t, rcc)

	require.NoError(t, rcc.Close())
}
