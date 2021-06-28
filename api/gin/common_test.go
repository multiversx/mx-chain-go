package gin

import (
	"errors"
	"testing"

	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/facade/disabled"
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

	args.Facade = disabled.NewDisabledNodeFacade("api interface")
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
}
