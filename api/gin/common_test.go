package gin

import (
	"errors"
	"testing"

	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/facade/initial"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestCommon_checkArgs(t *testing.T) {
	t.Parallel()

	args := ArgsNewWebServer{
		Facade:          nil,
		ApiConfig:       config.ApiRoutesConfig{},
		AntiFloodConfig: config.WebServerAntifloodConfig{},
	}
	err := checkArgs(args)
	require.True(t, errors.Is(err, apiErrors.ErrCannotCreateGinWebServer))

	args.Facade, err = initial.NewInitialNodeFacade("api interface", false, &testscommon.StatusMetricsStub{})
	require.NoError(t, err)
	err = checkArgs(args)
	require.NoError(t, err)
}

func TestCommon_isLogRouteEnabled(t *testing.T) {
	t.Parallel()

	routesConfigWithMissingLog := config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{},
	}
	require.False(t, isLogRouteEnabled(routesConfigWithMissingLog))

	routesConfig := config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"log": {
				Routes: []config.RouteConfig{
					{Name: "/log", Open: true},
				},
			},
		},
	}
	require.True(t, isLogRouteEnabled(routesConfig))
	require.False(t, isLogRouteEnabled(config.ApiRoutesConfig{}))
}
