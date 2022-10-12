package statusCore_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common/statistics"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory/statusCore"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewStatusCoreComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs()
	sccf := statusCore.NewStatusCoreComponentsFactory(args)

	require.NotNil(t, sccf)
}

func TestStatusCoreComponentsFactory_InvalidValueShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs()
	args.Config = config.Config{
		ResourceStats: config.ResourceStatsConfig{
			RefreshIntervalInSec: 0,
		},
	}
	sccf := statusCore.NewStatusCoreComponentsFactory(args)

	cc, err := sccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, statistics.ErrInvalidRefreshIntervalValue))
}

func TestStatusCoreComponentsFactory_CreateStatusCoreComponentsShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs()
	sccf := statusCore.NewStatusCoreComponentsFactory(args)

	cc, err := sccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

// ------------ Test CoreComponents --------------------
func TestStatusCoreComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs()
	sccf := statusCore.NewStatusCoreComponentsFactory(args)
	cc, err := sccf.Create()
	require.NoError(t, err)

	err = cc.Close()
	require.NoError(t, err)
}
