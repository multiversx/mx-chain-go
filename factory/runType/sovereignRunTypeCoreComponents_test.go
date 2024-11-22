package runType_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"
)

func TestSovereignRunTypeCoreComponentsFactory_CreateAndClose(t *testing.T) {
	t.Parallel()

	t.Run("nil run type core components factory", func(t *testing.T) {
		srccf, err := runType.NewSovereignRunTypeCoreComponentsFactory(nil)
		require.Nil(t, srccf)
		require.ErrorIs(t, errorsMx.ErrNilRunTypeCoreComponentsFactory, err)
	})
	t.Run("should work", func(t *testing.T) {
		rccf := runType.NewRunTypeCoreComponentsFactory(config.EpochConfig{})
		srccf, err := runType.NewSovereignRunTypeCoreComponentsFactory(rccf)
		require.NoError(t, err)
		require.NotNil(t, srccf)

		rcc, err := srccf.Create()
		require.NoError(t, err)
		require.NotNil(t, rcc)

		require.NoError(t, rcc.Close())
	})
}
