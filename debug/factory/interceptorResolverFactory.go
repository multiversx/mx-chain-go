package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/debug/resolver"
)

// NewInterceptorResolverFactory will instantiate an InterceptorResolverDebugHandler based on the provided config
func NewInterceptorResolverFactory(config config.InterceptorResolverDebugConfig) (InterceptorResolverDebugHandler, error) {
	if !config.Enabled {
		return resolver.NewDisabledInterceptorResolver()
	}

	return resolver.NewInterceptorResolver(config)
}
