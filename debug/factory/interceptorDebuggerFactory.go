package factory

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug/handler"
)

// NewInterceptorDebuggerFactory will instantiate an InterceptorDebugHandler based on the provided config
func NewInterceptorDebuggerFactory(config config.InterceptorResolverDebugConfig, ntpTime handler.NTPTime) (InterceptorDebugHandler, error) {
	if !config.Enabled {
		return handler.NewDisabledInterceptorDebugHandler(), nil
	}

	return handler.NewInterceptorDebugHandler(config, ntpTime)
}
