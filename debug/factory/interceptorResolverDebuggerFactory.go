package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/debug/resolver"
)

// NewInterceptorResolverDebuggerFactory will instantiate an InterceptorDebugHandler based on the provided config
func NewInterceptorResolverDebuggerFactory(config config.InterceptorResolverDebugConfig) (InterceptorDebugHandler, error) {
	if !config.Enabled {
		return resolver.NewDisabledInterceptorDebugHandler(), nil
	}

	return resolver.NewInterceptorResolver(config)
}
