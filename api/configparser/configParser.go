package configparser

import "github.com/ElrondNetwork/elrond-go/config"

// CheckEndpoint will return true if a configuration for the given endpoint exists and it is configured to be opened
func CheckEndpoint(endpointToCheck string, routesConfig config.APIPackageConfig) bool {
	for _, endpoint := range routesConfig.Routes {
		if endpoint.Name == endpointToCheck && endpoint.Open {
			return true
		}
	}

	return false
}
